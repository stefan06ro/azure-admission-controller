package errors

import "github.com/giantswarm/microerror"

var InvalidOperationError = &microerror.Error{
	Kind: "invalidOperationError",
}

// IsInvalidOperationError asserts invalidOperationError.
func IsInvalidOperationError(err error) bool {
	return microerror.Cause(err) == InvalidOperationError
}

var InvalidReleaseError = &microerror.Error{
	Kind: "invalidReleaseError",
}

// IsInvalidReleaseError asserts ParsingFailedError.
func IsInvalidReleaseError(err error) bool {
	return microerror.Cause(err) == InvalidReleaseError
}

var NotFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFoundError asserts NotFoundError.
func IsNotFoundError(err error) bool {
	return microerror.Cause(err) == NotFoundError
}

var ParsingFailedError = &microerror.Error{
	Kind: "ParsingFailedError",
}

// IsParsingFailed asserts ParsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == ParsingFailedError
}

var UnknownReleaseError = &microerror.Error{
	Kind: "UnknownReleaseError",
}

// IsUnknownReleaseError asserts parsingFailedError.
func IsUnknownReleaseError(err error) bool {
	return microerror.Cause(err) == UnknownReleaseError
}
