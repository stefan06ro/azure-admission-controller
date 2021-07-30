package azureupdate

import (
	"context"
	"reflect"
	"sort"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

type AzureConfigWebhookHandler struct {
	ctrlClient client.Client
	decoder    runtime.Decoder
	logger     micrologger.Logger
}

type AzureConfigWebhookHandlerConfig struct {
	CtrlClient client.Client
	Decoder    runtime.Decoder
	Logger     micrologger.Logger
}

func NewAzureConfigWebhookHandler(config AzureConfigWebhookHandlerConfig) (*AzureConfigWebhookHandler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Decoder == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Decoder must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	webhookHandler := &AzureConfigWebhookHandler{
		ctrlClient: config.CtrlClient,
		decoder:    config.Decoder,
		logger:     config.Logger,
	}

	return webhookHandler, nil
}

func (h *AzureConfigWebhookHandler) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &v1alpha1.AzureConfig{}
	if _, _, err := h.decoder.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureConfig CR: %v", err)
	}

	return cr, nil
}

func (h *AzureConfigWebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureConfigNewCR, err := key.ToAzureConfigPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	if !azureConfigNewCR.GetDeletionTimestamp().IsZero() {
		h.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
	}

	azureConfigOldCR, err := key.ToAzureConfigPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	oldVersion, err := semverhelper.GetSemverFromLabels(azureConfigOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newVersion, err := semverhelper.GetSemverFromLabels(azureConfigNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	if !oldVersion.Equals(newVersion) {
		return releaseversion.Validate(ctx, h.ctrlClient, oldVersion, newVersion)
	}

	// Don't allow change of Master CIDR.
	err = validateMasterCIDRUnchanged(azureConfigOldCR, azureConfigNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	// Don't allow change of Availability Zones.
	err = validateAvailabilityZonesUnchanged(azureConfigOldCR, azureConfigNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *AzureConfigWebhookHandler) Log(keyVals ...interface{}) {
	h.logger.Log(keyVals...)
}

func (h *AzureConfigWebhookHandler) Resource() string {
	return "azureconfig"
}

func validateMasterCIDRUnchanged(old *v1alpha1.AzureConfig, new *v1alpha1.AzureConfig) error {
	if old.Spec.Azure.VirtualNetwork.MasterSubnetCIDR != "" && old.Spec.Azure.VirtualNetwork.MasterSubnetCIDR != new.Spec.Azure.VirtualNetwork.MasterSubnetCIDR {
		return microerror.Maskf(masterCIDRChangeError, "Spec.Azure.VirtualNetwork.MasterSubnetCIDR change disallowed")
	}

	return nil
}

func validateAvailabilityZonesUnchanged(old *v1alpha1.AzureConfig, new *v1alpha1.AzureConfig) error {
	if len(old.Spec.Azure.AvailabilityZones) != 0 {
		oldAZs := make([]int, len(old.Spec.Azure.AvailabilityZones))
		copy(oldAZs, old.Spec.Azure.AvailabilityZones)
		sort.Ints(oldAZs)

		newAZs := make([]int, len(new.Spec.Azure.AvailabilityZones))
		copy(newAZs, new.Spec.Azure.AvailabilityZones)
		sort.Ints(newAZs)

		if !reflect.DeepEqual(oldAZs, newAZs) {
			return microerror.Maskf(availabilityZonesChangeError, "Spec.Azure.AvailabilityZones change disallowed")
		}
	}

	return nil
}
