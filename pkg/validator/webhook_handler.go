package validator

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type WebhookHandler interface {
	generic.Decoder
	generic.Logger
}

type WebhookCreateHandler interface {
	WebhookHandler
	OnCreateValidate(ctx context.Context, object interface{}) error
}

type WebhookUpdateHandler interface {
	WebhookHandler
	OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error
}
