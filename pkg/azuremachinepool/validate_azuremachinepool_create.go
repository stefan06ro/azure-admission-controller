package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	admitter := &CreateValidator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return admitter, nil
}

func (a *CreateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	azureMPNewCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMPNewCR); err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}
	if capi {
		return nil
	}

	err = azureMPNewCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelMatchesCluster(ctx, a.ctrlClient, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkInstanceTypeIsValid(ctx, a.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkAcceleratedNetworking(ctx, a.vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkStorageAccountTypeIsValid(ctx, a.vmcaps, azureMPNewCR)
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

	err = checkLocation(*azureMPNewCR, a.location)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
