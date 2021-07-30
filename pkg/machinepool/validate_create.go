package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
	machinePoolNewCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = machinePoolNewCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelMatchesCluster(ctx, h.ctrlClient, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.checkAvailabilityZones(ctx, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *WebhookHandler) checkAvailabilityZones(ctx context.Context, mp *capiexp.MachinePool) error {
	// Get the AzureMachinePool CR related to this MachinePool (we need it to get the VM type).
	if mp.Spec.Template.Spec.InfrastructureRef.Namespace == "" || mp.Spec.Template.Spec.InfrastructureRef.Name == "" {
		return microerror.Maskf(azureMachinePoolNotFoundError, "MachinePool's InfrastructureRef has to be set")
	}
	amp := capzexp.AzureMachinePool{}
	err := h.ctrlClient.Get(ctx, client.ObjectKey{Namespace: mp.Spec.Template.Spec.InfrastructureRef.Namespace, Name: mp.Spec.Template.Spec.InfrastructureRef.Name}, &amp)
	if err != nil {
		return microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool")
	}

	supportedZones, err := h.vmcaps.SupportedAZs(ctx, amp.Spec.Location, amp.Spec.Template.VMSize)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, zone := range mp.Spec.FailureDomains {
		if !inSlice(zone, supportedZones) {
			// Found one unsupported availability zone requested.
			return microerror.Maskf(unsupportedFailureDomainError, "You requested the Machine Pool with type %s to be placed in the following FailureDomains (aka Availability zones): %v but the VM type only supports %v in %s", amp.Spec.Template.VMSize, mp.Spec.FailureDomains, supportedZones, amp.Spec.Location)
		}
	}

	return nil
}

func inSlice(needle string, haystack []string) bool {
	for _, supported := range haystack {
		if needle == supported {
			return true
		}
	}
	return false
}
