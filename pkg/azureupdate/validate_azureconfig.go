package azureupdate

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type AzureConfigValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type AzureConfigValidatorConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

const (
	conditionCreating = "Creating"
	conditionUpdating = "Updating"
)

func NewAzureConfigValidator(config AzureConfigValidatorConfig) (*AzureConfigValidator, error) {
	admitter := &AzureConfigValidator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return admitter, nil
}

func (a *AzureConfigValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureConfigNewCR := &v1alpha1.AzureConfig{}
	azureConfigOldCR := &v1alpha1.AzureConfig{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureConfigNewCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse azureConfig CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureConfigOldCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse azureConfig CR: %v", err)
	}

	oldVersion, err := releaseversion.GetVersionFromCRLabels(azureConfigOldCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newVersion, err := releaseversion.GetVersionFromCRLabels(azureConfigNewCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	if !oldVersion.Equals(newVersion) {
		// If tenant cluster is already upgrading, we can't change the version any more.
		upgrading, status := clusterIsUpgrading(azureConfigOldCR)
		if upgrading {
			return false, microerror.Maskf(errors.InvalidOperationError, "cluster has condition: %s", status)
		}

		return releaseversion.UpgradeAllowed(ctx, a.k8sClient.G8sClient(), oldVersion, newVersion)
	}

	return true, nil
}

func (a *AzureConfigValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
