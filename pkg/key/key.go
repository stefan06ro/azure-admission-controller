package key

import (
	"fmt"
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
