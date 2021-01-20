package cluster

import (
	"encoding/json"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(cluster *capiv1alpha3.Cluster) *capiv1alpha3.Cluster

func Name(name string) BuilderOption {
	return func(cluster *capiv1alpha3.Cluster) *capiv1alpha3.Cluster {
		cluster.ObjectMeta.Name = name
		cluster.Labels[capiv1alpha3.ClusterLabelName] = name
		cluster.Labels[label.Cluster] = name
		return cluster
	}
}

func Labels(labels map[string]string) BuilderOption {
	return func(cluster *capiv1alpha3.Cluster) *capiv1alpha3.Cluster {
		for k, v := range labels {
			cluster.Labels[k] = v
		}
		return cluster
	}
}

func BuildCluster(opts ...BuilderOption) *capiv1alpha3.Cluster {
	clusterName := test.GenerateName()
	cluster := &capiv1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: capiv1alpha3.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion:    "5.0.0",
				capiv1alpha3.ClusterLabelName: clusterName,
				label.Cluster:                 clusterName,
				label.Organization:            "giantswarm",
				label.ReleaseVersion:          "13.0.0-alpha4",
			},
		},
		Spec: capiv1alpha3.ClusterSpec{
			ClusterNetwork: &capiv1alpha3.ClusterNetwork{
				APIServerPort: to.Int32Ptr(443),
				ServiceDomain: "cluster.local",
				Services: &capiv1alpha3.NetworkRanges{
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
