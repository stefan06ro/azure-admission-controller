package azurecluster

import (
	"reflect"

	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func ensureAPIServerLB(cr *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	apiServerLB := capzv1alpha3.LoadBalancerSpec{
		Name: key.APIServerLBName(cr.Name),
		SKU:  capzv1alpha3.SKU(key.APIServerLBSKU()),
		Type: capzv1alpha3.LBType(key.APIServerLBType()),
		FrontendIPs: []capzv1alpha3.FrontendIP{
			{Name: key.APIServerLBFrontendIPName(cr.Name)},
		},
	}

	if reflect.DeepEqual(apiServerLB, cr.Spec.NetworkSpec.APIServerLB) {
		// No need to make a patch, this is already the right value.
		return nil, nil
	}

	return mutator.PatchAdd("/spec/networkSpec/apiServerLB", apiServerLB), nil
}
