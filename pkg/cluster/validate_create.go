package cluster

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func (h *WebhookHandler) OnCreateValidate(ctx context.Context, object interface{}) error {
	clusterCR, err := key.ToClusterPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = clusterCR.ValidateCreate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelContainsExistingOrganization(ctx, h.ctrlClient, clusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateClusterNetwork(*clusterCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpoint(*clusterCR, h.baseDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
