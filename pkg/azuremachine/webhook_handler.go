package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type WebhookHandler struct {
	Decoder

	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type WebhookHandlerConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
	VMcaps     *vmcapabilities.VMSKU
}

func NewWebhookHandler(config WebhookHandlerConfig) (*WebhookHandler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	v := &WebhookHandler{
		Decoder: Decoder{},

		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return v, nil
}

func (h *WebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *WebhookHandler) Resource() string {
	return "azuremachine"
}

func (h *WebhookHandler) ensureOSDiskCachingType(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if len(azureMachine.Spec.OSDisk.CachingType) < 1 {
		return mutator.PatchAdd("/spec/osDisk/cachingType", key.OSDiskCachingType()), nil
	}

	return nil, nil
}
