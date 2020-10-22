package cluster

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

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

	clusterCR := &capiv1alpha3.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, clusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	patch, err := m.ensureClusterNetwork(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureServiceDomain(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointHost(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, clusterCR)
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
	return "cluster"
}

func (m *CreateMutator) ensureClusterNetwork(ctx context.Context, clusterCR *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	// Ensure ClusterNetwork is set.
	if clusterCR.Spec.ClusterNetwork == nil {
		clusterNetwork := capiv1alpha3.ClusterNetwork{
			APIServerPort: to.Int32Ptr(443),
			Services: &capiv1alpha3.NetworkRanges{
				CIDRBlocks: []string{
					"172.31.0.0/16",
				},
			},
		}

		return mutator.PatchAdd("/spec/clusterNetwork", clusterNetwork), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureServiceDomain(ctx context.Context, clusterCR *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	// Ensure ServiceDomain is set.
	if clusterCR.Spec.ClusterNetwork.ServiceDomain == "" {
		return mutator.PatchAdd("/spec/clusterNetwork/serviceDomain", fmt.Sprintf("%s.%s", clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", fmt.Sprintf("api.%s.%s", clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", 443), nil
	}

	return nil, nil
}
