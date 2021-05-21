package mutator

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type WebhookHandlerBase interface {
	generic.Decoder
	generic.Logger
	Resource() string
}

type WebhookCreateHandler interface {
	WebhookHandlerBase
	OnCreateMutate(ctx context.Context, object interface{}) ([]PatchOperation, error)
}

type WebhookUpdateHandler interface {
	WebhookHandlerBase
	OnUpdateMutate(ctx context.Context, oldObject interface{}, object interface{}) ([]PatchOperation, error)
}
