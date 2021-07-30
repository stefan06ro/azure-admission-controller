package machinepool

import (
	"context"
	"sort"

	"github.com/giantswarm/microerror"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	machinePoolNewCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	if !machinePoolNewCR.GetDeletionTimestamp().IsZero() {
		h.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
	}

	machinePoolOldCR, err := key.ToMachinePoolPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	err = machinePoolNewCR.ValidateUpdate(machinePoolOldCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(machinePoolOldCR, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = checkAvailabilityZonesUnchanged(ctx, machinePoolOldCR, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func checkAvailabilityZonesUnchanged(_ context.Context, oldMP *capiexp.MachinePool, newMP *capiexp.MachinePool) error {
	if len(oldMP.Spec.FailureDomains) != len(newMP.Spec.FailureDomains) {
		return microerror.Maskf(failureDomainWasChangedError, "Changing FailureDomains (availability zones) is not allowed.")
	}

	sort.Strings(oldMP.Spec.FailureDomains)
	sort.Strings(newMP.Spec.FailureDomains)

	for i := 0; i < len(oldMP.Spec.FailureDomains); i++ {
		if oldMP.Spec.FailureDomains[i] != newMP.Spec.FailureDomains[i] {
			return microerror.Maskf(failureDomainWasChangedError, "Changing FailureDomains (availability zones) is not allowed.")
		}
	}

	return nil
}
