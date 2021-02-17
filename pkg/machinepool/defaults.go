package machinepool

import (
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

const defaultReplicas int64 = 1

// setDefaultSpecValues checks if some optional field is not set, and sets
// default values defined by upstream Cluster API.
func setDefaultSpecValues(m mutator.Mutator, machinePool *capiexp.MachinePool) []mutator.PatchOperation {
	var patches []mutator.PatchOperation

	return patches
}
