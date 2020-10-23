package cluster

import (
	"context"

	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
)

func (a *UpdateValidator) validateRelease(ctx context.Context, clusterOldCR *capi.Cluster, clusterNewCR *capi.Cluster) (bool, error) {
	oldClusterVersion, err := semverhelper.GetSemverFromLabels(clusterOldCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(clusterNewCR.Labels)
	if err != nil {
		return false, microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}
