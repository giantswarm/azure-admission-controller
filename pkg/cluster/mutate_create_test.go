package cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestClusterCreateMutate(t *testing.T) {
	type testCase struct {
		name         string
		cluster      *capi.Cluster
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	clusterNetwork := &capi.ClusterNetwork{
		APIServerPort: to.Int32Ptr(443),
		ServiceDomain: "cluster.local",
		Services: &capi.NetworkRanges{
			CIDRBlocks: []string{
				"172.31.0.0/16",
			},
		},
	}

	operatorsLabels := map[string]string{
		label.AzureOperatorVersion:   "5.0.0",
		label.ClusterOperatorVersion: "1.2.3",
	}

	testCases := []testCase{
		{
			name:    "case 0: ControlPlaneEndpoint left empty",
			cluster: clusterObject("ab123", clusterNetwork, "", 0, operatorsLabels),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/controlPlaneEndpoint/host",
					Value:     "api.ab123.k8s.test.westeurope.azure.gigantic.io",
				},
				{
					Operation: "add",
					Path:      "/spec/controlPlaneEndpoint/port",
					Value:     443,
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 1: ControlPlaneEndpoint has a value",
			cluster:      clusterObject("ab123", clusterNetwork, "api.giantswarm.io", 123, operatorsLabels),
			patches:      []mutator.PatchOperation{},
			errorMatcher: nil,
		},
		{
			name:    "case 2: Operator version labels missing",
			cluster: clusterObject("ab123", clusterNetwork, "api.giantswarm.io", 123, nil),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
					Value:     "5.0.0",
				},
				{
					Operation: "add",
					Path:      "/metadata/labels/cluster-operator.giantswarm.io~1version",
					Value:     "1.2.3",
				},
			},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			release13 := &v1alpha1.Release{
				ObjectMeta: metav1.ObjectMeta{
					Name: "v13.0.0-alpha4",
				},
				Spec: v1alpha1.ReleaseSpec{
					Components: []v1alpha1.ReleaseSpecComponent{
						{
							Name:    "azure-operator",
							Version: "5.0.0",
						}, {
							Name:    "cluster-operator",
							Version: "1.2.3",
						},
					},
				},
			}
			err = ctrlClient.Create(ctx, release13)
			if err != nil {
				t.Fatal(err)
			}

			// Cluster with both operator annotations.
			ab123 := &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"azure-operator.giantswarm.io/version":   "5.0.0",
						"cluster-operator.giantswarm.io/version": "1.2.3",
					},
				},
			}
			err = ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				BaseDomain: "k8s.test.westeurope.azure.gigantic.io",
				CtrlClient: ctrlClient,
				CtrlReader: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run mutating webhook handler on Cluster creation.
			patches, err := handler.OnCreateMutate(context.Background(), tc.cluster)

			// Check if the error is the expected one.
			switch {
			case err == nil && tc.errorMatcher == nil:
				// fall through
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("expected %#v got %#v", nil, err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("expected %#v got %#v", "error", nil)
			case !tc.errorMatcher(err):
				t.Fatalf("unexpected error: %#v", err)
			}

			// Check if the validation result is the expected one.
			if len(tc.patches) != 0 || len(patches) != 0 {
				if !reflect.DeepEqual(tc.patches, patches) {
					t.Fatalf("Patches mismatch: expected %v, got %v", tc.patches, patches)
				}
			}
		})
	}
}

func clusterObject(clusterName string, clusterNetwork *capi.ClusterNetwork, controlPlaneEndpointHost string, controlPlaneEndpointPort int32, labels map[string]string) *capi.Cluster {
	mergedLabels := map[string]string{
		"cluster.x-k8s.io/cluster-name": clusterName,
		"giantswarm.io/cluster":         clusterName,
		"giantswarm.io/organization":    "giantswarm",
		"release.giantswarm.io/version": "13.0.0-alpha4",
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}

	cluster := capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "default",
			Labels:    mergedLabels,
		},
		Spec: capi.ClusterSpec{
			ClusterNetwork: clusterNetwork,
			ControlPlaneEndpoint: capi.APIEndpoint{
				Host: controlPlaneEndpointHost,
				Port: controlPlaneEndpointPort,
			},
		},
	}
	return &cluster
}
