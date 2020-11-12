package azuremachine

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

var locationWasChangedError = &microerror.Error{
	Kind: "locationWasChangedError",
}

// IsLocationWasChangedError asserts locationWasChangedError.
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

var sshFieldIsSetError = &microerror.Error{
	Kind: "sshFieldIsSetError",
}

// IsSSHFieldIsSetError asserts sshFieldIsSetError.
func IsSSHFieldIsSetError(err error) bool {
	return microerror.Cause(err) == sshFieldIsSetError
}
