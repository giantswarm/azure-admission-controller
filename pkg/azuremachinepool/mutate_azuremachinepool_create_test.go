package azuremachinepool

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachinePoolCreateMutate(t *testing.T) {
	tr := true
	type testCase struct {
		name         string
		nodePool     []byte
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:     fmt.Sprintf("case 0: unset storage account type with premium VM"),
			nodePool: azureMPRawObject("Standard_D4s_v3", &tr, "", desiredDataDisks, "westeurope", nil),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/template/osDisk/managedDisk/storageAccountType",
					Value:     "Premium_LRS",
				},
			},
			errorMatcher: nil,
		},
		{
			name:     fmt.Sprintf("case 1: unset storage account type with standard VM"),
			nodePool: azureMPRawObject("Standard_D4_v3", &tr, "", desiredDataDisks, "westeurope", nil),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/template/osDisk/managedDisk/storageAccountType",
					Value:     "Standard_LRS",
				},
			},
			errorMatcher: nil,
		},
		{
			name:     fmt.Sprintf("case 2: set data disks"),
			nodePool: azureMPRawObject("Standard_D4_v3", &tr, "Standard_LRS", nil, "westeurope", nil),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/template/dataDisks",
					Value:     desiredDataDisks,
				},
			},
			errorMatcher: nil,
		},
		{
			name:     fmt.Sprintf("case 3: set location"),
			nodePool: azureMPRawObject("Standard_D4_v3", &tr, "Standard_LRS", desiredDataDisks, "", nil),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/location",
					Value:     "westeurope",
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
			stubbedSKUs := map[string]compute.ResourceSku{
				"Standard_D4_v3": {
					Name: to.StringPtr("Standard_D4_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("False"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("False"),
						},
					},
				},
				"Standard_D4s_v3": {
					Name: to.StringPtr("Standard_D4s_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("False"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("True"),
						},
					},
				},
			}
			stubAPI := NewStubAPI(stubbedSKUs)
			vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
				Azure:  stubAPI,
				Logger: newLogger,
			})
			if err != nil {
				panic(microerror.JSON(err))
			}

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			release13 := &v1alpha1.Release{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "v13.0.0-alpha4",
					Namespace: "default",
				},
				Spec: v1alpha1.ReleaseSpec{
					Components: []v1alpha1.ReleaseSpecComponent{
						{
							Name:    "azure-operator",
							Version: "5.0.0",
						},
					},
				},
			}
			err = ctrlClient.Create(ctx, release13)
			if err != nil {
				t.Fatal(err)
			}

			// Cluster with both operator annotations.
			ab123 := &v1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"azure-operator.giantswarm.io/version": "5.0.0",
					},
				},
			}
			err = ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			admit := &CreateMutator{
				ctrlClient: ctrlClient,
				location:   "westeurope",
				logger:     newLogger,
				vmcaps:     vmcaps,
			}

			// Run admission request to validate AzureConfig updates.
			patches, err := admit.Mutate(context.Background(), getCreateMutateAdmissionRequest(tc.nodePool))

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
			if !reflect.DeepEqual(tc.patches, patches) {
				t.Fatalf("Patches mismatch: expected %v, got %v", tc.patches, patches)
			}
		})
	}
}

func getCreateMutateAdmissionRequest(newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}
