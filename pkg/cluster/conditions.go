package cluster

import (
	"context"

	"github.com/blang/semver"
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

	nodepoolsReleaseVersion := semver.Version{
		Major: 13,
		Minor: 0,
		Patch: 0,
	}

	// set Upgrading condition when upgrading to a new release, only if new release is newer than
	// or equal to 13.0.1 which is handling Upgrading condition
	if newClusterVersion.GT(oldClusterVersion) && newClusterVersion.GTE(nodepoolsReleaseVersion) {
		upgradingCondition := capiconditions.TrueCondition(aeV3conditions.UpgradingCondition)
		patch = mutator.PatchAdd("/status/conditions/-", []capiv1alpha3.Condition{*upgradingCondition})
	}

	return patch, nil
}
