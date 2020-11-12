package azuremachine

import (
	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateLocation(azureMachine capzv1alpha3.AzureMachine, expectedLocation string) error {
	if azureMachine.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureMachine.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}

func validateLocationUnchanged(old capzv1alpha3.AzureMachine, new capzv1alpha3.AzureMachine) error {
	if old.Spec.Location != new.Spec.Location {
		return microerror.Maskf(locationWasChangedError, "AzureMachine.Spec.Location can't be changed")
	}

	return nil
}
