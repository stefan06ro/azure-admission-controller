package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) OnUpdateMutate(ctx context.Context, _ interface{}, object interface{}) ([]mutator.PatchOperation, error) {
	var err error
	var result []mutator.PatchOperation
	azureMPCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureMPCROriginal := azureMPCR.DeepCopy()

	azureMPCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = mutator.GenerateFromObjectDiff(azureMPCROriginal, azureMPCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = mutator.SkipForPath("/spec/template/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}
