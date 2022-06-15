package machinepool

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/machinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

const (
	machinePoolNamespace = "org-giantswarm"
	machinePoolName      = "ab123"
)

func TestMachinePoolCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		machinePool  *capiexp.MachinePool
		vmType       string
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		//{
		//	name:         "case 0: instance type supporting [1,2,3], requested [1]",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.FailureDomains([]string{"1"})),
		//	vmType:       "Standard_A2_v2",
		//	errorMatcher: nil,
		//},
		//{
		//	name:         "case 1: instance type supporting [1,2], requested [3]",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.FailureDomains([]string{"3"})),
		//	vmType:       "Standard_A4_v2",
		//	errorMatcher: IsUnsupportedFailureDomainError,
		//},
		//{
		//	name:         "case 2: instance type supporting [1,2], requested [2,3]",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.FailureDomains([]string{"2,3"})),
		//	vmType:       "Standard_A4_v2",
		//	errorMatcher: IsUnsupportedFailureDomainError,
		//},
		//{
		//	name:         "case 3: instance type supporting [], requested [1]",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.FailureDomains([]string{"1"})),
		//	vmType:       "Standard_A8_v2",
		//	errorMatcher: IsUnsupportedFailureDomainError,
		//},
		//{
		//	name:         "case 4: instance type supporting [], requested []",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.FailureDomains([]string{})),
		//	vmType:       "Standard_A8_v2",
		//	errorMatcher: nil,
		//},
		//{
		//	name:         "case 5: AzureMachinePool does not exist",
		//	machinePool:  builder.BuildMachinePool(builder.AzureMachinePool("wrong"), builder.FailureDomains([]string{})),
		//	vmType:       "",
		//	errorMatcher: IsAzureMachinePoolNotFound,
		//},
		{
			name:         "case 6: Wrong Organization",
			machinePool:  builder.BuildMachinePool(builder.AzureMachinePool(machinePoolName), builder.Organization("wrongorg")),
			vmType:       "",
			errorMatcher: generic.IsNodepoolOrgDoesNotMatchClusterOrg,
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
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{"1", "2", "3"},
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
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{"1", "2"},
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
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{},
						},
					},
				},
			}
			vmcaps := unittest.NewVMCapsStubFactory(stubbedSKUs, newLogger)

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// Create AzureMachinePool.
			if tc.vmType != "" {
				amp := &capzexp.AzureMachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machinePoolName,
						Namespace: machinePoolNamespace,
					},
					Spec: capzexp.AzureMachinePoolSpec{
						Location: "westeurope",
						Template: capzexp.AzureMachinePoolMachineTemplate{
							VMSize: tc.vmType,
						},
					},
				}
				err = ctrlClient.Create(ctx, amp)
				if err != nil {
					t.Fatal(err)
				}
			}

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

			// Create cluster CR.
			cluster := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ab123",
					Labels: map[string]string{
						label.Cluster:      "ab123",
						label.Organization: "giantswarm",
					},
				},
			}
			err = ctrlClient.Create(ctx, cluster)
			if err != nil {
				t.Fatal(err)
			}

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				CtrlClient:    ctrlClient,
				Decoder:       unittest.NewFakeDecoder(),
				Logger:        newLogger,
				VMcapsFactory: vmcaps,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on MachinePool creation.
			err = handler.OnCreateValidate(ctx, tc.machinePool)

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
