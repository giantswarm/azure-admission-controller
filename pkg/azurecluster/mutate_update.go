package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type UpdateMutator struct {
	ctrlCache  ctrl.Reader
	ctrlClient ctrl.Client
	logger     micrologger.Logger
}

type UpdateMutatorConfig struct {
	CtrlCache  ctrl.Reader
	CtrlClient ctrl.Client
	Logger     micrologger.Logger
}

func NewUpdateMutator(config UpdateMutatorConfig) (*UpdateMutator, error) {
	if config.CtrlCache == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlCache must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &UpdateMutator{
		ctrlCache:  config.CtrlCache,
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

	azureClusterCR := &capz.AzureCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if capi {
		return []mutator.PatchOperation{}, nil
	}

	patch, err := ensureAPIServerLB(azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

<<<<<<< HEAD
	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, m.logger, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
=======
	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlCache, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
>>>>>>> master
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

<<<<<<< HEAD
	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, m.logger, azureClusterCR.GetObjectMeta(), "cluster-operator", label.ClusterOperatorVersion)
=======
	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlCache, azureClusterCR.GetObjectMeta(), "cluster-operator", label.ClusterOperatorVersion)
>>>>>>> master
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureClusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFrom(request.Object.Raw, azureClusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/networkSpec/vnet", capiPatches)
		capiPatches = patches.SkipForPath("/spec/networkSpec/subnets", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *UpdateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *UpdateMutator) Resource() string {
	return "azurecluster"
}
