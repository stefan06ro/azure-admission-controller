package machinepool

import (
	"context"
	"sort"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
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

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	machinePoolNewCR := &v1alpha3.MachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, machinePoolNewCR); err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse machinePool CR: %v", err)
	}
	machinePoolOldCR := &v1alpha3.MachinePool{}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, machinePoolOldCR); err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse machinePool CR: %v", err)
	}

	err := generic.ValidateOrganizationLabelUnchanged(machinePoolOldCR, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkAvailabilityZonesUnchanged(ctx, machinePoolOldCR, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func checkAvailabilityZonesUnchanged(ctx context.Context, oldMP *v1alpha3.MachinePool, newMP *v1alpha3.MachinePool) error {
	if len(oldMP.Spec.FailureDomains) != len(newMP.Spec.FailureDomains) {
		return microerror.Maskf(failureDomainWasChangedError, "Changing FailureDomains (availability zones) is not allowed.")
	}

	sort.Strings(oldMP.Spec.FailureDomains)
	sort.Strings(newMP.Spec.FailureDomains)

	for i := 0; i < len(oldMP.Spec.FailureDomains); i++ {
		if oldMP.Spec.FailureDomains[i] != newMP.Spec.FailureDomains[i] {
			return microerror.Maskf(failureDomainWasChangedError, "Changing FailureDomains (availability zones) is not allowed.")
		}
	}

	return nil
}
