package conditions

import (
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

type conditionChangeValidation func(oldCondition, newCondition *capi.Condition) error

func validateAll(oldCondition, newCondition *capi.Condition, validations []conditionChangeValidation) error {
	var err error

	for _, validation := range validations {
		err = validation(oldCondition, newCondition)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func validateNewStatusUnknownIsNotAllowed(_, newCondition *capi.Condition) error {
	// Rule: Setting Unknown condition status is not allowed.
	// Why: Unknown is treated as if the cluster condition is not set at all.
	//      If we are already setting the condition status, we should be able
	//      to tell if it is True or False.
	if newCondition != nil && newCondition.Status == corev1.ConditionUnknown {
		errorMessageFormat := "Setting Unknown status to %s condition is not allowed."
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessageFormat, newCondition.Type)
	}

	return nil
}

func validateNewStatusMustBeEitherTrueOrFalse(_, newCondition *capi.Condition) error {
	if newCondition == nil {
		// Condition is not set at all, so nothing to check.
		return nil
	}

	// Rule: We only allow True and False as condition status values.
	// Why: Condition status is practically a string, so some unsupported
	//      values could be set by mistake, which can lead to unexpected
	//      behavior.
	conditionStatusNotAllowed :=
		newCondition.Status != corev1.ConditionTrue &&
			newCondition.Status != corev1.ConditionFalse

	if conditionStatusNotAllowed {
		errorMessageFormat := "Allowed values for %s condition status are True and False, got %s."
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessageFormat, newCondition.Type, newCondition.Status)
	}

	return nil
}

func validateRemovingConditionIsNotAllowed(oldCondition, newCondition *capi.Condition) error {
	// Rule: Removing condition is not allowed.
	// Why: We need it :)
	if oldCondition != nil && newCondition == nil {
		errorMessageFormat := "Removing %s condition is not allowed."
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessageFormat, oldCondition.Type)
	}

	return nil
}
