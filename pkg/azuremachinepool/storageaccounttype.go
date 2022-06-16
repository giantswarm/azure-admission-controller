package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func checkStorageAccountTypeIsValid(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMachinePool *capzexp.AzureMachinePool) error {
	selectedStorageAccount := azureMachinePool.Spec.Template.OSDisk.ManagedDisk.StorageAccountType

	if selectedStorageAccount != string(compute.StorageAccountTypesStandardLRS) &&
		selectedStorageAccount != string(compute.StorageAccountTypesPremiumLRS) {
		// Storage account type is invalid.
		return microerror.Maskf(invalidStorageAccountTypeError, "Storage account type %q is invalid. Allowed values are %q and %q", selectedStorageAccount, string(compute.StorageAccountTypesStandardLRS), string(compute.StorageAccountTypesPremiumLRS))
	}

	// Storage account type is valid, check if it matches the VM type's support.
	if selectedStorageAccount == string(compute.StorageAccountTypesPremiumLRS) {
		// Premium is selected, VM type has to support it.
		supported, err := vmcaps.HasCapability(ctx, azureMachinePool.Spec.Location, azureMachinePool.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return microerror.Mask(err)
		}

		if !supported {
			return microerror.Maskf(premiumStorageNotSupportedByVMSizeError, "VM Type %s does not support Premium Storage", azureMachinePool.Spec.Template.VMSize)
		}
	}

	return nil
}
