package mutator

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

var azureOperatorVersionLabelNotFoundError = &microerror.Error{
	Kind: "azureOperatorVersionLabelNotFoundError",
}

// IsAzureOperatorVersionLabelNotFoundError asserts azureOperatorVersionLabelNotFoundError.
func IsAzureOperatorVersionLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == azureOperatorVersionLabelNotFoundError
}

var clusterLabelNotFoundError = &microerror.Error{
	Kind: "clusterLabelNotFoundError",
}

// IsClusterLabelNotFoundError asserts clusterLabelNotFoundError.
func IsClusterLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == clusterLabelNotFoundError
}

var componentNotFoundInReleaseError = &microerror.Error{
	Kind: "componentNotFoundInReleaseError",
}

// IsComponentNotFoundInReleaseError asserts componentNotFoundInReleaseError.
func IsComponentNotFoundInReleaseError(err error) bool {
	return microerror.Cause(err) == componentNotFoundInReleaseError
}

var releaseLabelNotFoundError = &microerror.Error{
	Kind: "releaseLabelNotFoundError",
}

// IsReleaseLabelNotFoundError asserts releaseLabelNotFoundError.
func IsReleaseLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == releaseLabelNotFoundError
}
