package cluster

import (
	"context"

	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *CreateMutator) getDefaultStatusPatch(ctx context.Context) (*mutator.PatchOperation, error) {
	// Get default Cluster conditions
	conditions := m.getDefaultConditions(ctx)

	// Create default Cluster status
	clusterStatus := capiv1alpha3.ClusterStatus{
		Conditions: conditions,
	}

	patch := mutator.PatchAdd("/status", clusterStatus)
	return patch, nil
}
