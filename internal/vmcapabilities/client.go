package vmcapabilities

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/giantswarm/microerror"
)

type Azure struct {
	resourceSkuClient *compute.ResourceSkusClient
}

type AzureConfig struct {
	ResourceSkuClient *compute.ResourceSkusClient
}

func NewAzureAPI(c AzureConfig) API {
	return &Azure{resourceSkuClient: c.ResourceSkuClient}
}

func (a *Azure) List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error) {
	skus := map[string]compute.ResourceSku{}

	iterator, err := a.resourceSkuClient.ListComplete(ctx, filter)
	if err != nil {
		return skus, microerror.Mask(err)
	}

	for iterator.NotDone() {
		sku := iterator.Value()
		skus[*sku.Name] = sku

		err := iterator.NextWithContext(ctx)
		if err != nil {
			return skus, microerror.Mask(err)
		}
	}

	return skus, nil
}
