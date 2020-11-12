package releaseversion

import (
	"github.com/giantswarm/microerror"
)

var releaseNotFoundError = &microerror.Error{
	Kind: "downgradingIsNotAllowedError",
}

// IsReleaseNotFoundError asserts releaseNotFoundError.
func IsReleaseNotFoundError(err error) bool {
	return microerror.Cause(err) == releaseNotFoundError
}

var downgradingIsNotAllowedError = &microerror.Error{
	Kind: "downgradingIsNotAllowedError",
}

// IsDowngradingIsNotAllowedError asserts downgradingIsNotAllowedError.
func IsDowngradingIsNotAllowedError(err error) bool {
	return microerror.Cause(err) == downgradingIsNotAllowedError
}

var upgradingToOrFromAlphaReleaseError = &microerror.Error{
	Kind: "upgradingToOrFromAlphaReleaseError",
}

// IsUpgradingToOrFromAlphaReleaseError asserts upgradingToOrFromAlphaReleaseError.
func IsUpgradingToOrFromAlphaReleaseError(err error) bool {
	return microerror.Cause(err) == upgradingToOrFromAlphaReleaseError
}

var skippingReleaseError = &microerror.Error{
	Kind: "skippingReleaseError",
}

// IsSkippingReleaseError asserts skippingReleaseError.
func IsSkippingReleaseError(err error) bool {
	return microerror.Cause(err) == skippingReleaseError
}
