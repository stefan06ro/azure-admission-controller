package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type CreateValidator struct {
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type CreateValidatorConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
	VMcaps     *vmcapabilities.VMSKU
}

func NewCreateValidator(config CreateValidatorConfig) (*CreateValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	v := &CreateValidator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return v, nil
}

func (a *CreateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	cr := &capzv1alpha3.AzureMachine{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, cr); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(cr)
	if err != nil {
		return microerror.Mask(err)
	}
	if capi {
		return nil
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

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
