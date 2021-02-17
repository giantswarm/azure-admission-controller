package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type UpdateMutator struct {
	ctrlClient ctrl.Client
	logger     micrologger.Logger
}

type UpdateMutatorConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger
}

func NewUpdateMutator(config UpdateMutatorConfig) (*UpdateMutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &UpdateMutator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return m, nil
}

func (m *UpdateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureMachineCR := &v1alpha3.AzureMachine{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureMachineCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}

	patch, err := m.ensureOSDiskCachingType(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureMachineCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFrom(request.Object.Raw, azureMachineCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *UpdateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *UpdateMutator) Resource() string {
	return "azuremachine"
}

func (m *UpdateMutator) ensureOSDiskCachingType(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if len(azureMachine.Spec.OSDisk.CachingType) < 1 {
		return mutator.PatchAdd("/spec/osDisk/cachingType", key.OSDiskCachingType()), nil
	}

	return nil, nil
}
