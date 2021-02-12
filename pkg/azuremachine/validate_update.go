package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type UpdateValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &UpdateValidator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	azureMachineNewCR := &capzv1alpha3.AzureMachine{}
	azureMachineOldCR := &capzv1alpha3.AzureMachine{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMachineNewCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureMachineOldCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}

	err := azureMachineNewCR.ValidateUpdate(azureMachineOldCR)
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

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
