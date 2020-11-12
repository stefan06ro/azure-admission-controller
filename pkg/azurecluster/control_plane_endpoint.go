package azurecluster

import (
	"reflect"

	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func validateControlPlaneEndpoint(azureCluster capzv1alpha3.AzureCluster, baseDomain string) error {
	host := key.GetControlPlaneEndpointHost(azureCluster.Name, baseDomain)
	if azureCluster.Spec.ControlPlaneEndpoint.Host != host {
		return microerror.Maskf(invalidControlPlaneEndpointHostError, "ControlPlaneEndpoint.Host can only be set to %s", host)
	}

	if azureCluster.Spec.ControlPlaneEndpoint.Port != key.ControlPlaneEndpointPort {
		return microerror.Maskf(invalidControlPlaneEndpointPortError, "ControlPlaneEndpoint.Port can only be set to %d", key.ControlPlaneEndpointPort)
	}

	return nil
}

func validateControlPlaneEndpointUnchanged(old capzv1alpha3.AzureCluster, new capzv1alpha3.AzureCluster) error {
	if !reflect.DeepEqual(old.Spec.ControlPlaneEndpoint, new.Spec.ControlPlaneEndpoint) {
		return microerror.Maskf(controlPlaneEndpointWasChangedError, "ControlPlaneEndpoint can't be changed.")
	}

	return nil
}
