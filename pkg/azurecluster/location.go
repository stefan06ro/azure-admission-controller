package azurecluster

import (
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateLocation(azureCluster capz.AzureCluster, expectedLocation string) error {
	if azureCluster.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureCluster.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}
