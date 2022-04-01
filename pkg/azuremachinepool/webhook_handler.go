package azuremachinepool

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/v2/internal/errors"
	"github.com/giantswarm/azure-admission-controller/v2/internal/vmcapabilities"
)

type WebhookHandler struct {
	ctrlClient    client.Client
	decoder       runtime.Decoder
	location      string
	logger        micrologger.Logger
	vmcapsFactory vmcapabilities.Factory
}

type WebhookHandlerConfig struct {
	CtrlClient    client.Client
	Decoder       runtime.Decoder
	Location      string
	Logger        micrologger.Logger
	VMcapsFactory vmcapabilities.Factory
}

func NewWebhookHandler(config WebhookHandlerConfig) (*WebhookHandler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Decoder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Decoder must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcapsFactory == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcapsFactory must not be empty", config)
	}

	handler := &WebhookHandler{
		ctrlClient:    config.CtrlClient,
		decoder:       config.Decoder,
		location:      config.Location,
		logger:        config.Logger,
		vmcapsFactory: config.VMcapsFactory,
	}

	return handler, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "azuremachinepool"
}

func (h *WebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &capzexp.AzureMachinePool{}
	if _, _, err := h.decoder.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureMachinePool CR: %v", err)
	}

	return cr, nil
}
