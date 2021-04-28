package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (a *Validator) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureMPNewCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	azureMPOldCR, err := key.ToAzureMachinePoolPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	err = azureMPNewCR.ValidateUpdate(azureMPOldCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(azureMPOldCR, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkInstanceTypeIsValid(ctx, a.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.checkAcceleratedNetworkingUpdateIsValid(ctx, azureMPOldCR, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.checkInstanceTypeChangeIsValid(ctx, azureMPOldCR, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.checkSpotVMOptionsUnchanged(ctx, azureMPOldCR, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.checkStorageAccountTypeUnchanged(ctx, azureMPOldCR, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkSSHKeyIsEmpty(ctx, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkDataDisks(ctx, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkLocationUnchanged(*azureMPOldCR, *azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Validator) checkAcceleratedNetworkingUpdateIsValid(ctx context.Context, azureMPOldCR *capzexp.AzureMachinePool, azureMPNewCR *capzexp.AzureMachinePool) error {
	if hasAcceleratedNetworkingPropertyChanged(ctx, azureMPOldCR, azureMPNewCR) {
		return microerror.Maskf(acceleratedNetworkingWasChangedError, "It is not possible to change the AcceleratedNetworking on an existing node pool")
	}

	if azureMPOldCR.Spec.Template.VMSize == azureMPNewCR.Spec.Template.VMSize {
		return nil
	}

	err := checkAcceleratedNetworking(ctx, a.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Validator) checkInstanceTypeChangeIsValid(ctx context.Context, azureMPOldCR *capzexp.AzureMachinePool, azureMPNewCR *capzexp.AzureMachinePool) error {
	// Check if the instance type has changed.
	if azureMPOldCR.Spec.Template.VMSize != azureMPNewCR.Spec.Template.VMSize {
		oldPremium, err := a.vmcaps.HasCapability(ctx, azureMPOldCR.Spec.Location, azureMPOldCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return microerror.Mask(err)
		}
		newPremium, err := a.vmcaps.HasCapability(ctx, azureMPNewCR.Spec.Location, azureMPNewCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return microerror.Mask(err)
		}

		if oldPremium && !newPremium {
			// We can't downgrade from a VM type supporting premium storage to one that doesn't.
			// Azure doesn't support that.
			return microerror.Maskf(switchToVmSizeThatDoesNotSupportAcceleratedNetworkingError, "Changing the node pool VM type from one that supports accelerated networking to one that does not is unsupported.")
		}
	}

	return nil
}

func (a *Validator) checkSpotVMOptionsUnchanged(ctx context.Context, azureMPOldCR *capzexp.AzureMachinePool, azureMPNewCR *capzexp.AzureMachinePool) error {

	switch {
	case (azureMPOldCR.Spec.Template.SpotVMOptions == nil && azureMPNewCR.Spec.Template.SpotVMOptions == nil):
		return nil
	case (azureMPOldCR.Spec.Template.SpotVMOptions == nil && azureMPNewCR.Spec.Template.SpotVMOptions != nil):
		return microerror.Maskf(spotVMOptionsWasChangedError, "can't enable spot instances for existing machine pool")
	case (azureMPOldCR.Spec.Template.SpotVMOptions != nil && azureMPNewCR.Spec.Template.SpotVMOptions == nil):
		return microerror.Maskf(spotVMOptionsWasChangedError, "can't disable spot instances for existing machine pool")
	case (azureMPOldCR.Spec.Template.SpotVMOptions.MaxPrice == nil && azureMPNewCR.Spec.Template.SpotVMOptions.MaxPrice != nil):
		return microerror.Maskf(spotVMOptionsWasChangedError, "can't change spot instance pricing for existing machine pool")
	case (azureMPOldCR.Spec.Template.SpotVMOptions.MaxPrice != nil && azureMPNewCR.Spec.Template.SpotVMOptions.MaxPrice == nil):
		return microerror.Maskf(spotVMOptionsWasChangedError, "can't change spot instance pricing for existing machine pool")
	case (azureMPOldCR.Spec.Template.SpotVMOptions.MaxPrice == nil && azureMPNewCR.Spec.Template.SpotVMOptions.MaxPrice == nil):
		return nil
	case (*azureMPOldCR.Spec.Template.SpotVMOptions.MaxPrice != *azureMPNewCR.Spec.Template.SpotVMOptions.MaxPrice):
		return microerror.Maskf(spotVMOptionsWasChangedError, "can't change spot instance pricing for existing machine pool")
	}

	return nil
}

// Checks if the storage account type of the osDisk is changed. This is never allowed.
func (a *Validator) checkStorageAccountTypeUnchanged(ctx context.Context, azureMPOldCR *capzexp.AzureMachinePool, azureMPNewCR *capzexp.AzureMachinePool) error {
	if azureMPOldCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType != azureMPNewCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType {
		return microerror.Maskf(storageAccountWasChangedError, "Changing the storage account type of the OS disk is not allowed.")
	}

	return nil
}
