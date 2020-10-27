package generic

import (
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GenericObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func newObjectWithOrganization(org *string) metav1.Object {
	obj := &GenericObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Unknown",
			APIVersion: "unknown.generic.example/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"giantswarm.io/cluster":                "ab123",
				"cluster.x-k8s.io/cluster-name":        "ab123",
				"cluster.x-k8s.io/control-plane":       "true",
				"giantswarm.io/machine-pool":           "ab123",
				"release.giantswarm.io/version":        "13.0.0",
			},
		},
	}

	if org != nil {
		obj.Labels[label.Organization] = *org
	}

	return obj
}
