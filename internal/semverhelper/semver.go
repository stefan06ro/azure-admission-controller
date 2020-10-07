package semverhelper

import (
	"github.com/blang/semver"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

const (
	versionLabel = "release.giantswarm.io/version"
)

func GetSemverFromLabels(labels map[string]string) (semver.Version, error) {
	version, ok := labels[versionLabel]
	if !ok {
		return semver.Version{}, microerror.Maskf(errors.ParsingFailedError, "CR didn't have expected label %s", versionLabel)
	}

	return GetSemverFromString(version)
}

func GetSemverFromString(version string) (semver.Version, error) {
	return semver.ParseTolerant(version)
}
