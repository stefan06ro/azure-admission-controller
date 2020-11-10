package errors

import "github.com/giantswarm/microerror"

var InvalidUpgradingConditionMessageFormatError = &microerror.Error{
	Kind: "InvalidUpgradingConditionMessageFormatError",
}

// IsInvalidUpgradingConditionMessageFormat asserts InvalidUpgradingConditionMessageFormatError.
func IsInvalidUpgradingConditionMessageFormat(err error) bool {
	return microerror.Cause(err) == InvalidUpgradingConditionMessageFormatError
}

var InvalidReleaseVersionInUpgradingConditionMessageError = &microerror.Error{
	Kind: "InvalidReleaseVersionInUpgradingConditionMessageError",
}

// IsInvalidReleaseVersionInUpgradingConditionMessage asserts InvalidReleaseVersionInUpgradingConditionMessageError.
func IsInvalidReleaseVersionInUpgradingConditionMessage(err error) bool {
	return microerror.Cause(err) == InvalidReleaseVersionInUpgradingConditionMessageError
}

var InvalidConditionStatusError = &microerror.Error{
	Kind: "InvalidConditionStatusError",
}

// IsInvalidConditionStatus asserts InvalidConditionStatusError.
func IsInvalidConditionStatus(err error) bool {
	return microerror.Cause(err) == InvalidConditionStatusError
}

var InvalidConditionModificationError = &microerror.Error{
	Kind: "InvalidConditionModification",
}

// IsInvalidConditionModification asserts InvalidConditionModification.
func IsInvalidConditionModification(err error) bool {
	return microerror.Cause(err) == InvalidConditionModificationError
}

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
