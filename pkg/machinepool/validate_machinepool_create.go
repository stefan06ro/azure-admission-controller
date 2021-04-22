package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type Validator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type ValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
	VMcaps     *vmcapabilities.VMSKU
}

func NewValidator(config ValidatorConfig) (*Validator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	admitter := &Validator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return admitter, nil
}

func (a *Validator) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &capiexp.MachinePool{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse MachinePool CR: %v", err)
	}

	return cr, nil
}

func (a *Validator) Validate(ctx context.Context, object interface{}) error {
	machinePoolNewCR, err := key.ToMachinePoolPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = machinePoolNewCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelMatchesCluster(ctx, a.ctrlClient, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.checkAvailabilityZones(ctx, machinePoolNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func (a *Validator) checkAvailabilityZones(ctx context.Context, mp *capiexp.MachinePool) error {
	// Get the AzureMachinePool CR related to this MachinePool (we need it to get the VM type).
	if mp.Spec.Template.Spec.InfrastructureRef.Namespace == "" || mp.Spec.Template.Spec.InfrastructureRef.Name == "" {
		return microerror.Maskf(azureMachinePoolNotFoundError, "MachinePool's InfrastructureRef has to be set")
	}
	amp := capzexp.AzureMachinePool{}
	err := a.ctrlClient.Get(ctx, client.ObjectKey{Namespace: mp.Spec.Template.Spec.InfrastructureRef.Namespace, Name: mp.Spec.Template.Spec.InfrastructureRef.Name}, &amp)
	if err != nil {
		return microerror.Maskf(azureMachinePoolNotFoundError, "AzureMachinePool has to be created before the related MachinePool")
	}

	supportedZones, err := a.vmcaps.SupportedAZs(ctx, amp.Spec.Location, amp.Spec.Template.VMSize)
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
