package key

import "fmt"

const (
	ControlPlaneEndpointPort  = 443
	ClusterNetworkServiceCIDR = "172.31.0.0/16"
)

func GetControlPlaneEndpointHost(clusterName string, baseDomain string) string {
	return fmt.Sprintf("api.%s.%s", clusterName, baseDomain)
}

func ServiceDomain(clusterName string, baseDomain string) string {
	return fmt.Sprintf("%s.%s", clusterName, baseDomain)
}
