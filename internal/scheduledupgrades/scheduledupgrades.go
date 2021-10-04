package scheduledupgrades

import (
	"context"
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
)

func ValidateClusterAnnotationUpgradeTime(oldCluster *capi.Cluster, newCluster *capi.Cluster) error {
	if updateTime, ok := newCluster.GetAnnotations()[annotation.UpdateScheduleTargetTime]; ok {
		if oldCluster != nil {
			if updateTimeOld, ok := oldCluster.GetAnnotations()[annotation.UpdateScheduleTargetTime]; ok {
				if updateTime == updateTimeOld {
					return nil
				}
			}
		}
		if !ValidateUpgradeScheduleTime(updateTime) {
			return microerror.Maskf(notAllowedError,
				fmt.Sprintf("Cluster annotation '%s' value '%s' is not valid. Value must be in RFC822 format and UTC time zone (e.g. 30 Jan 21 15:04 UTC) and should be a date 16 mins - 6months in the future.",
					annotation.UpdateScheduleTargetTime,
					updateTime),
			)
		}
	}
	return nil
}

func ValidateUpgradeScheduleTime(updateTime string) bool {
	// parse time
	t, err := time.Parse(time.RFC822, updateTime)
	if err != nil {
		return false
	}
	// check whether it is UTC
	if t.Location().String() != "UTC" {
		return false
	}
	// time already passed or is less than 16 minutes in the future
	if t.Before(time.Now().UTC().Add(16 * time.Minute)) {
		return false
	}
	// time is 6 months or more in the future (6 months are 4380 hours)
	if t.Sub(time.Now().UTC()) > 4380*time.Hour {
		return false
	}
	return true
}

func ValidateClusterAnnotationUpgradeRelease(ctx context.Context, client client.Client, cluster *capi.Cluster) error {
	if targetRelease, ok := cluster.GetAnnotations()[annotation.UpdateScheduleTargetRelease]; ok {
		oldVersion, err := semverhelper.GetSemverFromLabels(cluster.Labels)
		if err != nil {
			return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from the Cluster being updated")
		}

		newVersion, err := semver.Parse(targetRelease)
		if err != nil {
			return microerror.Mask(err)
		}

		err = releaseversion.Validate(ctx, client, oldVersion, newVersion)
		if err != nil {
			return microerror.Maskf(notAllowedError,
				fmt.Sprintf("Cluster annotation '%s' value '%s' is not valid. Value must be an existing giant swarm release version above the current release version %s and must not have a v prefix. %v",
					annotation.UpdateScheduleTargetRelease,
					targetRelease,
					oldVersion.String(),
					err),
			)
		}

		if oldVersion.Equals(newVersion) {
			return microerror.Maskf(notAllowedError,
				fmt.Sprintf("Cluster annotation '%s' value '%s' is not valid. The target release must be newer than the current cluster's release.",
					annotation.UpdateScheduleTargetRelease,
					targetRelease,
				),
			)
		}
	}
	return nil
}
