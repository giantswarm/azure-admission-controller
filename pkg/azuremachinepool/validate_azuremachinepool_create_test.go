package azuremachinepool

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
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
		nodePool     []byte
		errorMatcher func(err error) bool
	}

	var testCases []testCase

	for i, instanceType := range unsupportedInstanceType {
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", i*3, instanceType),
			nodePool:     azureMPRawObject(instanceType, &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks),
			errorMatcher: IsInvalidOperationError,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking disabled", i*3+1, instanceType),
			nodePool:     azureMPRawObject(instanceType, &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks),
			errorMatcher: nil,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking nil", i*3+2, instanceType),
			nodePool:     azureMPRawObject(instanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks),
			errorMatcher: nil,
		})
	}

	// Non existing instance type.
	{
		instanceType := "this_is_a_random_name"
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", len(testCases)-1, instanceType),
			nodePool:     azureMPRawObject(instanceType, &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks),
			errorMatcher: vmcapabilities.IsSkuNotFoundError,
		})
	}

	{
		testCases = append(testCases, testCase{
			name: fmt.Sprintf("case %d: data disks already set", len(testCases)),
			nodePool: azureMPRawObject("Standard_A2_v2", &tr, string(compute.StorageAccountTypesStandardLRS), []capzv1alpha3.DataDisk{
				{
					NameSuffix: "docker",
					DiskSizeGB: 50,
					Lun:        to.Int32Ptr(21),
				},
				{
					NameSuffix: "kubelet",
					DiskSizeGB: 50,
					Lun:        to.Int32Ptr(22),
				},
			}),
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

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// Create default GiantSwarm organization.
			organization := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
				Spec: securityv1alpha1.OrganizationSpec{},
			}
			err = ctrlClient.Create(ctx, organization)
			if err != nil {
				t.Fatal(err)
			}

			stubbedSKUs := map[string]compute.ResourceSku{
				"Standard_A2_v2": {
					Name: to.StringPtr("Standard_A2_v2"),
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
					},
				},
				"Standard_A4_v2": {
					Name: to.StringPtr("Standard_A4_v2"),
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
					},
				},
				"Standard_A8_v2": {
					Name: to.StringPtr("Standard_A8_v2"),
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
					},
				},
				"Standard_D2_v3": {
					Name: to.StringPtr("Standard_D2_v3"),
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
					},
				},
				"Standard_D2s_v3": {
					Name: to.StringPtr("Standard_D2s_v3"),
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

			admit := &CreateValidator{
				ctrlClient: ctrlClient,
				logger:     newLogger,
				vmcaps:     vmcaps,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(ctx, getCreateAdmissionRequest(tc.nodePool))

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
		})
	}
}

type StubAPI struct {
	stubbedSKUs map[string]compute.ResourceSku
}

func NewStubAPI(stubbedSKUs map[string]compute.ResourceSku) vmcapabilities.API {
	return &StubAPI{stubbedSKUs: stubbedSKUs}
}

func (s *StubAPI) List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error) {
	return s.stubbedSKUs, nil
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
