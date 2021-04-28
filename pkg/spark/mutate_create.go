package spark

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type Mutator struct {
	Decoder

	ctrlClient client.Client
	logger     micrologger.Logger
}

type MutatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewMutator(config MutatorConfig) (*Mutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &Mutator{
		Decoder: Decoder{},

		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return m, nil
}

func (m *Mutator) OnCreateMutate(ctx context.Context, object interface{}) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	sparkCR, err := key.ToSparkPtr(object)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}

	patch, err := mutator.EnsureReleaseVersionLabel(ctx, m.ctrlClient, sparkCR.GetObjectMeta())
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "spark"
}
