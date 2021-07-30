package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	azureClusterCR, err := key.ToAzureClusterPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	azureClusterCROriginal := azureClusterCR.DeepCopy()

	patch, err := h.ensureControlPlaneEndpointHost(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = h.ensureControlPlaneEndpointPort(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = h.ensureLocation(ctx, azureClusterCR)
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

	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, h.ctrlReader, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureClusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFromObjectDiff(azureClusterCROriginal, azureClusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/networkSpec/vnet", capiPatches)
		capiPatches = patches.SkipForPath("/spec/networkSpec/subnets", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (h *WebhookHandler) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, h.baseDomain)), nil
	}

	return nil, nil
}

func (h *WebhookHandler) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}

func (h *WebhookHandler) ensureLocation(ctx context.Context, azureCluster *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if azureCluster.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", h.location), nil
	}

	return nil, nil
}
