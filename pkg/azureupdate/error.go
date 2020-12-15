package azureupdate

import (
	"github.com/giantswarm/microerror"
)

var availabilityZonesChangeError = &microerror.Error{
	Kind: "availabilityZonesChangeError",
}

// IsAvailabilityZonesChange asserts availabilityZonesChangeError.
func IsAvailabilityZonesChange(err error) bool {
	return microerror.Cause(err) == availabilityZonesChangeError
}

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
