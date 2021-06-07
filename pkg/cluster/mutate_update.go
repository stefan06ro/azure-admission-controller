package cluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type UpdateMutator struct {
	ctrlCache  ctrl.Reader
	ctrlClient ctrl.Client
	logger     micrologger.Logger
}

type UpdateMutatorConfig struct {
	CtrlCache  ctrl.Reader
	CtrlClient ctrl.Client
	Logger     micrologger.Logger
}

func NewUpdateMutator(config UpdateMutatorConfig) (*UpdateMutator, error) {
	if config.CtrlCache == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlCache must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &UpdateMutator{
		ctrlCache:  config.CtrlCache,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return m, nil
}

func (m *UpdateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	clusterCR := &capi.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, clusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if capi {
		return []mutator.PatchOperation{}, nil
	}

	patch, err := mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlCache, clusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlCache, clusterCR.GetObjectMeta(), "cluster-operator", label.ClusterOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	clusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFrom(request.Object.Raw, clusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *UpdateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *UpdateMutator) Resource() string {
	return "cluster"
}
