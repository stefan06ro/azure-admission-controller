package generic

import (
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"

	"github.com/blang/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FirstCAPIRelease is the first GS release that runs on CAPI controllers
	FirstCAPIRelease = "20.0.0-v1alpha3"
)

func IsCAPIRelease(meta metav1.Object) (bool, error) {
	if meta.GetLabels()[label.ReleaseVersion] == "" {
		return false, nil
	}
	releaseVersion, err := ReleaseVersion(meta)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse release version from object")
	}
	return IsCAPIVersion(releaseVersion)
}

// IsCAPIVersion returns whether a given releaseVersion is using CAPI controllers
func IsCAPIVersion(releaseVersion *semver.Version) (bool, error) {
	CAPIVersion, err := semver.New(FirstCAPIRelease)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to get first CAPI release version")
	}
	return releaseVersion.GE(*CAPIVersion), nil
}

func ReleaseVersion(meta metav1.Object) (*semver.Version, error) {
	version, ok := meta.GetLabels()[label.ReleaseVersion]
	if !ok {
		return nil, microerror.Maskf(parsingFailedError, "unable to get release version from Object %s", meta.GetName())
	}
	version = strings.TrimPrefix(version, "v")
	return semver.New(version)
}
