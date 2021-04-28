package cluster

import (
	"context"
	"reflect"

	aeconditions "github.com/giantswarm/apiextensions/v3/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/azure-admission-controller/internal/conditions"
	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	clusterNewCR, err := key.ToClusterPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	clusterOldCR, err := key.ToClusterPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	err = clusterNewCR.ValidateUpdate(clusterOldCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(clusterOldCR, clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateClusterNetworkUnchanged(*clusterOldCR, *clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpointUnchanged(*clusterOldCR, *clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = conditions.ValidateClusterConditions(clusterOldCR, clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return h.validateRelease(ctx, clusterOldCR, clusterNewCR)
}

func validateClusterNetworkUnchanged(old capi.Cluster, new capi.Cluster) error {
	// Was nil and stayed nil. Not good but not changed so ok from this validator point of view.
	if old.Spec.ClusterNetwork == nil && new.Spec.ClusterNetwork == nil {
		return nil
	}

	// Was nil or became nil.
	if old.Spec.ClusterNetwork == nil && new.Spec.ClusterNetwork != nil ||
		old.Spec.ClusterNetwork != nil && new.Spec.ClusterNetwork == nil {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Check APIServerPort and ServiceDomain is unchanged.
	if *old.Spec.ClusterNetwork.APIServerPort != *new.Spec.ClusterNetwork.APIServerPort ||
		old.Spec.ClusterNetwork.ServiceDomain != new.Spec.ClusterNetwork.ServiceDomain {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Was nil and stayed nil. Not good but not changed so ok from this validator point of view.
	if old.Spec.ClusterNetwork.Services == nil && new.Spec.ClusterNetwork.Services == nil {
		return nil
	}

	// Check Services have not blanked out.
	if old.Spec.ClusterNetwork.Services == nil && new.Spec.ClusterNetwork.Services != nil ||
		old.Spec.ClusterNetwork.Services != nil && new.Spec.ClusterNetwork.Services == nil {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Check Services didn't change.
	if !reflect.DeepEqual(*old.Spec.ClusterNetwork.Services, *new.Spec.ClusterNetwork.Services) {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	return nil
}

func (h *WebhookHandler) validateRelease(ctx context.Context, clusterOldCR *capi.Cluster, clusterNewCR *capi.Cluster) error {
	oldClusterVersion, err := semverhelper.GetSemverFromLabels(clusterOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from the Cluster being updated")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(clusterNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from applied Cluster")
	}

	if !newClusterVersion.Equals(oldClusterVersion) {
		// Upgrade is triggered, let's check if we allow it
		if capiconditions.IsTrue(clusterOldCR, aeconditions.CreatingCondition) {
			return microerror.Maskf(errors.InvalidOperationError, "upgrade cannot be initiated now, Cluster condition %s is set to True, cluster is currently being created", aeconditions.CreatingCondition)
		} else if capiconditions.IsTrue(clusterOldCR, aeconditions.UpgradingCondition) {
			return microerror.Maskf(errors.InvalidOperationError, "upgrade cannot be initiated now, Cluster condition %s is set to True, cluster is already being upgraded", aeconditions.UpgradingCondition)
		}
	}

	return releaseversion.Validate(ctx, h.ctrlClient, oldClusterVersion, newClusterVersion)
}
