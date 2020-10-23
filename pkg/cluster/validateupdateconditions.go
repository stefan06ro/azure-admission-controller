package cluster

import (
	"context"

	"k8s.io/api/admission/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func (a *UpdateValidator) validateConditions(ctx context.Context, request *v1beta1.AdmissionRequest, oldClusterCR *capi.Cluster, newClusterCR *capi.Cluster) (bool, error) {
	return false, nil
}
