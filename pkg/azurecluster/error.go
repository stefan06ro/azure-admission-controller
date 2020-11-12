package azurecluster

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

var parsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == parsingFailedError
}

var invalidControlPlaneEndpointHostError = &microerror.Error{
	Kind: "invalidControlPlaneEndpointHostError",
}

// IsInvalidControlPlaneEndpointHostError asserts invalidControlPlaneEndpointHostError.
func IsInvalidControlPlaneEndpointHostError(err error) bool {
	return microerror.Cause(err) == invalidControlPlaneEndpointHostError
}

var invalidControlPlaneEndpointPortError = &microerror.Error{
	Kind: "invalidControlPlaneEndpointPortError",
}

// IsInvalidControlPlaneEndpointPortError asserts invalidControlPlaneEndpointPortError.
func IsInvalidControlPlaneEndpointPortError(err error) bool {
	return microerror.Cause(err) == invalidControlPlaneEndpointPortError
}

var controlPlaneEndpointWasChangedError = &microerror.Error{
	Kind: "controlPlaneEndpointWasChangedError",
}

// IsControlPlaneEndpointWasChangedError asserts controlPlaneEndpointWasChangedError.
func IsControlPlaneEndpointWasChangedError(err error) bool {
	return microerror.Cause(err) == controlPlaneEndpointWasChangedError
}

var locationWasChangedError = &microerror.Error{
	Kind: "locationWasChangedError",
}

// IsFailureDomainWasChangedError asserts locationWasChangedError.
func IsLocationWasChangedError(err error) bool {
	return microerror.Cause(err) == locationWasChangedError
}

var unexpectedLocationError = &microerror.Error{
	Kind: "unexpectedLocationError",
}

// IsUnexpectedLocationError asserts unexpectedLocationError.
func IsUnexpectedLocationError(err error) bool {
	return microerror.Cause(err) == unexpectedLocationError
}
