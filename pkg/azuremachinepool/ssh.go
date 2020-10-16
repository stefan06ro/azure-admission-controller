package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

func checkSSHKeyIsEmpty(ctx context.Context, mp *expcapzv1alpha3.AzureMachinePool) error {
	if mp.Spec.Template.SSHPublicKey != "" {
		return microerror.Maskf(invalidOperationError, "AzureMachinePool.Spec.Template.SSHPublicKey is unsupported and must be empty.")
	}

	return nil
}
