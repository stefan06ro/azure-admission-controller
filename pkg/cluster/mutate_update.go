package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (h *WebhookHandler) OnUpdateMutate(ctx context.Context, _ interface{}, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	clusterCR, err := key.ToClusterPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	clusterCROriginal := clusterCR.DeepCopy()

	patch, err := mutator.EnsureComponentVersionLabelFromRelease(ctx, h.ctrlReader, clusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, h.ctrlReader, clusterCR.GetObjectMeta(), "cluster-operator", label.ClusterOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	clusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFromObjectDiff(clusterCROriginal, clusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		result = append(result, capiPatches...)
	}

	return result, nil
}
