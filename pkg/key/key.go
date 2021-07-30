package key

import (
	"fmt"

	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

const (
	ControlPlaneEndpointPort  = 443
	ClusterNetworkServiceCIDR = "172.31.0.0/16"
)

func GetControlPlaneEndpointHost(clusterName string, baseDomain string) string {
	return fmt.Sprintf("api.%s.%s", clusterName, baseDomain)
}

func ServiceDomain() string {
	return "cluster.local"
}

func APIServerLBName(clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, "API", "PublicLoadBalancer")
}

func APIServerLBSKU() string {
	return "Standard"
}

func APIServerLBType() string {
	return "Public"
}

func APIServerLBFrontendIPName(clusterName string) string {
	return fmt.Sprintf("%s-%s", APIServerLBName(clusterName), "Frontend")
}

func OSDiskCachingType() string {
	return "ReadWrite"
}

func MasterSubnetName(clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, "VirtualNetwork", "MasterSubnet")
}

func ToClusterPtr(v interface{}) (*capi.Cluster, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	customObjectPointer, ok := v.(*capi.Cluster)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	return customObjectPointer, nil
}
