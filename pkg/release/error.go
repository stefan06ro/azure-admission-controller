package release

import (
	"github.com/giantswarm/microerror"
)

var ReleaseNotFoundError = &microerror.Error{
	Kind: "ReleaseNotFoundError",
}

// IsReleaseNotFoundError asserts ReleaseNotFoundError.
func IsReleaseNotFoundError(err error) bool {
	return microerror.Cause(err) == ReleaseNotFoundError
}
