package azuremachinepool

import (
	"context"
	"reflect"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

var desiredDataDisks = []capzv1alpha3.DataDisk{
	{
		NameSuffix: "docker",
		DiskSizeGB: 100,
		Lun:        to.Int32Ptr(21),
	},
	{
		NameSuffix: "kubelet",
		DiskSizeGB: 100,
		Lun:        to.Int32Ptr(22),
	},
}

func checkDataDisks(ctx context.Context, mp *expcapzv1alpha3.AzureMachinePool) error {
	if !reflect.DeepEqual(mp.Spec.Template.DataDisks, desiredDataDisks) {
		return microerror.Maskf(invalidOperationError, "AzureMachinePool.Spec.Template.DataDisks is reserved and should not be set.")
	}

	return nil
}
