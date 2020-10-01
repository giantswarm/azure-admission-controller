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

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	vmcaps    *vmcapabilities.VMSKU
}

type UpdateValidatorConfig struct {
	Logger micrologger.Logger
	VMcaps *vmcapabilities.VMSKU
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}
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

	admitter := &UpdateValidator{
		k8sClient: k8sClient,
		logger:    config.Logger,
		vmcaps:    config.VMcaps,
	}

	return admitter, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureMPNewCR := &expcapzv1alpha3.AzureMachinePool{}
	azureMPOldCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMPNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureMPOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	// Instance type might be changed so we check it is still valid and supported.
	valid, err := checkInstanceTypeIsValid(ctx, a.vmcaps, *azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !valid {
		return false, microerror.Maskf(invalidOperationError, "Instance type is invalid or unsupported")
	}

	if !isAcceleratedNetworkingUnchanged(ctx, *azureMPOldCR, *azureMPNewCR) {
		return false, microerror.Maskf(invalidOperationError, "It is not possible to change the AcceleratedNetworking on an existing node pool")
	}

	isNewVmSizeSupportingAcceleratedNetworking, err := isNewVmSizeSupportingAcceleratedNetworking(ctx, a.vmcaps, *azureMPOldCR, *azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !isNewVmSizeSupportingAcceleratedNetworking {
		return false, microerror.Maskf(invalidOperationError, "This node pool has AcceleratedNetworking enabled. It is not possible to use a VM type that does not support it.")
	}

	return true, nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
