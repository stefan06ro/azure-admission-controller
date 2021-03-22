package azuremachinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type UpdateMutator struct {
	ctrlClient ctrl.Client
	logger     micrologger.Logger
}

type UpdateMutatorConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger
}

func NewUpdateMutator(config UpdateMutatorConfig) (*UpdateMutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &UpdateMutator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return m, nil
}

func (m *UpdateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var err error
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureMPCR := &expcapzv1alpha3.AzureMachinePool{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureMPCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse azureMachinePool CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(azureMPCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if capi {
		return []mutator.PatchOperation{}, nil
	}

	azureMPCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFrom(request.Object.Raw, azureMPCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/template/sshPublicKey", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *UpdateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *UpdateMutator) Resource() string {
	return "azurecluster"
}
