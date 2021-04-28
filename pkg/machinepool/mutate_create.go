package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	machinePoolCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	machinePoolCROriginal := machinePoolCR.DeepCopy()

	defaultSpecValues := setDefaultSpecValues(h, machinePoolCR)
	if defaultSpecValues != nil {
		result = append(result, defaultSpecValues...)
	}

	patch, err := mutator.EnsureReleaseVersionLabel(ctx, h.ctrlClient, machinePoolCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, h.ctrlClient, machinePoolCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	autoscalingPatches := ensureAutoscalingAnnotations(h, machinePoolCR)
	if autoscalingPatches != nil {
		result = append(result, autoscalingPatches...)
	}

	machinePoolCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = mutator.GenerateFromObjectDiff(machinePoolCROriginal, machinePoolCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		result = append(result, capiPatches...)
	}

	return result, nil
}
