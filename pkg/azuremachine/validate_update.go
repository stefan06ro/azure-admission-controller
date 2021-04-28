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

func (a *Validator) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureMachineNewCR, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return microerror.Mask(err)
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
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureMachine (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureMachineNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureMachine (after edit)")
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}
