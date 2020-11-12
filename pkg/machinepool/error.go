package machinepool

import (
	"github.com/giantswarm/microerror"
)

var azureMachinePoolNotFoundError = &microerror.Error{
	Kind: "azureMachinePoolNotFoundError",
}

// IsAzureMachinePoolNotFound asserts azureMachinePoolNotFoundError.
func IsAzureMachinePoolNotFound(err error) bool {
	return microerror.Cause(err) == azureMachinePoolNotFoundError
}

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

var unsupportedFailureDomainError = &microerror.Error{
	Kind: "unsupportedFailureDomainError",
}

// IsUnsupportedFailureDomainError asserts unsupportedFailureDomainError.
func IsUnsupportedFailureDomainError(err error) bool {
	return microerror.Cause(err) == unsupportedFailureDomainError
}

var locationWithNoFailureDomainSupportError = &microerror.Error{
	Kind: "locationWithNoFailureDomainSupportError",
}

// IsLocationWithNoFailureDomainSupportError asserts locationWithNoFailureDomainSupportError.
func IsLocationWithNoFailureDomainSupportError(err error) bool {
	return microerror.Cause(err) == locationWithNoFailureDomainSupportError
}

var failureDomainWasChangedError = &microerror.Error{
	Kind: "failureDomainWasChangedError",
}

// IsFailureDomainWasChangedError asserts failureDomainWasChangedError.
func IsFailureDomainWasChangedError(err error) bool {
	return microerror.Cause(err) == failureDomainWasChangedError
}
