package release

import (
	"github.com/giantswarm/microerror"
)

var releaseNotFoundError = &microerror.Error{
	Kind: "releaseNotFoundError",
}

// IsReleaseNotFoundError asserts releaseNotFoundError.
func IsReleaseNotFoundError(err error) bool {
	return microerror.Cause(err) == releaseNotFoundError
}
