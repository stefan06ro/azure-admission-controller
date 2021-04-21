package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type Validator struct {
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type ValidatorConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
	VMcaps     *vmcapabilities.VMSKU
}

func NewCreateValidator(config ValidatorConfig) (*Validator, error) {
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

	admitter := &Validator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return admitter, nil
}

func (a *Validator) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &capzexp.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureMachinePool CR: %v", err)
	}

	return cr, nil
}

func (a *Validator) Validate(ctx context.Context, object interface{}) error {
	azureMPNewCR, err := key.ToAzureMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
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

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
