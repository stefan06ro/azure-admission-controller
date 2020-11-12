package generic

import (
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

type GenericObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func newObjectWithOrganization(org *string) metav1.Object {
	obj := &GenericObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				label.AzureOperatorVersion:                "5.0.0",
				label.Cluster:                             "ab123",
				capiv1alpha3.ClusterLabelName:             "ab123",
				capiv1alpha3.MachineControlPlaneLabelName: "true",
				label.MachinePool:                         "ab123",
				label.ReleaseVersion:                      "13.0.0",
			},
		},
	}

	if org != nil {
		obj.Labels[label.Organization] = *org
	}

	return obj
}
