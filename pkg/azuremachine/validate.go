package azuremachine

import (
	"context"
	"fmt"
	"strings"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type Config struct {
	Logger micrologger.Logger
}

func NewAzureClusterConfigValidator(config Config) (*Validator, error) {
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load key kubeconfig: %v", err)
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
				releasev1alpha1.AddToScheme,
			},
			Logger: config.Logger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	validator := &Validator{
		k8sClient: k8sClient,
		logger:    config.Logger,
	}

	return validator, nil
}

func (a *Validator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	AzureMachineNewCR := &capz.AzureMachine{}
	AzureMachineOldCR := &capz.AzureMachine{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, AzureMachineNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, AzureMachineOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}

	oldClusterVersion := AzureMachineOldCR.Labels[label.ReleaseVersion]
	newClusterVersion := AzureMachineNewCR.Labels[label.ReleaseVersion]

	if oldClusterVersion != newClusterVersion {
		if isAlphaRelease(oldClusterVersion) || isAlphaRelease(newClusterVersion) {
			return false, microerror.Maskf(invalidOperationError, "It is not possible to upgrade to or from an alpha release")
		}
	}

	return true, nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func isAlphaRelease(release string) bool {
	return strings.Contains(release, "alpha")
}
