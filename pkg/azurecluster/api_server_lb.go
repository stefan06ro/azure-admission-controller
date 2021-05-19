package azurecluster

import (
	"reflect"

	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func ensureAPIServerLB(cr *capz.AzureCluster) (*mutator.PatchOperation, error) {
	apiServerLB := capz.LoadBalancerSpec{
		Name: key.APIServerLBName(cr.Name),
		SKU:  capz.SKU(key.APIServerLBSKU()),
		Type: capz.LBType(key.APIServerLBType()),
		FrontendIPs: []capz.FrontendIP{
			{Name: key.APIServerLBFrontendIPName(cr.Name)},
		},
	}

	if reflect.DeepEqual(apiServerLB, cr.Spec.NetworkSpec.APIServerLB) {
		// No need to make a patch, this is already the right value.
		return nil, nil
	}

	return mutator.PatchAdd("/spec/networkSpec/apiServerLB", apiServerLB), nil
}
