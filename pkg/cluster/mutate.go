package cluster

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Mutator struct {
	Decoder

	baseDomain string
	ctrlClient client.Client
	logger     micrologger.Logger
}

type MutatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewMutator(config MutatorConfig) (*Mutator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &Mutator{
		Decoder: Decoder{},

		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return v, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "cluster"
}
