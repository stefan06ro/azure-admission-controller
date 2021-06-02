package releaseversion

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

const (
	ignoreReleaseAnnotation = "release.giantswarm.io/ignore"
)

func Validate(ctx context.Context, ctrlCLient client.Client, oldVersion semver.Version, newVersion semver.Version) error {
	if oldVersion.Equals(newVersion) {
		return nil
	}

	availableReleases, err := availableReleases(ctx, ctrlCLient)
	if err != nil {
		return err
	}

	// Check if new release exists.
	if !included(availableReleases, newVersion) {
		return microerror.Maskf(releaseNotFoundError, "release %s was not found in this installation", newVersion)
	}

	// Skip validations for ignored releases.
	if isOldOrNewReleaseIgnored(availableReleases, oldVersion, newVersion) {
		return nil
	}

	// Downgrades are not allowed.
	if newVersion.LT(oldVersion) {
		return microerror.Maskf(downgradingIsNotAllowedError, "downgrading is not allowed (attempted to downgrade from %s to %s)", oldVersion, newVersion)
	}

	// Check if either version is an alpha one.
	if isAlphaRelease(oldVersion.String()) || isAlphaRelease(newVersion.String()) {
		return microerror.Maskf(upgradingToOrFromAlphaReleaseError, "It is not possible to upgrade to or from an alpha release")
	}

	// Remove alpha and ignored releases from remaining validations logic.
	availableReleases = filterOutAlphaAndIgnoredAndDeprecatedReleases(availableReleases)

	if oldVersion.Major != newVersion.Major || oldVersion.Minor != newVersion.Minor {
		// The major or minor version is changed. We support this only for sequential minor releases (no skip allowed).
		for _, release := range availableReleases {
			if release.Version.EQ(oldVersion) || release.Version.EQ(newVersion) {
				continue
			}
			// Look for a release with higher major or higher minor than the oldVersion and is LT the newVersion
			if release.Version.GT(oldVersion) && release.Version.LT(newVersion) &&
				(oldVersion.Major != release.Version.Major || oldVersion.Minor != release.Version.Minor) &&
				(newVersion.Major != release.Version.Major || newVersion.Minor != release.Version.Minor) {
				// Skipped one major or minor release.
				return microerror.Maskf(skippingReleaseError, "Upgrading from %s to %s is not allowed (skipped %s)", oldVersion, newVersion, release.Version)
			}
		}
	}

	return nil
}

func availableReleases(ctx context.Context, ctrlClient client.Client) ([]*release, error) {
	var releases []*release
	releaseList := &v1alpha1.ReleaseList{}
	err := ctrlClient.List(ctx, releaseList)
	if err != nil {
		return []*release{}, microerror.Mask(err)
	}

	for _, releaseCR := range releaseList.Items {
		parsed, err := semver.ParseTolerant(releaseCR.Name)
		if err != nil {
			return []*release{}, microerror.Maskf(errors.InvalidReleaseError, "Unable to parse release %s to a semver.Release", releaseCR.Name)
		}

		releaseObject := releaseCR
		release := release{
			Version: &parsed,
			CR:      &releaseObject,
		}

		releases = append(releases, &release)
	}

	return releases, nil
}

func filterOutAlphaAndIgnoredAndDeprecatedReleases(releases []*release) []*release {
	var result []*release

	for _, release := range releases {
		if isAlphaRelease(release.Version.String()) {
			continue
		}

		if isIgnoredRelease(release.CR) {
			continue
		}

		if isDeprecatedRelease(release.CR) {
			continue
		}

		result = append(result, release)
	}

	return result
}

func isOldOrNewReleaseIgnored(releases []*release, oldVersion, newVersion semver.Version) bool {
	for _, release := range releases {
		if release.Version.EQ(oldVersion) || release.Version.EQ(newVersion) {
			if isIgnoredRelease(release.CR) {
				return true
			}
		}
	}

	return false
}

func included(releases []*release, releaseVersion semver.Version) bool {
	for _, r := range releases {
		if r.Version.EQ(releaseVersion) {
			return true
		}
	}

	return false
}

func isAlphaRelease(release string) bool {
	return strings.Contains(release, "alpha")
}

func isIgnoredRelease(releaseCR *v1alpha1.Release) bool {
	ignoreValue, isIgnoreAnnotationSet := releaseCR.Annotations[ignoreReleaseAnnotation]
	if isIgnoreAnnotationSet && strings.ToLower(ignoreValue) == "true" {
		return true
	}

	return false
}

func isDeprecatedRelease(releaseCR *v1alpha1.Release) bool {
	return releaseCR.Spec.State == v1alpha1.StateDeprecated
}
