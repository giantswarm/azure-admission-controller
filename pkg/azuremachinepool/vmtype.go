package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

const (
	minMemory = 16
	minCPUs   = 4
)

func checkInstanceTypeIsValid(ctx context.Context, vmcaps *vmcapabilities.VMSKU, mp expcapzv1alpha3.AzureMachinePool) (bool, error) {
	memory, err := vmcaps.Memory(ctx, mp.Spec.Location, mp.Spec.Template.VMSize)
	if err != nil {
		return false, microerror.Mask(err)
	}

	cpu, err := vmcaps.CPUs(ctx, mp.Spec.Location, mp.Spec.Template.VMSize)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if memory < minMemory {
		return false, nil
	}

	if cpu < minCPUs {
		return false, nil
	}

	return true, nil
}
