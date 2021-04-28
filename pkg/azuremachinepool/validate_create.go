package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
	azureMPNewCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = azureMPNewCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelMatchesCluster(ctx, h.ctrlClient, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkInstanceTypeIsValid(ctx, h.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkAcceleratedNetworking(ctx, h.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkStorageAccountTypeIsValid(ctx, h.vmcaps, azureMPNewCR)
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

	err = checkLocation(*azureMPNewCR, h.location)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
