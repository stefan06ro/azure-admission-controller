package conditions

import (
	aeconditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func ValidateClusterConditions(oldClusterCR *capi.Cluster, newClusterCR *capi.Cluster) error {
	var err error

	err = ValidateCreatingCondition(oldClusterCR, newClusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ValidateUpgradingCondition(oldClusterCR, newClusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func ValidateCreatingCondition(oldClusterCR *capi.Cluster, newClusterCR *capi.Cluster) error {
	var err error
	oldCreatingCondition := capiconditions.Get(oldClusterCR, aeconditions.CreatingCondition)
	newCreatingCondition := capiconditions.Get(newClusterCR, aeconditions.CreatingCondition)

	// See functions for details about validations.
	validations := []conditionChangeValidation{
		validateNewStatusUnknownIsNotAllowed,
		validateNewStatusMustBeEitherTrueOrFalse,
		validateRemovingConditionIsNotAllowed,
	}

	err = validateAll(oldCreatingCondition, newCreatingCondition, validations)
	if err != nil {
		return microerror.Mask(err)
	}

	creationWasCompleted := oldCreatingCondition != nil && oldCreatingCondition.Status == corev1.ConditionFalse
	nowIsInCreation := newCreatingCondition != nil && newCreatingCondition.Status == corev1.ConditionTrue

	// Rule: Creating cannot be changed from False to True.
	// Why: This change would mean that the cluster creation has started again,
	//      which is not possible.
	if creationWasCompleted && nowIsInCreation {
		const errorMessage = "Modifying Creating condition from False to True is not allowed, " +
			"as that would mean that the cluster creation is started again, which is not possible"
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessage)
	}

	return nil
}

func ValidateUpgradingCondition(oldClusterCR *capi.Cluster, newClusterCR *capi.Cluster) error {
	var err error
	oldUpgradingCondition := capiconditions.Get(oldClusterCR, aeconditions.UpgradingCondition)
	newUpgradingCondition := capiconditions.Get(newClusterCR, aeconditions.UpgradingCondition)

	// See functions for details about validations.
	validations := []conditionChangeValidation{
		validateNewStatusUnknownIsNotAllowed,
		validateNewStatusMustBeEitherTrueOrFalse,
		validateRemovingConditionIsNotAllowed,
	}

	err = validateAll(oldUpgradingCondition, newUpgradingCondition, validations)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
