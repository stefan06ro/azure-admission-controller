package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func checkStorageAccountTypeIsValid(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMachinePool *expcapzv1alpha3.AzureMachinePool) error {
	selectedStorageAccount := azureMachinePool.Spec.Template.OSDisk.ManagedDisk.StorageAccountType

	if selectedStorageAccount != string(compute.StorageAccountTypesStandardLRS) &&
		selectedStorageAccount != string(compute.StorageAccountTypesPremiumLRS) {
		// Storage account type is invalid.
		return microerror.Maskf(invalidOperationError, "Storage account type %q is invalid. Allowed values are %q and %q", selectedStorageAccount, string(compute.StorageAccountTypesStandardLRS), string(compute.StorageAccountTypesPremiumLRS))
	}

	// Storage account type is valid, check if it matches the VM type's support.
	if selectedStorageAccount == string(compute.StorageAccountTypesPremiumLRS) {
		// Premium is selected, VM type has to support it.
		supported, err := vmcaps.HasCapability(ctx, azureMachinePool.Spec.Location, azureMachinePool.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return microerror.Mask(err)
		}

		if !supported {
			return microerror.Maskf(invalidOperationError, "VM Type %s does not support Premium Storage", azureMachinePool.Spec.Template.VMSize)
		}
	}

	return nil
}
