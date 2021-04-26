package azuremachine

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type Mutator struct {
	Decoder
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
}

type MutatorConfig struct {
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
}

func NewMutator(config MutatorConfig) (*Mutator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}

	m := &Mutator{
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,

		Decoder: Decoder{},
	}

	return m, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "azuremachine"
}

func (m *Mutator) ensureOSDiskCachingType(ctx context.Context, azureMachine *v1alpha3.AzureMachine) (*mutator.PatchOperation, error) {
	if len(azureMachine.Spec.OSDisk.CachingType) < 1 {
		return mutator.PatchAdd("/spec/osDisk/cachingType", key.OSDiskCachingType()), nil
	}

	return nil, nil
}
