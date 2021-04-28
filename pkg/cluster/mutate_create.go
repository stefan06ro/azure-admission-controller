package cluster

import (
	"context"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	clusterCR, err := key.ToClusterPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	clusterCROriginal := clusterCR.DeepCopy()

	patch, err := m.ensureClusterNetwork(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointHost(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, clusterCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	clusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = mutator.GenerateFromObjectDiff(clusterCROriginal, clusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *Mutator) ensureClusterNetwork(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	// Ensure ClusterNetwork is set.
	if clusterCR.Spec.ClusterNetwork == nil {
		clusterNetwork := capi.ClusterNetwork{
			APIServerPort: to.Int32Ptr(key.ControlPlaneEndpointPort),
			ServiceDomain: key.ServiceDomain(),
			Services: &capi.NetworkRanges{
				CIDRBlocks: []string{
					key.ClusterNetworkServiceCIDR,
				},
			},
		}

		return mutator.PatchAdd("/spec/clusterNetwork", clusterNetwork), nil
	}

	return nil, nil
}

func (m *Mutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *Mutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}
