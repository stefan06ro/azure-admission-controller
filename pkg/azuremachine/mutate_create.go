package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureMachineCR, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureMachineCROriginal := azureMachineCR.DeepCopy()

	patch, err := m.ensureLocation(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureOSDiskCachingType(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, azureMachineCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureMachineCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = mutator.GenerateFromObjectDiff(azureMachineCROriginal, azureMachineCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = mutator.SkipForPath("/spec/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *Mutator) ensureLocation(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if azureMachine.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", m.location), nil
	}

	return nil, nil
}
