package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
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

	err = generic.ValidateOrganizationLabelContainsExistingOrganization(ctx, h.ctrlClient, azureClusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpoint(*azureClusterCR, h.baseDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateLocation(*azureClusterCR, h.location)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
