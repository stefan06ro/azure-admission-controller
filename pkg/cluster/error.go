package cluster

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

var emptyClusterNetworkError = &microerror.Error{
	Kind: "emptyClusterNetworkError",
}

// IsEmptyClusterNetworkError asserts emptyClusterNetworkError.
func IsEmptyClusterNetworkError(err error) bool {
	return microerror.Cause(err) == emptyClusterNetworkError
}

var emptyClusterNetworkServicesError = &microerror.Error{
	Kind: "emptyClusterNetworkServicesError",
}

// IsEmptyClusterNetworkServicesError asserts emptyClusterNetworkServicesError.
func IsEmptyClusterNetworkServicesError(err error) bool {
	return microerror.Cause(err) == emptyClusterNetworkServicesError
}

var unexpectedAPIServerPortError = &microerror.Error{
	Kind: "unexpectedAPIServerPortError",
}

// IsUnexpectedAPIServerPortError asserts unexpectedAPIServerPortError.
func IsUnexpectedAPIServerPortError(err error) bool {
	return microerror.Cause(err) == unexpectedAPIServerPortError
}

var unexpectedServiceDomainError = &microerror.Error{
	Kind: "unexpectedServiceDomainError",
}

// IsUnexpectedServiceDomainError asserts unexpectedServiceDomainError.
func IsUnexpectedServiceDomainError(err error) bool {
	return microerror.Cause(err) == unexpectedServiceDomainError
}

var unexpectedCIDRBlocksError = &microerror.Error{
	Kind: "unexpectedCIDRBlocksError",
}

// IsUnexpectedCIDRBlocksError asserts unexpectedCIDRBlocksError.
func IsUnexpectedCIDRBlocksError(err error) bool {
	return microerror.Cause(err) == unexpectedCIDRBlocksError
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

var clusterNetworkWasChangedError = &microerror.Error{
	Kind: "clusterNetworkWasChangedError",
}

// IsClusterNetworkWasChangedError asserts clusterNetworkWasChangedError.
func IsClusterNetworkWasChangedError(err error) bool {
	return microerror.Cause(err) == clusterNetworkWasChangedError
}
