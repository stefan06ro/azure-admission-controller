package releaseversion

import (
	"github.com/blang/semver"
	"github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
)

type release struct {
	Version *semver.Version
	CR      *v1alpha1.Release
}
