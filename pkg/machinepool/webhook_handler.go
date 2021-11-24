package machinepool

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

type WebhookHandler struct {
	ctrlClient    client.Client
	decoder       runtime.Decoder
	logger        micrologger.Logger
	vmcapsFactory vmcapabilities.Factory
}

type WebhookHandlerConfig struct {
	CtrlClient    client.Client
	Decoder       runtime.Decoder
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
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcapsFactory == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcapsFactory must not be empty", config)
	}

	handler := &WebhookHandler{
		ctrlClient:    config.CtrlClient,
		decoder:       config.Decoder,
		logger:        config.Logger,
		vmcapsFactory: config.VMcapsFactory,
	}

	return handler, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "machinepool"
}

func (h *WebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &capiexp.MachinePool{}
	if _, _, err := h.decoder.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse MachinePool CR: %v", err)
	}

	return cr, nil
}
