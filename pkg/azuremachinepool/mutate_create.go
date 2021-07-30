package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureMPCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureMPCROriginal := azureMPCR.DeepCopy()

	patch, err := h.ensureLocation(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = h.ensureStorageAccountType(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = h.ensureDataDisks(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureReleaseVersionLabel(ctx, h.ctrlClient, azureMPCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, h.ctrlClient, azureMPCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureMPCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFromObjectDiff(azureMPCROriginal, azureMPCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/template/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (h *WebhookHandler) ensureStorageAccountType(ctx context.Context, mpCR *capzexp.AzureMachinePool) (*mutator.PatchOperation, error) {
	if mpCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType == "" {
		// We need to set the default value as it is missing.

		location := mpCR.Spec.Location
		if location == "" {
			// The location was empty and we are adding it using this same mutator.
			// We assume it will be set to the installation's location.
			location = h.location
		}

		// Check if the VM has Premium Storage capability.
		premium, err := h.vmcaps.HasCapability(ctx, location, mpCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
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

func (h *WebhookHandler) ensureDataDisks(_ context.Context, mpCR *capzexp.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Template.DataDisks) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/template/dataDisks", desiredDataDisks), nil
}

func (h *WebhookHandler) ensureLocation(_ context.Context, mpCR *capzexp.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Location) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/location", h.location), nil
}
