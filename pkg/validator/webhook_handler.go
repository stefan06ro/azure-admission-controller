package validator

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type WebhookHandlerBase interface {
	generic.Decoder
	generic.Logger
}

type WebhookCreateHandler interface {
	WebhookHandlerBase
	OnCreateValidate(ctx context.Context, object interface{}) error
}

type WebhookUpdateHandler interface {
	WebhookHandlerBase
	OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error
}
