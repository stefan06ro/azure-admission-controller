package azurecluster

import (
	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func validateLocation(azureCluster capzv1alpha3.AzureCluster, expectedLocation string) error {
	if azureCluster.Spec.Location != expectedLocation {
		return microerror.Maskf(errors.InvalidOperationError, "AzureCluster.Spec.Location can only be set to %s", expectedLocation)
	}

	return nil
}

func validateLocationUnchanged(oldAzureCluster capzv1alpha3.AzureCluster, newAzureCluster capzv1alpha3.AzureCluster) error {
	if oldAzureCluster.Spec.Location != newAzureCluster.Spec.Location {
		return microerror.Maskf(errors.InvalidOperationError, "AzureCluster.Spec.Location can't be changed")
	}

	return nil
}
