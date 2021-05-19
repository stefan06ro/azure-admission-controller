package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

func checkSSHKeyIsEmpty(ctx context.Context, mp *capzexp.AzureMachinePool) error {
	if mp.Spec.Template.SSHPublicKey != "" {
		return microerror.Maskf(sshFieldIsSetError, "AzureMachinePool.Spec.Template.SSHPublicKey is unsupported and must be empty.")
	}

	return nil
}
