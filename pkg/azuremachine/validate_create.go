package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/v2/internal/errors"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
	cr, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = cr.ValidateCreate()
	err = errors.IgnoreCAPIErrorForField("sshPublicKey", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelContainsExistingOrganization(ctx, h.ctrlClient, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkSSHKeyIsEmpty(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	vmcaps, err := h.vmcapsFactory.GetClient(ctx, h.ctrlClient, cr.ObjectMeta)
	if err != nil {
		return microerror.Mask(err)
	}

	supportedAZs, err := vmcaps.SupportedAZs(ctx, h.location, cr.Spec.VMSize)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateFailureDomain(*cr, supportedAZs)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
