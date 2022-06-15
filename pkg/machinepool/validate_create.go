package machinepool

import (
	"context"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/capzexp/v1alpha3"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
	machinePoolNewCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = machinePoolNewCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelMatchesCluster(ctx, h.ctrlClient, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.checkAvailabilityZones(ctx, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *WebhookHandler) checkAvailabilityZones(ctx context.Context, mp *capiexp.MachinePool) error {
	// Get the AzureMachinePool CR related to this MachinePool (we need it to get the VM type).
	if mp.Spec.Template.Spec.InfrastructureRef.Namespace == "" || mp.Spec.Template.Spec.InfrastructureRef.Name == "" {
		return microerror.Maskf(azureMachinePoolNotFoundError, "MachinePool's InfrastructureRef has to be set")
	}

	var location string
	var vmsize string
	// Try with the non-exp AMP
	{
		amp := capzexp.AzureMachinePool{}
		err := h.ctrlClient.Get(ctx, client.ObjectKey{Namespace: mp.Spec.Template.Spec.InfrastructureRef.Namespace, Name: mp.Spec.Template.Spec.InfrastructureRef.Name}, &amp)
		if errors.IsNotFound(err) {
			// Did not find, we fallback to the exp AMP.
		} else if err != nil {
			return microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool (looking for %q in ns %q)", mp.Spec.Template.Spec.InfrastructureRef.Name, mp.Spec.Template.Spec.InfrastructureRef.Namespace)
		} else {
			location = amp.Spec.Location
			vmsize = amp.Spec.Template.VMSize
		}
	}

	// Fallback to exp AMP
	if location == "" || vmsize == "" {
		amp := v1alpha3.AzureMachinePool{}
		err := h.ctrlClient.Get(ctx, client.ObjectKey{Namespace: mp.Spec.Template.Spec.InfrastructureRef.Namespace, Name: mp.Spec.Template.Spec.InfrastructureRef.Name}, &amp)
		if err != nil {
			return microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool (looking for %q in ns %q)", mp.Spec.Template.Spec.InfrastructureRef.Name, mp.Spec.Template.Spec.InfrastructureRef.Namespace)
		}

		location = amp.Spec.Location
		vmsize = amp.Spec.Template.VMSize
	}

	if location == "" || vmsize == "" {
		return microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool (looking for %q in ns %q)", mp.Spec.Template.Spec.InfrastructureRef.Name, mp.Spec.Template.Spec.InfrastructureRef.Namespace)
	}

	vmcaps, err := h.vmcapsFactory.GetClient(ctx, h.ctrlClient, mp.ObjectMeta)
	if err != nil {
		return microerror.Mask(err)
	}

	supportedZones, err := vmcaps.SupportedAZs(ctx, location, vmsize)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, zone := range mp.Spec.FailureDomains {
		if !inSlice(zone, supportedZones) {
			// Found one unsupported availability zone requested.
			return microerror.Maskf(unsupportedFailureDomainError, "You requested the Machine Pool with type %s to be placed in the following FailureDomains (aka Availability zones): %v but the VM type only supports %v in %s", vmsize, mp.Spec.FailureDomains, supportedZones, location)
		}
	}

	return nil
}

func inSlice(needle string, haystack []string) bool {
	for _, supported := range haystack {
		if needle == supported {
			return true
		}
	}
	return false
}
