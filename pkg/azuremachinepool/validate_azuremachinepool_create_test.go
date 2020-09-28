package azuremachinepool

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachinePoolCreateValidate(t *testing.T) {
	tr := true
	fa := false
	unsupportedInstanceType := []string{
		"Standard_A2_v2",
		"Standard_A4_v2",
		"Standard_A8_v2",
		"Standard_D2_v3",
		"Standard_D2s_v3",
	}
	type testCase struct {
		name         string
		vmSize       string
		nodePool     []byte
		allowed      bool
		errorMatcher func(err error) bool
	}

	var testCases []testCase

	for i, instanceType := range unsupportedInstanceType {
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", i*3, instanceType),
			nodePool:     azureMPRawObject(instanceType, &tr),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking disabled", i*3+1, instanceType),
			nodePool:     azureMPRawObject(instanceType, &fa),
			allowed:      true,
			errorMatcher: nil,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking nil", i*3+2, instanceType),
			nodePool:     azureMPRawObject(instanceType, nil),
			allowed:      true,
			errorMatcher: nil,
		})
	}

	// Non existing instance type.
	{
		instanceType := "this_is_a_random_name"
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", len(testCases), instanceType),
			nodePool:     azureMPRawObject(instanceType, &tr),
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		})
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
			admit := &CreateValidator{
				k8sClient: fakeK8sClient,
				logger:    newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			allowed, err := admit.Validate(context.Background(), getCreateAdmissionRequest(tc.nodePool))

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

func getCreateAdmissionRequest(newMP []byte) *v1beta1.AdmissionRequest {
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
