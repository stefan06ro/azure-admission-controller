package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnUpdateMutate(_ context.Context, _ interface{}, object interface{}) ([]mutator.PatchOperation, error) {
	var err error
	var result []mutator.PatchOperation
	machinePoolCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	machinePoolCROriginal := machinePoolCR.DeepCopy()

	// Ensure autoscaling annotations are set.
	patch := ensureAutoscalingAnnotations(h, machinePoolCR)
	if patch != nil {
		result = append(result, patch...)
	}

	machinePoolCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFromObjectDiff(machinePoolCROriginal, machinePoolCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		result = append(result, capiPatches...)
	}

	return result, nil
}
