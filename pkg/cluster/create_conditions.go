package cluster

import (
	"context"

	aeV3conditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *CreateMutator) ensureConditions(ctx context.Context, newCluster *capiv1alpha3.Cluster) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	patch, err := m.ensureCreatingCondition(ctx, newCluster)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}

func (m *CreateMutator) ensureCreatingCondition(ctx context.Context, newCluster *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	if capiconditions.IsTrue(newCluster, aeV3conditions.CreatingCondition) {
		return nil, nil
	}

	var patch *mutator.PatchOperation
	creatingCondition := capiconditions.TrueCondition(aeV3conditions.CreatingCondition)
	patch = mutator.PatchAdd("/status/conditions", []capiv1alpha3.Condition{*creatingCondition})
	m.Log(ctx, "level", "debug", "message", "set Creating condition")

	return patch, nil
}
