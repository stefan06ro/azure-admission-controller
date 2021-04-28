package mutator

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type WebhookHandler interface {
	generic.Decoder
	generic.Logger
	Resource() string
}

type WebhookCreateHandler interface {
	WebhookHandler
	OnCreateMutate(ctx context.Context, object interface{}) ([]PatchOperation, error)
}

type WebhookUpdateHandler interface {
	WebhookHandler
	OnUpdateMutate(ctx context.Context, oldObject interface{}, object interface{}) ([]PatchOperation, error)
}
