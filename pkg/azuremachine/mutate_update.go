package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnUpdateMutate(ctx context.Context, _ interface{}, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureMachineCR, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureMachineCROriginal := azureMachineCR.DeepCopy()

	patch, err := h.ensureOSDiskCachingType(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureMachineCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFromObjectDiff(azureMachineCROriginal, azureMachineCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}
