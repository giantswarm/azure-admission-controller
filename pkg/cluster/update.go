package cluster

import (
	"context"

	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type UpdateValidatorConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	v := &UpdateValidator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return v, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	clusterNewCR := &capiv1alpha3.Cluster{}
	clusterOldCR := &capiv1alpha3.Cluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, clusterNewCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, clusterOldCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	oldClusterVersion, err := releaseversion.GetVersionFromCRLabels(clusterOldCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := releaseversion.GetVersionFromCRLabels(clusterNewCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	return releaseversion.UpgradeAllowed(ctx, a.k8sClient.G8sClient(), oldClusterVersion, newClusterVersion)
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
