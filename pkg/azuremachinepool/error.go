package azuremachinepool

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

var vmsizeDoesNotSupportAcceleratedNetworkingError = &microerror.Error{
	Kind: "vmsizeDoesNotSupportAcceleratedNetworkingError",
}

// IsVmsizeDoesNotSupportAcceleratedNetworkingError asserts vmsizeDoesNotSupportAcceleratedNetworkingError.
func IsVmsizeDoesNotSupportAcceleratedNetworkingError(err error) bool {
	return microerror.Cause(err) == vmsizeDoesNotSupportAcceleratedNetworkingError
}

var datadisksFieldIsSetError = &microerror.Error{
	Kind: "datadisksFieldIsSetError",
}

// IsDatadisksFieldIsSetError asserts datadisksFieldIsSetError.
func IsDatadisksFieldIsSetError(err error) bool {
	return microerror.Cause(err) == datadisksFieldIsSetError
}

var locationWasChangedError = &microerror.Error{
	Kind: "locationWasChangedError",
}

// IsLocationWasChangedError asserts locationWasChangedError.
func IsLocationWasChangedError(err error) bool {
	return microerror.Cause(err) == locationWasChangedError
}

var acceleratedNetworkingWasChangedError = &microerror.Error{
	Kind: "acceleratedNetworkingWasChangedError",
}

// IsAcceleratedNetworkingWasChangedError asserts acceleratedNetworkingWasChangedError.
func IsAcceleratedNetworkingWasChangedError(err error) bool {
	return microerror.Cause(err) == acceleratedNetworkingWasChangedError
}

var storageAccountWasChangedError = &microerror.Error{
	Kind: "storageAccountWasChangedError",
}

// IsStorageAccountWasChangedError asserts storageAccountWasChangedError.
func IsStorageAccountWasChangedError(err error) bool {
	return microerror.Cause(err) == storageAccountWasChangedError
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

var insufficientMemoryError = &microerror.Error{
	Kind: "insufficientMemoryError",
}

// IsInsufficientMemoryError asserts insufficientMemoryError.
func IsInsufficientMemoryError(err error) bool {
	return microerror.Cause(err) == insufficientMemoryError
}

var insufficientCPUError = &microerror.Error{
	Kind: "insufficientCPUError",
}

// IsInsufficientCPUError asserts insufficientCPUError.
func IsInsufficientCPUError(err error) bool {
	return microerror.Cause(err) == insufficientCPUError
}

var switchToVmSizeThatDoesNotSupportAcceleratedNetworkingError = &microerror.Error{
	Kind: "switchToVmSizeThatDoesNotSupportAcceleratedNetworkingError",
}

// IsSwitchToVmSizeThatDoesNotSupportAcceleratedNetworkingError asserts switchToVmSizeThatDoesNotSupportAcceleratedNetworkingError.
func IsSwitchToVmSizeThatDoesNotSupportAcceleratedNetworkingError(err error) bool {
	return microerror.Cause(err) == switchToVmSizeThatDoesNotSupportAcceleratedNetworkingError
}

var premiumStorageNotSupportedByVMSizeError = &microerror.Error{
	Kind: "premiumStorageNotSupportedByVMSizeError",
}

// IsPremiumStorageNotSupportedByVMSizeError asserts premiumStorageNotSupportedByVMSizeError.
func IsPremiumStorageNotSupportedByVMSizeError(err error) bool {
	return microerror.Cause(err) == premiumStorageNotSupportedByVMSizeError
}

var invalidStorageAccountTypeError = &microerror.Error{
	Kind: "invalidStorageAccountTypeError",
}

// IsInvalidStorageAccountTypeError asserts invalidStorageAccountTypeError.
func IsInvalidStorageAccountTypeError(err error) bool {
	return microerror.Cause(err) == invalidStorageAccountTypeError
}
