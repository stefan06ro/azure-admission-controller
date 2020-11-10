package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type CreateMutatorConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
	VMcaps     *vmcapabilities.VMSKU
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	m := &CreateMutator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return m, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureMPCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureMPCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	patch, err := m.ensureLocation(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureStorageAccountType(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureDataDisks(ctx, azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.EnsureReleaseVersionLabel(ctx, m.ctrlClient, azureMPCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.CopyComponentVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, azureMPCR.GetObjectMeta(), label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "azuremachinepool"
}

func (m *CreateMutator) ensureStorageAccountType(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if mpCR.Spec.Template.OSDisk.ManagedDisk.StorageAccountType == "" {
		// We need to set the default value as it is missing.

		location := mpCR.Spec.Location
		if location == "" {
			// The location was empty and we are adding it using this same mutator.
			// We assume it will be set to the installation's location.
			location = m.location
		}

		// Check if the VM has Premium Storage capability.
		premium, err := m.vmcaps.HasCapability(ctx, location, mpCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var storageAccountType string
		{
			if premium {
				storageAccountType = string(compute.StorageAccountTypesPremiumLRS)
			} else {
				storageAccountType = string(compute.StorageAccountTypesStandardLRS)
			}
		}

		return mutator.PatchAdd("/spec/template/osDisk/managedDisk/storageAccountType", storageAccountType), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureDataDisks(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Template.DataDisks) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/template/dataDisks", desiredDataDisks), nil
}

func (m *CreateMutator) ensureLocation(ctx context.Context, mpCR *expcapzv1alpha3.AzureMachinePool) (*mutator.PatchOperation, error) {
	if len(mpCR.Spec.Location) > 0 {
		return nil, nil
	}

	return mutator.PatchAdd("/spec/location", m.location), nil
}
