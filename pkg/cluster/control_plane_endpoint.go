package cluster

import (
	"reflect"

	"github.com/giantswarm/microerror"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func validateClusterNetwork(cluster capiv1alpha3.Cluster) error {
	if cluster.Spec.ClusterNetwork == nil {
		return microerror.Maskf(emptyClusterNetworkError, "ClusterNetwork can't be null")
	}

	if *cluster.Spec.ClusterNetwork.APIServerPort != key.ControlPlaneEndpointPort {
		return microerror.Maskf(unexpectedAPIServerPortError, "ClusterNetwork.APIServerPort can only be set to %d", key.ControlPlaneEndpointPort)
	}

	if cluster.Spec.ClusterNetwork.ServiceDomain != key.ServiceDomain() {
		return microerror.Maskf(unexpectedServiceDomainError, "ClusterNetwork.ServiceDomain can only be set to %s", key.ServiceDomain())
	}

	if cluster.Spec.ClusterNetwork.Services == nil {
		return microerror.Maskf(emptyClusterNetworkServicesError, "ClusterNetwork.Services can't be null")
	}

	if !reflect.DeepEqual(cluster.Spec.ClusterNetwork.Services.CIDRBlocks, []string{key.ClusterNetworkServiceCIDR}) {
		return microerror.Maskf(unexpectedCIDRBlocksError, "ClusterNetwork.Services.CIDRBlocks can only be set to [%s]", key.ClusterNetworkServiceCIDR)
	}

	return nil
}

func validateControlPlaneEndpoint(cluster capiv1alpha3.Cluster, baseDomain string) error {
	host := key.GetControlPlaneEndpointHost(cluster.Name, baseDomain)
	if cluster.Spec.ControlPlaneEndpoint.Host != host {
		return microerror.Maskf(invalidControlPlaneEndpointHostError, "ControlPlaneEndpoint.Host can only be set to %s", host)
	}

	if cluster.Spec.ControlPlaneEndpoint.Port != key.ControlPlaneEndpointPort {
		return microerror.Maskf(invalidControlPlaneEndpointPortError, "ControlPlaneEndpoint.Port can only be set to %d", key.ControlPlaneEndpointPort)
	}

	return nil
}

func validateControlPlaneEndpointUnchanged(old capiv1alpha3.Cluster, new capiv1alpha3.Cluster) error {
	if !reflect.DeepEqual(old.Spec.ControlPlaneEndpoint, new.Spec.ControlPlaneEndpoint) {
		return microerror.Maskf(controlPlaneEndpointWasChangedError, "ControlPlaneEndpoint can't be changed.")
	}

	return nil
}
