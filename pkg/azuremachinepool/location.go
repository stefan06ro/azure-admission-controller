package azuremachinepool

import (
	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

func checkLocation(azureMachinePool capzexp.AzureMachinePool, expectedLocation string) error {
	if azureMachinePool.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureMachinePool.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}

func checkLocationUnchanged(old capzexp.AzureMachinePool, new capzexp.AzureMachinePool) error {
	if old.Spec.Location != new.Spec.Location {
		return microerror.Maskf(locationWasChangedError, "AzureMachinePool.Spec.Location can't be changed")
	}

	return nil
}
