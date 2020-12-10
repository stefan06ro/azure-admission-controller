package azureupdate

import (
	"context"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
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

func (a *AzureConfigValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	azureConfigNewCR := &v1alpha1.AzureConfig{}
	azureConfigOldCR := &v1alpha1.AzureConfig{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureConfigNewCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse azureConfig CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureConfigOldCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse azureConfig CR: %v", err)
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
