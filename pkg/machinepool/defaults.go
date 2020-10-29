package machinepool

import (
	"context"
	"fmt"

	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

// setDefaultSpecValues checks if some optional field is not set, and sets
// default values defined by upstream Cluster API.
func (m *CreateMutator) setDefaultSpecValues(ctx context.Context, machinePool *capiexp.MachinePool) []mutator.PatchOperation {
	var patches []mutator.PatchOperation

	defaultSpecReplicas := m.setDefaultReplicaValue(ctx, machinePool)
	if defaultSpecReplicas != nil {
		patches = append(patches, *defaultSpecReplicas)
	}

	return patches
}

// setDefaultReplicaValue checks if Spec.Replicas has been set, and if it is
// not, it sets its value to 1.
func (m *CreateMutator) setDefaultReplicaValue(ctx context.Context, machinePool *capiexp.MachinePool) *mutator.PatchOperation {
	if machinePool.Spec.Replicas == nil {
		const defaultReplicas = "1"
		m.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting default MachinePool.Spec.Replicas to %s", defaultReplicas))
		return mutator.PatchAdd("/spec/replicas", defaultReplicas)
	}

	return nil
}
