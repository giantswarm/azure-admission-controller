package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type CreateValidator struct {
	baseDomain string
	logger     micrologger.Logger
}

type CreateValidatorConfig struct {
	BaseDomain string
	Logger     micrologger.Logger
}

func NewCreateValidator(config CreateValidatorConfig) (*CreateValidator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &CreateValidator{
		baseDomain: config.BaseDomain,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *CreateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	err := validateControlPlaneEndpoint(*azureClusterCR, a.baseDomain)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
