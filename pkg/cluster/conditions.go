package cluster

import (
	"context"

	aeV3conditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func (m *UpdateMutator) ensureConditions(ctx context.Context, oldCluster *capiv1alpha3.Cluster, newCluster *capiv1alpha3.Cluster) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	patch, err := m.ensureUpdateCondition(ctx, oldCluster, newCluster)
	if err != nil {
		return []mutator.PatchOperation{}, microerror.Mask(err)
	}
	if patch != nil {
		result = append(result, *patch)
	}

	return result, nil
}

func (m *UpdateMutator) ensureUpdateCondition(ctx context.Context, oldCluster *capiv1alpha3.Cluster, newCluster *capiv1alpha3.Cluster) (*mutator.PatchOperation, error) {
	var patch *mutator.PatchOperation

	oldClusterVersion, err := semverhelper.GetSemverFromLabels(oldCluster.Labels)
	if err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(newCluster.Labels)
	if err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	if newClusterVersion.GT(oldClusterVersion) {
		upgradingCondition := capiconditions.TrueCondition(aeV3conditions.UpgradingCondition)
		patch = mutator.PatchAdd("/status/conditions/-/", *upgradingCondition)
	}

	return patch, nil
}
