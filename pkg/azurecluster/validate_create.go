package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type CreateValidator struct {
	baseDomain string
	ctrlClient client.Client
	location   string
	logger     micrologger.Logger
}

type CreateValidatorConfig struct {
	BaseDomain string
	CtrlClient client.Client
	Location   string
	Logger     micrologger.Logger
}

func NewCreateValidator(config CreateValidatorConfig) (*CreateValidator, error) {
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

	v := &CreateValidator{
		baseDomain: config.BaseDomain,
		ctrlClient: config.CtrlClient,
		location:   config.Location,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *CreateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureClusterCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	err := azureClusterCR.ValidateCreate()
	err = errors.IgnoreCAPIErrorForField("metadata.Name", err)
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

func (a *CreateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}
