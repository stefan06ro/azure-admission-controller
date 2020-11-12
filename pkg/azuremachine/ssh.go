package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func checkSSHKeyIsEmpty(ctx context.Context, mp *capzv1alpha3.AzureMachine) error {
	if mp.Spec.SSHPublicKey != "" {
		return microerror.Maskf(sshFieldIsSetError, "AzureMachine.Spec.SSHPublicKey is unsupported and must be empty.")
	}

	return nil
}
