package cluster

import (
	"context"
	"strings"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	logger micrologger.Logger
}

type UpdateValidatorConfig struct {
	Logger micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	v := &UpdateValidator{
		logger: config.Logger,
	}

	return v, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	clusterNewCR := &capiv1alpha3.Cluster{}
	clusterOldCR := &capiv1alpha3.Cluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, clusterNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse Cluster CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, clusterOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	oldClusterVersion := clusterOldCR.Labels[label.ReleaseVersion]
	newClusterVersion := clusterNewCR.Labels[label.ReleaseVersion]

	if oldClusterVersion != newClusterVersion {
		if isAlphaRelease(oldClusterVersion) || isAlphaRelease(newClusterVersion) {
			return false, microerror.Maskf(invalidOperationError, "It is not possible to upgrade to or from an alpha release")
		}
	}

	return true, nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func isAlphaRelease(release string) bool {
	return strings.Contains(release, "alpha")
}
