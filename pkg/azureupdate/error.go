package azureupdate

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var masterCIDRChangeError = &microerror.Error{
	Kind: "masterCIDRChangeError",
}

// IsMasterCIDRChange asserts masterCIDRChangeError.
func IsMasterCIDRChange(err error) bool {
	return microerror.Cause(err) == masterCIDRChangeError
}
