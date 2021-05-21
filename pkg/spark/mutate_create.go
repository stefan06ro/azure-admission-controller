package spark

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type CreateMutatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &CreateMutator{
		ctrlClient: config.CtrlClient,
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

	sparkCR := &v1alpha1.Spark{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, sparkCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse Spark CR: %v", err)
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

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "spark"
}
