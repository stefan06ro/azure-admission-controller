package azureupdate

import (
	"context"

	"github.com/blang/semver"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

type AzureClusterConfigWebhookHandler struct {
	ctrlClient client.Client
	decoder    runtime.Decoder
	logger     micrologger.Logger
}

type AzureClusterConfigWebhookHandlerConfig struct {
	CtrlClient client.Client
	Decoder    runtime.Decoder
	Logger     micrologger.Logger
}

func NewAzureClusterConfigWebhookHandler(config AzureClusterConfigWebhookHandlerConfig) (*AzureClusterConfigWebhookHandler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Decoder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Decoder must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	webhookHandler := &AzureClusterConfigWebhookHandler{
		ctrlClient: config.CtrlClient,
		decoder:    config.Decoder,
		logger:     config.Logger,
	}

	return webhookHandler, nil
}

func (h *AzureClusterConfigWebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &corev1alpha1.AzureClusterConfig{}
	if _, _, err := h.decoder.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}

	return cr, nil
}

func (h *AzureClusterConfigWebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureClusterConfigNewCR, err := key.ToAzureClusterConfigPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	if !azureClusterConfigNewCR.GetDeletionTimestamp().IsZero() {
		h.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
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
		return releaseversion.Validate(ctx, h.ctrlClient, oldVersion, newVersion)
	}

	return nil
}

func (h *AzureClusterConfigWebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *AzureClusterConfigWebhookHandler) Resource() string {
	return "azureclusterconfig"
}

func getSemver(version string) (semver.Version, error) {
	return semver.ParseTolerant(version)
}
