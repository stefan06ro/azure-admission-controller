package machinepool

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/giantswarm/microerror"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

const (
	// TODO move annotation to shared library for AWS to use.
	annotationAutoFailureDomain = "giantswarm.io/auto-availability-zones-count"
)

func (m *CreateMutator) ensureAutomaticFailureDomains(ctx context.Context, mp *v1alpha3.MachinePool) ([]mutator.PatchOperation, error) {
	if numAzsStr, ok := mp.GetAnnotations()[annotationAutoFailureDomain]; ok {
		// Try to parse the number of requested availability zones.
		numAzs, err := strconv.Atoi(numAzsStr)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		if numAzs >= 1 {
			if mp.Spec.Template.Spec.InfrastructureRef.Namespace == "" || mp.Spec.Template.Spec.InfrastructureRef.Name == "" {
				return []mutator.PatchOperation{}, microerror.Maskf(azureMachinePoolNotFoundError, "MachinePool's InfrastructureRef has to be set")
			}

			amp := expcapzv1alpha3.AzureMachinePool{}
			err := m.ctrlClient.Get(ctx, client.ObjectKey{Namespace: mp.Spec.Template.Spec.InfrastructureRef.Namespace, Name: mp.Spec.Template.Spec.InfrastructureRef.Name}, &amp)
			if err != nil {
				return []mutator.PatchOperation{}, microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool")
			}

			supportedAZs, err := m.vmcaps.SupportedAZs(ctx, amp.Spec.Location, amp.Spec.Template.VMSize)
			if err != nil {
				return []mutator.PatchOperation{}, microerror.Mask(err)
			}

			if numAzs > len(supportedAZs) {
				// User requested too many AZs.
				return []mutator.PatchOperation{}, microerror.Maskf(tooManyFailureDomainsRequestedError, "Location %s has %d failure domains, but %d were requested", amp.Spec.Location, len(supportedAZs), numAzs)
			}

			// Request is legit, select numAzs random failure domains from the list of supported.
			rand.Seed(time.Now().Unix())
			var azs []string
			for len(azs) < numAzs {
				r := rand.Intn(len(supportedAZs))
				azs = append(azs, supportedAZs[r])
				supportedAZs = append(supportedAZs[:r], supportedAZs[r+1:]...)
			}

			// Sort Azs.
			sort.Strings(azs)

			return []mutator.PatchOperation{
				// Set failure domains
				*mutator.PatchAdd("/spec/failureDomains", azs),
				// Remove annotation
				*mutator.PatchRemove(fmt.Sprintf("/metadata/annotations/%s", escapeJSONPatchString(annotationAutoFailureDomain))),
			}, nil
		}
	}

	return []mutator.PatchOperation{}, nil
}
