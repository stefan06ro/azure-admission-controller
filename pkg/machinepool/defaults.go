package machinepool

import (
	"fmt"

	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

const defaultReplicas int64 = 1

// setDefaultSpecValues checks if some optional field is not set, and sets
// default values defined by upstream Cluster API.
func setDefaultSpecValues(m mutator.Mutator, machinePool *capiexp.MachinePool) []mutator.PatchOperation {
	var patches []mutator.PatchOperation

	defaultSpecReplicas := setDefaultReplicaValue(m, machinePool)
	if defaultSpecReplicas != nil {
		patches = append(patches, *defaultSpecReplicas)
	}

	return patches
}

// setDefaultReplicaValue checks if Spec.Replicas has been set, and if it is
// not, it sets its value to 1.
func setDefaultReplicaValue(m mutator.Mutator, machinePool *capiexp.MachinePool) *mutator.PatchOperation {
	if machinePool.Spec.Replicas == nil {
		m.Log("level", "debug", "message", fmt.Sprintf("setting default MachinePool.Spec.Replicas to %d", defaultReplicas))
		return mutator.PatchAdd("/spec/replicas", defaultReplicas)
	}

	return nil
}
