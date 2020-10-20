package cluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutatorConfig struct {
	Logger micrologger.Logger
}

type CreateMutator struct {
	logger micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &CreateMutator{
		logger: config.Logger,
	}

	return m, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	newCluster := &capiv1alpha3.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, newCluster); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	defaultStatusPatch, err := m.getDefaultStatusPatch(ctx)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}

	result = append(result, *defaultStatusPatch)
	return result, nil
}

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "cluster"
}
