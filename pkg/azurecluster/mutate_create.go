package azurecluster

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	baseDomain string
	logger     micrologger.Logger
}

type CreateMutatorConfig struct {
	BaseDomain string
	Logger     micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &CreateMutator{
		baseDomain: config.BaseDomain,
		logger:     config.Logger,
	}

	return v, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	patch, err := m.ensureControlPlaneEndpointHost(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "azurecluster"
}

func (m *CreateMutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", fmt.Sprintf("api.%s.%s", clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", 443), nil
	}

	return nil, nil
}
