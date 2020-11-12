package azuremachinepool

import (
	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

func checkLocation(azureMachinePool expcapzv1alpha3.AzureMachinePool, expectedLocation string) error {
	if azureMachinePool.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureMachinePool.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}

func checkLocationUnchanged(old expcapzv1alpha3.AzureMachinePool, new expcapzv1alpha3.AzureMachinePool) error {
	if old.Spec.Location != new.Spec.Location {
		return microerror.Maskf(locationWasChangedError, "AzureMachinePool.Spec.Location can't be changed")
	}

	return nil
}
