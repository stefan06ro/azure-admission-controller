package cluster

import (
	"encoding/json"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(cluster *capi.Cluster) *capi.Cluster

func Name(name string) BuilderOption {
	return func(cluster *capi.Cluster) *capi.Cluster {
		cluster.ObjectMeta.Name = name
		cluster.Labels[capi.ClusterLabelName] = name
		cluster.Labels[label.Cluster] = name
		return cluster
	}
}

func Labels(labels map[string]string) BuilderOption {
	return func(cluster *capi.Cluster) *capi.Cluster {
		for k, v := range labels {
			cluster.Labels[k] = v
		}
		return cluster
	}
}

func ControlPlaneEndpoint(controlPlaneEndpointHost string, controlPlaneEndpointPort int32) BuilderOption {
	return func(cluster *capi.Cluster) *capi.Cluster {
		cluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpointHost
		cluster.Spec.ControlPlaneEndpoint.Port = controlPlaneEndpointPort
		return cluster
	}
}

func WithDeletionTimestamp() BuilderOption {
	return func(cluster *capi.Cluster) *capi.Cluster {
		now := metav1.Now()
		cluster.ObjectMeta.SetDeletionTimestamp(&now)
		return cluster
	}
}

func BuildCluster(opts ...BuilderOption) *capi.Cluster {
	clusterName := test.GenerateName()
	cluster := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: capi.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion: "5.0.0",
				capi.ClusterLabelName:      clusterName,
				label.Cluster:              clusterName,
				label.Organization:         "giantswarm",
				label.ReleaseVersion:       "13.0.0-alpha4",
			},
		},
		Spec: capi.ClusterSpec{
			ClusterNetwork: &capi.ClusterNetwork{
				APIServerPort: to.Int32Ptr(443),
				ServiceDomain: "cluster.local",
				Services: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"172.31.0.0/16",
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(cluster)
	}

	return cluster
}

func BuildClusterAsJson(opts ...BuilderOption) []byte {
	cluster := BuildCluster(opts...)

	byt, _ := json.Marshal(cluster)

	return byt
}
