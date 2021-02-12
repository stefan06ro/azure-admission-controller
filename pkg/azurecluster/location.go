package azurecluster

import (
	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateLocation(azureCluster capzv1alpha3.AzureCluster, expectedLocation string) error {
	if azureCluster.Spec.Location != expectedLocation {
		return microerror.Maskf(unexpectedLocationError, "AzureCluster.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}
