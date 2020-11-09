// +build k8srequired

package organization

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func TestFailWhenOrganizationDoesNotExist(t *testing.T) {
	ctx := context.Background()
	var err error

	cluster := &capiv1alpha3.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.Organization: "non-existing",
			},
		},
		Spec: capiv1alpha3.ClusterSpec{
			ClusterNetwork:       nil,
			ControlPlaneEndpoint: capiv1alpha3.APIEndpoint{},
			ControlPlaneRef:      nil,
			InfrastructureRef: &v1.ObjectReference{
				Kind:      "AzureCluster",
				Namespace: "org-giantswarm",
				Name:      "my-cluster",
			},
		},
	}
	err = appTest.CtrlClient().Create(ctx, cluster)
	if err == nil {
		t.Fatalf("it should fail when using a non existing organization")
	}
}
