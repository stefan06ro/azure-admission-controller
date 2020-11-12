package generic

import (
	"github.com/giantswarm/microerror"
)

var organizationLabelNotFoundError = &microerror.Error{
	Kind: "organizationLabelNotFoundError",
}

// IsOrganizationLabelNotFoundError asserts organizationLabelNotFoundError.
func IsOrganizationLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == organizationLabelNotFoundError
}

var organizationNotFoundError = &microerror.Error{
	Kind: "organizationNotFoundError",
}

// IsOrganizationNotFoundError asserts organizationNotFoundError.
func IsOrganizationNotFoundError(err error) bool {
	return microerror.Cause(err) == organizationNotFoundError
}

var organizationLabelWasChangedError = &microerror.Error{
	Kind: "organizationLabelWasChangedError",
}

// IsOrganizationLabelWasChangedError asserts organizationLabelWasChangedError.
func IsOrganizationLabelWasChangedError(err error) bool {
	return microerror.Cause(err) == organizationLabelWasChangedError
}

var clusterLabelNotFoundError = &microerror.Error{
	Kind: "clusterLabelNotFoundError",
}

// IsClusterLabelNotFoundError asserts clusterLabelNotFoundError.
func IsClusterLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == clusterLabelNotFoundError
}

var releaseLabelNotFoundError = &microerror.Error{
	Kind: "releaseLabelNotFoundError",
}

// IsReleaseLabelNotFoundError asserts releaseLabelNotFoundError.
func IsReleaseLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == releaseLabelNotFoundError
}

var azureOperatorVersionLabelNotFoundError = &microerror.Error{
	Kind: "azureOperatorVersionLabelNotFoundError",
}

// IsReleaseLabelNotFoundError asserts azureOperatorVersionLabelNotFoundError.
func IsAzureOperatorVersionLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == azureOperatorVersionLabelNotFoundError
}

var componentNotFoundInReleaseError = &microerror.Error{
	Kind: "componentNotFoundInReleaseError",
}

// IsComponentNotFoundInReleaseError asserts componentNotFoundInReleaseError.
func IsComponentNotFoundInReleaseError(err error) bool {
	return microerror.Cause(err) == componentNotFoundInReleaseError
}
