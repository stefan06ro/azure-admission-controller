package unittest

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

type ResourceSkuStubAPI struct {
	stubbedSKUs map[string]compute.ResourceSku
}

func NewEmptyResourceSkuStubAPI() vmcapabilities.API {
	return &ResourceSkuStubAPI{stubbedSKUs: map[string]compute.ResourceSku{}}
}

func NewResourceSkuStubAPI(stubbedSKUs map[string]compute.ResourceSku) vmcapabilities.API {
	return &ResourceSkuStubAPI{stubbedSKUs: stubbedSKUs}
}

func (s *ResourceSkuStubAPI) List(_ context.Context, _ string) (map[string]compute.ResourceSku, error) {
	return s.stubbedSKUs, nil
}
