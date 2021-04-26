package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (a *Validator) Validate(ctx context.Context, object interface{}) error {
	cr, err := key.ToAzureMachinePtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = cr.ValidateCreate()
	err = errors.IgnoreCAPIErrorForField("sshPublicKey", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelContainsExistingOrganization(ctx, a.ctrlClient, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkSSHKeyIsEmpty(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateLocation(*cr, a.location)
	if err != nil {
		return microerror.Mask(err)
	}

	supportedAZs, err := a.vmcaps.SupportedAZs(ctx, cr.Spec.Location, cr.Spec.VMSize)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateFailureDomain(*cr, supportedAZs)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
