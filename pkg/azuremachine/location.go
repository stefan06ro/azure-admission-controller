package azuremachine

import (
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateLocation(azureMachine capz.AzureMachine, expectedLocation string) error {
	if azureMachine.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureMachine.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}

func validateLocationUnchanged(old capz.AzureMachine, new capz.AzureMachine) error {
	if old.Spec.Location != new.Spec.Location {
		return microerror.Maskf(locationWasChangedError, "AzureMachine.Spec.Location can't be changed")
	}

	return nil
}
