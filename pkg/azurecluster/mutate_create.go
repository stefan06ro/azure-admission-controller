package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
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

	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
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

	patch, err = m.ensureAPIServerLBName(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureAPIServerLBSKU(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureAPIServerLBType(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureAPIServerLBFrontendIPs(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = m.ensureControlPlaneSubnet(ctx, azureClusterCR)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
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
	return "azurecluster"
}

func (m *CreateMutator) ensureControlPlaneEndpointHost(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Host == "" {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/host", key.GetControlPlaneEndpointHost(clusterCR.Name, m.baseDomain)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneEndpointPort(ctx context.Context, clusterCR *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if clusterCR.Spec.ControlPlaneEndpoint.Port == 0 {
		return mutator.PatchAdd("/spec/controlPlaneEndpoint/port", key.ControlPlaneEndpointPort), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureLocation(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if azureCluster.Spec.Location == "" {
		return mutator.PatchAdd("/spec/location", m.location), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureAPIServerLBName(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.Name) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/name", key.APIServerLBName(azureCluster.Name)), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureAPIServerLBSKU(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.SKU) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/sku", key.APIServerLBSKU()), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureAPIServerLBType(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.Type) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/type", key.APIServerLBType()), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureAPIServerLBFrontendIPs(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.FrontendIPs) < 1 {
		frontendIPs := []capzv1alpha3.FrontendIP{
			{Name: key.APIServerLBFrontendIPName(azureCluster.Name)},
		}

		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/frontendIPs", frontendIPs), nil
	}

	return nil, nil
}

func (m *CreateMutator) ensureControlPlaneSubnet(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	hasControlPlaneSubnet := false
	for _, subnet := range azureCluster.Spec.NetworkSpec.Subnets {
		if subnet.Role == capzv1alpha3.SubnetControlPlane {
			hasControlPlaneSubnet = true
			break
		}
	}

	if !hasControlPlaneSubnet {
		subnets := azureCluster.Spec.NetworkSpec.Subnets[:]
		subnets = append(subnets, &capzv1alpha3.SubnetSpec{
			Role: capzv1alpha3.SubnetControlPlane,
			Name: key.MasterSubnetName(azureCluster.Name),
		})

		return mutator.PatchAdd("/spec/networkSpec/subnets", subnets), nil
	}

	return nil, nil
}
