package machinepool

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	logger micrologger.Logger
}

type CreateMutatorConfig struct {
	Logger micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &CreateMutator{
		logger: config.Logger,
	}

	return m, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	m.logger.LogCtx(ctx, "level", "debug", "message", "mutating MachinePool create")

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	machinePoolCR := &capiexp.MachinePool{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, machinePoolCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse MachinePool CR: %v", err)
	}

	defaultSpecValues := m.setDefaultSpecValues(ctx, machinePoolCR)
	if defaultSpecValues != nil {
		result = append(result, defaultSpecValues...)
	}

	return result, nil
}

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "machinepool"
}
