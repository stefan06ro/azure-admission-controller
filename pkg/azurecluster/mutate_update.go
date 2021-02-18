package azurecluster

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/patches"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type UpdateMutator struct {
	ctrlClient ctrl.Client
	logger     micrologger.Logger
}

type UpdateMutatorConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger
}

func NewUpdateMutator(config UpdateMutatorConfig) (*UpdateMutator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	m := &UpdateMutator{
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

	azureClusterCR := &capz.AzureCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return []mutator.PatchOperation{}, microerror.Maskf(parsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	patch, err := m.ensureAPIServerLBName(ctx, azureClusterCR)
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

	patch, err = generic.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, azureClusterCR.GetObjectMeta(), "azure-operator", label.AzureOperatorVersion)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	patch, err = generic.EnsureComponentVersionLabelFromRelease(ctx, m.ctrlClient, azureClusterCR.GetObjectMeta(), "cluster-operator", label.ClusterOperatorVersion)
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

func (m *UpdateMutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *UpdateMutator) Resource() string {
	return "azurecluster"
}

func (m *UpdateMutator) ensureAPIServerLBName(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.Name) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/name", key.APIServerLBName(azureCluster.Name)), nil
	}

	return nil, nil
}

func (m *UpdateMutator) ensureAPIServerLBSKU(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.SKU) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/sku", key.APIServerLBSKU()), nil
	}

	return nil, nil
}

func (m *UpdateMutator) ensureAPIServerLBType(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.Type) < 1 {
		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/type", key.APIServerLBType()), nil
	}

	return nil, nil
}

func (m *UpdateMutator) ensureAPIServerLBFrontendIPs(ctx context.Context, azureCluster *capzv1alpha3.AzureCluster) (*mutator.PatchOperation, error) {
	if len(azureCluster.Spec.NetworkSpec.APIServerLB.FrontendIPs) < 1 {
		frontendIPs := []capzv1alpha3.FrontendIP{
			{Name: key.APIServerLBFrontendIPName(azureCluster.Name)},
		}

		return mutator.PatchAdd("/spec/networkSpec/apiServerLB/frontendIPs", frontendIPs), nil
	}

	return nil, nil
}
