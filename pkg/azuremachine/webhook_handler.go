package azuremachine

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

type Validator struct {
	Decoder

	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type ValidatorConfig struct {
	CtrlClient client.Client
	Location   string
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
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	v := &Validator{
		Decoder: Decoder{},

		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return v, nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
