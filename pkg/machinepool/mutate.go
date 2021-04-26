package machinepool

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "machinepool"
}
