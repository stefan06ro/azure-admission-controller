package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type CreateMutator struct {
	baseDomain string
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
}

type CreateMutatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
}

func NewCreateMutator(config CreateMutatorConfig) (*CreateMutator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}

	v := &CreateMutator{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
	}

	return v, nil
}

func (m *CreateMutator) Mutate(ctx context.Context, request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		m.logger.LogCtx(ctx, "level", "debug", "message", "Dry run is not supported. Request processing stopped.")
		return result, nil
	}

	azureClusterCR := &capz.AzureCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if capi {
		return []mutator.PatchOperation{}, nil
	}

	patch, err := m.ensureControlPlaneEndpointHost(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureLocation(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = ensureAPIServerLB(azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = mutator.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	azureClusterCR.Default()
	{
		var capiPatches []mutator.PatchOperation
		capiPatches, err = patches.GenerateFrom(request.Object.Raw, azureClusterCR)
		if err != nil {
			return []mutator.PatchOperation{}, microerror.Mask(err)
		}

		capiPatches = patches.SkipForPath("/spec/networkSpec/vnet", capiPatches)
		capiPatches = patches.SkipForPath("/spec/networkSpec/subnets", capiPatches)

		result = append(result, capiPatches...)
	}

	return result, nil
}

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "azurecluster"
}

func (m *CreateMutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureLocation(ctx context.Context, azureCluster *capz.AzureCluster) (*mutator.PatchOperation, error) {
	if azureCluster.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", m.location), nil
	}

	return nil, nil
}
