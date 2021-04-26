package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *Mutator) Mutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureClusterCR, err := key.ToAzureClusterPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureClusterCROriginal := azureClusterCR.DeepCopy()

	patch, err := m.ensureControlPlaneEndpointHost(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureLocation(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = ensureAPIServerLB(azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureClusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = mutator.GenerateFromObjectDiff(azureClusterCROriginal, azureClusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = mutator.SkipForPath("/spec/networkSpec/vnet", capiPatches)
		capiPatches = mutator.SkipForPath("/spec/networkSpec/subnets", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *Mutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *Mutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}

func (m *Mutator) ensureLocation(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if azureCluster.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", m.location), nil
	}

	return nil, nil
}
