package spark

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	sparkCR, err := key.ToSparkPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}

	patch, err := mutator.EnsureReleaseVersionLabel(ctx, h.ctrlClient, sparkCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}
