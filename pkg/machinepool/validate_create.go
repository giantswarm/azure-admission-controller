package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"

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

	return nil
}
