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

const (
	minMemory = 16
	minCPUs   = 4
)

type CreateValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	vmcaps    *vmcapabilities.VMSKU
}

type CreateValidatorConfig struct {
	Logger micrologger.Logger
	VMcaps *vmcapabilities.VMSKU
}

func NewCreateValidator(config CreateValidatorConfig) (*CreateValidator, error) {
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

	// Check if instance type is valid and supported.
	valid, err := a.CheckInstanceTypeIsValid(ctx, *azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !valid {
		return false, microerror.Maskf(invalidOperationError, "Instance type is invalid or unsupported")
	}

	valid, err = a.CheckAcceleratedNetworking(ctx, *azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !valid {
		return false, microerror.Maskf(invalidOperationError, "Accelerated Networking is not supported by the selected machine type")
	}

	return false, microerror.Maskf(invalidOperationError, "Instance type does not support accelerated networking")
}

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func (a *CreateValidator) CheckInstanceTypeIsValid(ctx context.Context, mp expcapzv1alpha3.AzureMachinePool) (bool, error) {
	memory, err := a.vmcaps.Memory(ctx, mp.Spec.Location, mp.Spec.Template.VMSize)
	if err != nil {
		return false, microerror.Mask(err)
	}

	cpu, err := a.vmcaps.CPUs(ctx, mp.Spec.Location, mp.Spec.Template.VMSize)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if memory < minMemory {
		return false, nil
	}

	if cpu < minCPUs {
		return false, nil
	}

	return true, nil
}

func (a *CreateValidator) CheckAcceleratedNetworking(ctx context.Context, mp expcapzv1alpha3.AzureMachinePool) (bool, error) {
	// If the instance type is invalid, the following function returns an error.
	acceleratedNetworkingAvailable, err := a.vmcaps.HasCapability(ctx, mp.Spec.Location, mp.Spec.Template.VMSize, vmcapabilities.CapabilityAcceleratedNetworking)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// Accelerated networking is disabled (false) or in auto-detect mode (nil). This is always allowed.
	if mp.Spec.Template.AcceleratedNetworking == nil || !*mp.Spec.Template.AcceleratedNetworking {
		return true, nil
	}

	// Accelerated networking is enabled (true).
	return acceleratedNetworkingAvailable, nil
}
