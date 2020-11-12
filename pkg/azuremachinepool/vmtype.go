package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

const (
	minMemory = 16
	minCPUs   = 4
)

func checkInstanceTypeIsValid(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMachinePool *expcapzv1alpha3.AzureMachinePool) error {
	memory, err := vmcaps.Memory(ctx, azureMachinePool.Spec.Location, azureMachinePool.Spec.Template.VMSize)
	if err != nil {
		return microerror.Mask(err)
	}

	cpu, err := vmcaps.CPUs(ctx, azureMachinePool.Spec.Location, azureMachinePool.Spec.Template.VMSize)
	if err != nil {
		return microerror.Mask(err)
	}

	if memory < minMemory {
		return microerror.Maskf(insufficientMemoryError, "Memory has to be greater than %d GBs", minMemory)
	}

	if cpu < minCPUs {
		return microerror.Maskf(insufficientCPUError, "Number of cores has to be greater than %d", minCPUs)
	}

	return nil
}
