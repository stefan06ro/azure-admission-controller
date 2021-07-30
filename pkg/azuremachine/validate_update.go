package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureMachineNewCR, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	if !azureMachineNewCR.GetDeletionTimestamp().IsZero() {
		h.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
	}

	azureMachineOldCR, err := key.ToAzureMachinePtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	err = azureMachineNewCR.ValidateUpdate(azureMachineOldCR)
	err = errors.IgnoreCAPIErrorForField("sshPublicKey", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkSSHKeyIsEmpty(ctx, azureMachineNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(azureMachineOldCR, azureMachineNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateLocationUnchanged(*azureMachineOldCR, *azureMachineNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateFailureDomainUnchanged(*azureMachineOldCR, *azureMachineNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	oldClusterVersion, err := semverhelper.GetSemverFromLabels(azureMachineOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureMachineNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	return releaseversion.Validate(ctx, h.ctrlClient, oldClusterVersion, newClusterVersion)
}
