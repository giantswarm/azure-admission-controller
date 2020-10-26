package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type UpdateValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &UpdateValidator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureClusterNewCR := &capzv1alpha3.AzureCluster{}
	azureClusterOldCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureClusterNewCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureClusterOldCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	err := generic.ValidateOrganizationLabelUnchanged(azureClusterOldCR, azureClusterNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = validateControlPlaneEndpointUnchanged(*azureClusterOldCR, *azureClusterNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}

	oldClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterOldCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterNewCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}