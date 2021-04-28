package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureMPCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureMPCROriginal := azureMPCR.DeepCopy()

	patch, err := m.ensureLocation(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureStorageAccountType(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureDataDisks(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureReleaseVersionLabel(ctx, m.ctrlClient, azureMPCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, azureMPCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

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

func (m *Mutator) ensureStorageAccountType(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if mpCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType == "" {
		// We need to set the default value as it is missing.

		location := mpCR.Spec.Location
		if location == "" {
			// The location was empty and we are adding it using this same mutator.
			// We assume it will be set to the installation's location.
			location = m.location
		}

		// Check if the VM has Premium Storage capability.
		premium, err := m.vmcaps.HasCapability(ctx, location, mpCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var storageAccountType string
		{
			if premium {
				storageAccountType = string(compute.StorageAccountTypesPremiumLRS)
			} else {
				storageAccountType = string(compute.StorageAccountTypesStandardLRS)
			}
		}

		return mutator.PatchAdd("/spec/template/osDisk/managedDisk/storageAccountType", storageAccountType), nil
	}

	return nil, nil
}

func (m *Mutator) ensureDataDisks(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Template.DataDisks) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/template/dataDisks", desiredDataDisks), nil
}

func (m *Mutator) ensureLocation(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Location) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/location", m.location), nil
}
