package machinepool

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

type Validator struct {
	Decoder

	ctrlClient client.Client
	logger     micrologger.Logger
	vmcaps     *vmcapabilities.VMSKU
}

type ValidatorConfig struct {
	CtrlClient client.Client
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
	if config.VMcaps == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VMcaps must not be empty", config)
	}

	admitter := &Validator{
		Decoder: Decoder{},

		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		vmcaps:     config.VMcaps,
	}

	return admitter, nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
