package cluster

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Validator struct {
	Decoder

	baseDomain string
	ctrlClient client.Client
	logger     micrologger.Logger
}

type ValidatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewValidator(config ValidatorConfig) (*Validator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &Validator{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,

		Decoder: Decoder{},
	}

	return v, nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
