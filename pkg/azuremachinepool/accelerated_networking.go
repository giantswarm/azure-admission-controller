package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func checkAcceleratedNetworking(ctx context.Context, vmcaps *vmcapabilities.VMSKU, mp expcapzv1alpha3.AzureMachinePool) (bool, error) {
	// If the instance type is invalid, the following function returns an error.
	acceleratedNetworkingAvailable, err := vmcaps.HasCapability(ctx, mp.Spec.Location, mp.Spec.Template.VMSize, vmcapabilities.CapabilityAcceleratedNetworking)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// Accelerated networking is disabled (false) or in auto-detect mode (nil). This is always allowed.
	if mp.Spec.Template.AcceleratedNetworking == nil || !*mp.Spec.Template.AcceleratedNetworking {
		return true, nil
	}

	// Accelerated networking is enabled (true).
	return acceleratedNetworkingAvailable, nil
}

func isAcceleratedNetworkingUnchanged(ctx context.Context, old expcapzv1alpha3.AzureMachinePool, new expcapzv1alpha3.AzureMachinePool) bool {
	if old.Spec.Template.AcceleratedNetworking == nil && new.Spec.Template.AcceleratedNetworking != nil ||
		old.Spec.Template.AcceleratedNetworking != nil && new.Spec.Template.AcceleratedNetworking == nil {
		return false
	}

	if *old.Spec.Template.AcceleratedNetworking != *new.Spec.Template.AcceleratedNetworking {
		return false
	}

	return true
}

func isNewVmSizeSupportingAcceleratedNetworking(ctx context.Context, vmcaps *vmcapabilities.VMSKU, old expcapzv1alpha3.AzureMachinePool, new expcapzv1alpha3.AzureMachinePool) (bool, error) {
	if old.Spec.Template.VMSize == new.Spec.Template.VMSize {
		// VM size unchanged, all good.
		return true, nil
	}
	if old.Spec.Template.AcceleratedNetworking != nil && *old.Spec.Template.AcceleratedNetworking {
		// Accelerated networking was explicitly enabled. New instance type needs support for accelerated networking.
		supported, err := vmcaps.HasCapability(ctx, new.Spec.Location, new.Spec.Template.VMSize, vmcapabilities.CapabilityAcceleratedNetworking)
		if err != nil {
			return false, microerror.Mask(err)
		}
		if !supported {
			// Edited the node pool to use an instance type that does not support accelerated networking.
			return false, nil
		}

		// Edited the node pool to use an instance type that supports accelerated networking.
		return true, nil
	}

	// Accelerated Networking was either nil or false. Any VM size changed is allowed.
	return true, nil
}
