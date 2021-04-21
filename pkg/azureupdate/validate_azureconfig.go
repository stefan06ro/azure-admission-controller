package azureupdate

import (
	"context"
	"reflect"
	"sort"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type AzureConfigValidator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type AzureConfigValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewAzureConfigValidator(config AzureConfigValidatorConfig) (*AzureConfigValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	admitter := &AzureConfigValidator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return admitter, nil
}

func (a *AzureConfigValidator) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	cr := &v1alpha1.AzureConfig{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, cr); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureConfig CR: %v", err)
	}

	return cr, nil
}

func (a *AzureConfigValidator) ValidateUpdate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureConfigNewCR, err := key.ToAzureConfigPtr(object)
	if err != nil {
		return microerror.Mask(err)
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
		return releaseversion.Validate(ctx, a.ctrlClient, oldVersion, newVersion)
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

func (a *AzureConfigValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
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
