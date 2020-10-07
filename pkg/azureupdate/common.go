package azureupdate

import (
	corev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	providerv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func clusterIsUpgrading(cr *v1alpha1.AzureConfig) (bool, string) {
	for _, cond := range cr.Status.Cluster.Conditions {
		if cond.Type == conditionUpdating {
			return true, conditionUpdating
		}
		if cond.Type == conditionCreating {
			return true, conditionCreating
		}
	}

	return false, ""
}

func getFakeCtrlClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	err := v1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = corev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = expcapiv1alpha3.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = expcapzv1alpha3.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capiv1alpha3.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capzv1alpha3.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = providerv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = releasev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	return fake.NewFakeClientWithScheme(scheme), nil
}
