package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
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

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	azureMPNewCR := &expcapzv1alpha3.AzureMachinePool{}
	azureMPOldCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureMPNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureMPOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	err := checkInstanceTypeIsValid(ctx, a.vmcaps, azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = checkAcceleratedNetworkingUpdateIsValid(ctx, a.vmcaps, azureMPOldCR, azureMPNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func checkAcceleratedNetworkingUpdateIsValid(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMPOldCR *expcapzv1alpha3.AzureMachinePool, azureMPNewCR *expcapzv1alpha3.AzureMachinePool) error {
	if hasAcceleratedNetworkingPropertyChanged(ctx, azureMPOldCR, azureMPNewCR) {
		return microerror.Maskf(invalidOperationError, "It is not possible to change the AcceleratedNetworking on an existing node pool")
	}

	if azureMPOldCR.Spec.Template.VMSize == azureMPNewCR.Spec.Template.VMSize {
		return nil
	}

	err := checkAcceleratedNetworking(ctx, vmcaps, azureMPNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
