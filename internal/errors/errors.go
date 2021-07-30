package errors

import (
	"errors"
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	capiErrorMessageRegexp = regexp.MustCompile(`(.*is invalid:\s)(\[)?(.*?)(\]|$)`)
)

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

func IgnoreCAPIErrorForField(field string, err error) error {
	if status := apierrors.APIStatus(nil); errors.As(err, &status) {
		errStatus := status.Status()
		if errStatus.Reason != "Invalid" {
			return err
		}

		if errStatus.Details == nil {
			return err
		}

		// Remove any errors for the given field.
		var causes []metav1.StatusCause
		{
			for _, cause := range errStatus.Details.Causes {
				if !strings.HasPrefix(cause.Field, field) {
					causes = append(causes, cause)
				}
			}

			if len(causes) < 1 {
				// No errors left, all clear.
				return nil
			}
		}

		matches := capiErrorMessageRegexp.FindAllStringSubmatch(errStatus.Message, 1)
		var messageBuilder strings.Builder
		{
			messageBuilder.WriteString(matches[0][1])
			messageBuilder.WriteString("[")

			for i, cause := range causes {
				messageBuilder.WriteString(cause.Field)
				messageBuilder.WriteString(": ")
				messageBuilder.WriteString(cause.Message)

				if len(causes)-i > 1 {
					messageBuilder.WriteString(", ")
				}
			}

			messageBuilder.WriteString("]")
		}

		errStatus.Details.Causes = causes
		errStatus.Message = messageBuilder.String()

		return apierrors.FromObject(&errStatus)
	}

	return err
}

var WrongTypeError = &microerror.Error{
	Kind: "WrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == WrongTypeError
}
