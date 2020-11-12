package azurecluster

import (
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster

func Name(name string) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		azureCluster.ObjectMeta.Name = name
		azureCluster.Labels[capiv1alpha3.ClusterLabelName] = name
		azureCluster.Labels[label.Cluster] = name
		azureCluster.Spec.ControlPlaneEndpoint.Host = fmt.Sprintf("api.%s.k8s.test.westeurope.azure.gigantic.io", name)
		return azureCluster
	}
}

func Labels(labels map[string]string) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		for k, v := range labels {
			azureCluster.Labels[k] = v
		}
		return azureCluster
	}
}

func Location(location string) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		azureCluster.Spec.Location = location
		return azureCluster
	}
}

func ControlPlaneEndpoint(controlPlaneEndpointHost string, controlPlaneEndpointPort int32) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		azureCluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpointHost
		azureCluster.Spec.ControlPlaneEndpoint.Port = controlPlaneEndpointPort
		return azureCluster
	}
}

func BuildAzureCluster(opts ...BuilderOption) *capzv1alpha3.AzureCluster {
	clusterName := test.GenerateName()
	azureCluster := &capzv1alpha3.AzureCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureCluster",
			APIVersion: capzv1alpha3.GroupVersion.String(),
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
		Spec: capzv1alpha3.AzureClusterSpec{
			ResourceGroup: clusterName,
			Location:      "westeurope",
			ControlPlaneEndpoint: capiv1alpha3.APIEndpoint{
				Host: fmt.Sprintf("api.%s.k8s.test.westeurope.azure.gigantic.io", clusterName),
				Port: 443,
			},
		},
	}

	for _, opt := range opts {
		opt(azureCluster)
	}

	return azureCluster
}

func BuildAzureClusterAsJson(opts ...BuilderOption) []byte {
	azureCluster := BuildAzureCluster(opts...)

	byt, _ := json.Marshal(azureCluster)

	return byt
}
