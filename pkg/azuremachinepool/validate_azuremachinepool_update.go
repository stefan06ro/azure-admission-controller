package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	logger micrologger.Logger
	vmcaps *vmcapabilities.VMSKU
}

type UpdateValidatorConfig struct {
	Logger micrologger.Logger
	VMcaps *vmcapabilities.VMSKU
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	admitter := &UpdateValidator{
		logger: config.Logger,
		vmcaps: config.VMcaps,
	}

	return admitter, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	azureMPNewCR := &expcapzv1alpha3.AzureMachinePool{}
	azureMPOldCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMPNewCR); err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureMPOldCR); err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	err := generic.ValidateOrganizationLabelUnchanged(azureMPOldCR, azureMPNewCR)
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

	return nil
}

func (a *UpdateValidator) checkAcceleratedNetworkingUpdateIsValid(ctx context.Context, azureMPOldCR *expcapzv1alpha3.AzureMachinePool, azureMPNewCR *expcapzv1alpha3.AzureMachinePool) error {
	if hasAcceleratedNetworkingPropertyChanged(ctx, azureMPOldCR, azureMPNewCR) {
		return microerror.Maskf(invalidOperationError, "It is not possible to change the AcceleratedNetworking on an existing node pool")
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

func (a *UpdateValidator) checkInstanceTypeChangeIsValid(ctx context.Context, azureMPOldCR *expcapzv1alpha3.AzureMachinePool, azureMPNewCR *expcapzv1alpha3.AzureMachinePool) error {
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
			return microerror.Maskf(invalidOperationError, "Changing the node pool VM type from one that supports accelerated networking to one that does not is unsupported.")
		}
	}

	return nil
}

// Checks if the storage account type of the osDisk is changed. This is never allowed.
func (a *UpdateValidator) checkStorageAccountTypeUnchanged(ctx context.Context, azureMPOldCR *expcapzv1alpha3.AzureMachinePool, azureMPNewCR *expcapzv1alpha3.AzureMachinePool) error {
	if azureMPOldCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType != azureMPNewCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType {
		return microerror.Maskf(invalidOperationError, "Changing the storage account type of the OS disk is not allowed.")
	}

	return nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
