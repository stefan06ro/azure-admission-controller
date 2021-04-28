package spark

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type WebhookHandler struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type WebhookHandlerConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewWebhookHandler(config WebhookHandlerConfig) (*WebhookHandler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &WebhookHandler{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return m, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "spark"
}

func (h *WebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	sparkCR := &v1alpha1.Spark{}
	if _, _, err := mutator.Deserializer.Decode(rawObject.Raw, nil, sparkCR); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse Spark CR: %v", err)
	}

	return sparkCR, nil
}
