package vmcapabilities

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
)

type API interface {
	List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error)
}
