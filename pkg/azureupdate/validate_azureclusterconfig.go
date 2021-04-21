package azureupdate

import (
	"context"

	"github.com/blang/semver"
	corev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func (a *AzureClusterConfigValidator) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &corev1alpha1.AzureClusterConfig{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}

	return cr, nil
}

func (a *AzureClusterConfigValidator) ValidateUpdate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureClusterConfigNewCR, err := key.ToAzureClusterConfigPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	azureClusterConfigOldCR, err := key.ToAzureClusterConfigPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	oldVersion, err := getSemver(azureClusterConfigOldCR.Spec.Guest.ReleaseVersion)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureClusterConfig (before edit)")
	}
	newVersion, err := getSemver(azureClusterConfigNewCR.Spec.Guest.ReleaseVersion)
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
