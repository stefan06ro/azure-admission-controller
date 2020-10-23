package azuremachinepool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	logger micrologger.Logger
	vmcaps *vmcapabilities.VMSKU
}

type CreateMutatorConfig struct {
	Logger micrologger.Logger
	VMcaps *vmcapabilities.VMSKU
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	m := &CreateMutator{
		logger: config.Logger,
		vmcaps: config.VMcaps,
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

	patch, err := generic.EnsureOrganizationLabelNormalized(ctx, azureMPCR)
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

		// Check if the VM has Premium Storage capability.
		premium, err := m.vmcaps.HasCapability(ctx, mpCR.Spec.Location, mpCR.Spec.Template.VMSize, vmcapabilities.CapabilityPremiumIO)
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
