package cluster

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

type WebhookHandler struct {
	baseDomain string
	decoder    runtime.Decoder
	ctrlReader client.Reader
	ctrlClient client.Client
	logger     micrologger.Logger
}

type WebhookHandlerConfig struct {
	BaseDomain string
	Decoder    runtime.Decoder
	CtrlReader client.Reader
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewWebhookHandler(config WebhookHandlerConfig) (*WebhookHandler, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.Decoder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Decoder must not be empty", config)
	}
	if config.CtrlReader == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlReader must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &WebhookHandler{
		baseDomain: config.BaseDomain,
		decoder:    config.Decoder,
		ctrlReader: config.CtrlReader,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return v, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "cluster"
}

func (h *WebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &capi.Cluster{}
	if _, _, err := h.decoder.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	return cr, nil
}
