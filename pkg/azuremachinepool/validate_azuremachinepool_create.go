package azuremachinepool

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	restclient "k8s.io/client-go/rest"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type CreateValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type CreateValidatorConfig struct {
	Logger micrologger.Logger
}

func NewCreateValidator(config CreateValidatorConfig) (*CreateValidator, error) {
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load key kubeconfig: %v", err)
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha3.AddToScheme,
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

	admitter := &CreateValidator{
		k8sClient: k8sClient,
		logger:    config.Logger,
	}

	return admitter, nil
}

func (a *CreateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureMPNewCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMPNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	// If the instance type is invalid, the following function returns an error.
	capabilities, err := vmcapabilities.FromInstanceType(azureMPNewCR.Spec.Template.VMSize)
	if err != nil {
		return false, microerror.Maskf(invalidOperationError, "Instance type is invalid or unsupported")
	}

	// Accelerated networking is disabled. Always allowed.
	if azureMPNewCR.Spec.Template.AcceleratedNetworking == nil || !*azureMPNewCR.Spec.Template.AcceleratedNetworking {
		return true, nil
	}

	if capabilities.SupportsAcceleratedNetworking {
		return true, nil
	}

	return false, microerror.Maskf(invalidOperationError, "Instance type does not support accelerated networking")
}

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
