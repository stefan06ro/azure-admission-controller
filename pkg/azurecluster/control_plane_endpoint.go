package azurecluster

import (
	"reflect"

	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func validateControlPlaneEndpoint(azureCluster capz.AzureCluster, baseDomain string) error {
	host := key.GetControlPlaneEndpointHost(azureCluster.Name, baseDomain)
	if azureCluster.Spec.ControlPlaneEndpoint.Host != host {
		return microerror.Maskf(invalidControlPlaneEndpointHostError, "ControlPlaneEndpoint.Host can only be set to %s", host)
	}

	if azureCluster.Spec.ControlPlaneEndpoint.Port != key.ControlPlaneEndpointPort {
		return microerror.Maskf(invalidControlPlaneEndpointPortError, "ControlPlaneEndpoint.Port can only be set to %d", key.ControlPlaneEndpointPort)
	}

	return nil
}

func validateControlPlaneEndpointUnchanged(old capz.AzureCluster, new capz.AzureCluster) error {
	if !reflect.DeepEqual(old.Spec.ControlPlaneEndpoint, new.Spec.ControlPlaneEndpoint) {
		return microerror.Maskf(controlPlaneEndpointWasChangedError, "ControlPlaneEndpoint can't be changed.")
	}

	return nil
}
