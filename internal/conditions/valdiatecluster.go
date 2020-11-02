package conditions

import (
	"github.com/blang/semver"
	aeconditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
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
	creationWasCompleted := capiconditions.IsFalse(oldClusterCR, aeconditions.CreatingCondition)
	nowIsInCreation := capiconditions.IsTrue(newClusterCR, aeconditions.CreatingCondition)

	if creationWasCompleted && nowIsInCreation {
		const errorMessage = "Modifying Creating condition from False to True is not allowed, " +
			"as that would mean that the cluster creation is started again, which is not possible"
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessage)
	}

	return nil
}

func ValidateUpgradingCondition(oldClusterCR *capi.Cluster, newClusterCR *capi.Cluster) error {
	newUpgradingCondition := capiconditions.Get(newClusterCR, aeconditions.UpgradingCondition)

	// Setting Unknown is not allowed, we should be able to tell if the cluster
	// is being upgraded or not.
	if newUpgradingCondition != nil && newUpgradingCondition.Status == corev1.ConditionUnknown {
		errorMessage := "Setting Unknown status to Upgrading conditions is not allowed"
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessage)
	}

	// We only allow True and False as condition status values. Any other new
	// or custom condition status values are not allowed.
	conditionStatusNotAllowed := newUpgradingCondition != nil &&
		newUpgradingCondition.Status != corev1.ConditionTrue &&
		newUpgradingCondition.Status != corev1.ConditionFalse

	if conditionStatusNotAllowed {
		errorMessageFormat := "Allowed values for Upgrading condition status are True and False, got %s"
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessageFormat, newUpgradingCondition.Status)
	}

	oldUpgradingCondition := capiconditions.Get(oldClusterCR, aeconditions.UpgradingCondition)

	// Removing Upgrading condition is not allowed.
	if oldUpgradingCondition != nil && newUpgradingCondition == nil {
		errorMessage := "Removing Upgrading condition is not allowed"
		return microerror.Maskf(errors.InvalidConditionModificationError, errorMessage)
	}

	// When we are setting Upgrading condition to False because the upgrade has
	// not been started, we expect that the current release version is also
	// set.
	// When we are setting Upgrading condition to False because the upgrade has
	// been completed, we expect that the new release version is also set.
	if newUpgradingCondition != nil &&
		newUpgradingCondition.Status == corev1.ConditionFalse &&
		(newUpgradingCondition.Reason == aeconditions.UpgradeNotStartedReason ||
			newUpgradingCondition.Reason == aeconditions.UpgradeCompletedReason) {

		message, err := aeconditions.DeserializeUpgradingConditionMessage(newUpgradingCondition.Message)
		if err != nil {
			errorMessage := "Could not parse error message, expected serialized JSON for type %T"
			return microerror.Maskf(errors.InvalidUpgradingConditionMessageFormatError, errorMessage, message)
		}
		lastSavedReleaseVersion, err := semver.New(message.ReleaseVersion)
		if err != nil {
			return microerror.Mask(err)
		}

		releaseVersionString := newClusterCR.GetLabels()[label.ReleaseVersion]
		desiredReleaseVersion, err := semver.New(releaseVersionString)
		if err != nil {
			return microerror.Mask(err)
		}

		// The release version that is saved in the condition represents the
		// release version with which the cluster was created, or release
		// version to which the cluster was successfully upgraded the last
		// time. Therefore, the release version set in label
		// 'release.giantswarm.io/version', which represents the desired
		// cluster version, must be either (1) same as the release version that
		// is saved in condition, which is the state after cluster creation or
		// cluster upgrade, or (2) greater than the release version that is
		// saved in condition, which is state when the upgrade has been
		// initiated.
		if !desiredReleaseVersion.GTE(*lastSavedReleaseVersion) {
			errorMessage := "Release version that is set in the CR's" +
				"'release.giantswarm.io/version' label must be greater " +
				"than or equal to the release version that is set in the " +
				"Upgrading condition message, expected %s>=%s"
			return microerror.Maskf(
				errors.InvalidReleaseVersionInUpgradingConditionMessageError,
				errorMessage,
				releaseVersionString, message.ReleaseVersion)
		}
	}

	return nil
}
