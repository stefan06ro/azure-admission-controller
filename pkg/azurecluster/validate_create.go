package azurecluster

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type Validator struct {
	baseDomain string
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
}

type ValidatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
}

func NewValidator(config ValidatorConfig) (*Validator, error) {
	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Location == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Location must not be empty", config)
	}

	v := &Validator{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *Validator) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, azureClusterCR); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	return azureClusterCR, nil
}

func (a *Validator) Validate(ctx context.Context, object interface{}) error {
	azureClusterCR, err := key.ToAzureClusterPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = azureClusterCR.ValidateCreate()
	err = errors.IgnoreCAPIErrorForField("metadata.Name", err)
	err = errors.IgnoreCAPIErrorForField("spec.networkSpec.subnets", err)
	err = errors.IgnoreCAPIErrorForField("spec.SubscriptionID", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelContainsExistingOrganization(ctx, a.ctrlClient, azureClusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpoint(*azureClusterCR, a.baseDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateLocation(*azureClusterCR, a.location)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Validator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
