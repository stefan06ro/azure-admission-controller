package semverhelper

import (
	"github.com/blang/semver"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func GetSemverFromLabels(labels map[string]string) (semver.Version, error) {
	version, ok := labels[label.ReleaseVersion]
	if !ok {
		return semver.Version{}, microerror.Maskf(errors.ParsingFailedError, "CR didn't have expected label %s", label.ReleaseVersion)
	}

	return GetSemverFromString(version)
}

func GetSemverFromString(version string) (semver.Version, error) {
	return semver.ParseTolerant(version)
}
