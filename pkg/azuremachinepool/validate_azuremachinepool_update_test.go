package azuremachinepool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachinePoolUpdateValidate(t *testing.T) {
	tr := true
	fa := false
	supportedInstanceType := ""
	type testCase struct {
		name         string
		vmSize       string
		oldNodePool  []byte
		newNodePool  []byte
		allowed      bool
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: Enabled and untouched",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			newNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name:         "case 1: Disabled and untouched",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			newNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name:         "case 2: Changed from true to false",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			newNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 3: Changed from false to true",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			newNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 4: changed from nil to true",
			oldNodePool:  azureMPRawObject(supportedInstanceType, nil),
			newNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 5: changed from true to nil",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &tr),
			newNodePool:  azureMPRawObject(supportedInstanceType, nil),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 6: changed from nil to false",
			oldNodePool:  azureMPRawObject(supportedInstanceType, nil),
			newNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 7: changed from false to nil",
			oldNodePool:  azureMPRawObject(supportedInstanceType, &fa),
			newNodePool:  azureMPRawObject(supportedInstanceType, nil),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
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
			fakeK8sClient := unittest.FakeK8sClient()
			admit := &UpdateValidator{
				k8sClient: fakeK8sClient,
				logger:    newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			allowed, err := admit.Validate(context.Background(), getUpdateAdmissionRequest(tc.oldNodePool, tc.newNodePool))

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
			if tc.allowed != allowed {
				t.Fatalf("expected %v to be equal to %v", tc.allowed, allowed)
			}
		})
	}
}

func getUpdateAdmissionRequest(oldMP []byte, newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    oldMP,
			Object: nil,
		},
	}

	return req
}

func azureMPRawObject(vmSize string, acceleratedNetworkingEnabled *bool) []byte {
	mp := v1alpha3.AzureMachinePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureMachinePool",
			APIVersion: "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"giantswarm.io/cluster":                "ab123",
				"giantswarm.io/machine-pool":           "ab123",
				"giantswarm.io/organization":           "giantswarm",
				"release.giantswarm.io/version":        "13.0.0",
			},
		},
		Spec: v1alpha3.AzureMachinePoolSpec{
			Location: "westeurope",
			Template: v1alpha3.AzureMachineTemplate{
				VMSize:                vmSize,
				AcceleratedNetworking: acceleratedNetworkingEnabled,
			},
		},
	}
	byt, _ := json.Marshal(mp)
	return byt
}
