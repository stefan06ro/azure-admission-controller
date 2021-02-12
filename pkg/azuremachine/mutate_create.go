package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
}

type CreateMutatorConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}

	m := &CreateMutator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
	}

	return m, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureMachineCR := &v1alpha3.AzureMachine{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureMachineCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse AzureMachine CR: %v", err)
	}

	patch, err := m.ensureLocation(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureOSDiskCachingType(ctx, azureMachineCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, azureMachineCR.GetObjectMeta())
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
	return "azuremachine"
}

func (m *CreateMutator) ensureLocation(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if azureMachine.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", m.location), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureOSDiskCachingType(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if len(azureMachine.Spec.OSDisk.CachingType) < 1 {
		return mutator.PatchAdd("/spec/osDisk/cachingType", key.OSDiskCachingType()), nil
	}

	return nil, nil
}
