package machinepool

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	logger micrologger.Logger
}

type UpdateValidatorConfig struct {
	Logger micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	admitter := &UpdateValidator{
		logger: config.Logger,
	}

	return admitter, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) (bool, error) {
	machinePoolNewCR := &v1alpha3.MachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, machinePoolNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse machinePool CR: %v", err)
	}
	machinePoolOldCR := &v1alpha3.MachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, machinePoolOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse machinePool CR: %v", err)
	}

	err := checkAvailabilityZonesUnchanged(ctx, machinePoolOldCR, machinePoolNewCR)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func checkAvailabilityZonesUnchanged(ctx context.Context, oldMP *v1alpha3.MachinePool, newMP *v1alpha3.MachinePool) error {
	if !reflect.DeepEqual(oldMP.Spec.FailureDomains, newMP.Spec.FailureDomains) {
		return microerror.Maskf(invalidOperationError, "Changing FailureDomains (availability zones) is not allowed.")
	}

	return nil
}
