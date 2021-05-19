package cluster

import (
	"context"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
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
	logger     micrologger.Logger
}

type CreateMutatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
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

	v := &CreateMutator{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
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

	clusterCR := &capi.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, clusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	capi, err := generic.IsCAPIRelease(clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if capi {
		return []mutator.PatchOperation{}, nil
	}

	patch, err := m.ensureClusterNetwork(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointHost(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneEndpointPort(ctx, clusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, m.ctrlClient, clusterCR.GetObjectMeta())
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

func (m *CreateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *CreateMutator) Resource() string {
	return "cluster"
}

func (m *CreateMutator) ensureClusterNetwork(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	// Ensure ClusterNetwork is set.
	if clusterCR.Spec.ClusterNetwork == nil {
		clusterNetwork := capi.ClusterNetwork{
			APIServerPort: to.Int32Ptr(key.ControlPlaneEndpointPort),
			ServiceDomain: key.ServiceDomain(),
			Services: &capi.NetworkRanges{
				CIDRBlocks: []string{
					key.ClusterNetworkServiceCIDR,
				},
			},
		}

		return mutator.PatchAdd("/spec/clusterNetwork", clusterNetwork), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capi.Cluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}
