package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) OnUpdateMutate(ctx context.Context, _ interface{}, object interface{}) ([]mutator.PatchOperation, error) {
	var err error
	var result []mutator.PatchOperation
	machinePoolCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	machinePoolCROriginal := machinePoolCR.DeepCopy()

	// Values for some optional spec fields could be removed in a CR update, so
	// here we set them back again to their default values.
	defaultSpecValues := setDefaultSpecValues(m, machinePoolCR)
	if defaultSpecValues != nil {
		result = append(result, defaultSpecValues...)
	}

	// Ensure autoscaling annotations are set.
	patch := ensureAutoscalingAnnotations(m, machinePoolCR)
	if patch != nil {
		result = append(result, patch...)
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
