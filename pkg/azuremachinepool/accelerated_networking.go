package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func checkAcceleratedNetworking(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMachinePool *expcapzv1alpha3.AzureMachinePool) error {
	// Accelerated networking is disabled (false) or in auto-detect mode (nil). This is always allowed.
	if azureMachinePool.Spec.Template.AcceleratedNetworking == nil || !*azureMachinePool.Spec.Template.AcceleratedNetworking {
		return nil
	}

	isSupported, err := isAcceleratedNetworkingSupportedOnVmSize(ctx, vmcaps, azureMachinePool)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isSupported {
		return microerror.Maskf(vmsizeDoesNotSupportAcceleratedNetworkingError, "The new VMSize does not support AcceleratedNetworking")
	}

	return nil
}

func isAcceleratedNetworkingSupportedOnVmSize(ctx context.Context, vmcaps *vmcapabilities.VMSKU, azureMachinePool *expcapzv1alpha3.AzureMachinePool) (bool, error) {
	acceleratedNetworkingAvailable, err := vmcaps.HasCapability(ctx, azureMachinePool.Spec.Location, azureMachinePool.Spec.Template.VMSize, vmcapabilities.CapabilityAcceleratedNetworking)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return acceleratedNetworkingAvailable, nil
}

func hasAcceleratedNetworkingPropertyChanged(ctx context.Context, old *expcapzv1alpha3.AzureMachinePool, new *expcapzv1alpha3.AzureMachinePool) bool {
	if old.Spec.Template.AcceleratedNetworking != nil {
		if new.Spec.Template.AcceleratedNetworking != nil {
			return *old.Spec.Template.AcceleratedNetworking != *new.Spec.Template.AcceleratedNetworking
		} else {
			return true
		}
	}

	// Old AcceleratedNetworking is nil
	if new.Spec.Template.AcceleratedNetworking != nil {
		return true
	}

	// Both are nil
	return false
}
