package azureupdate

import (
	"context"

	"github.com/blang/semver"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type AzureClusterConfigValidator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type AzureClusterConfigValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewAzureClusterConfigValidator(config AzureClusterConfigValidatorConfig) (*AzureClusterConfigValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	azureClusterValidator := &AzureClusterConfigValidator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return azureClusterValidator, nil
}

func (a *AzureClusterConfigValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	AzureClusterConfigNewCR := &corev1alpha1.AzureClusterConfig{}
	AzureClusterConfigOldCR := &corev1alpha1.AzureClusterConfig{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, AzureClusterConfigNewCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, AzureClusterConfigOldCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}

	if !AzureClusterConfigNewCR.GetDeletionTimestamp().IsZero() {
		a.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
	}

	oldVersion, err := getSemver(AzureClusterConfigOldCR.Spec.Guest.ReleaseVersion)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureClusterConfig (before edit)")
	}
	newVersion, err := getSemver(AzureClusterConfigNewCR.Spec.Guest.ReleaseVersion)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureClusterConfig (after edit)")
	}

	if !oldVersion.Equals(newVersion) {
		return releaseversion.Validate(ctx, a.ctrlClient, oldVersion, newVersion)
	}

	return nil
}

func (a *AzureClusterConfigValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func getSemver(version string) (semver.Version, error) {
	return semver.ParseTolerant(version)
}
