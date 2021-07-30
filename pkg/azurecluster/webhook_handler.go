package azurecluster

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type WebhookHandler struct {
	baseDomain string
	ctrlCache  client.Reader
	ctrlClient client.Client
	decoder    runtime.Decoder
	location   string
	logger     micrologger.Logger
}

type WebhookHandlerConfig struct {
	BaseDomain string
	CtrlCache  client.Reader
	CtrlClient client.Client
	Decoder    runtime.Decoder
	Location   string
	Logger     micrologger.Logger
}

func NewWebhookHandler(config WebhookHandlerConfig) (*WebhookHandler, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlCache == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlCache must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Decoder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Decoder must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}

	v := &WebhookHandler{
		baseDomain: config.BaseDomain,
		ctrlCache:  config.CtrlCache,
		ctrlClient: config.CtrlClient,
		decoder:    config.Decoder,
		location:   config.Location,
		logger:     config.Logger,
	}

	return v, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "azurecluster"
}

func (h *WebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	azureClusterCR := &capz.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, azureClusterCR); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	return azureClusterCR, nil
}
