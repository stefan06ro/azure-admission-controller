package azuremachinepool

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool

func AcceleratedNetworking(acceleratedNetworking *bool) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Template.AcceleratedNetworking = acceleratedNetworking
		return azureMachinePool
	}
}

func Cluster(clusterName string) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Labels[capi.ClusterLabelName] = clusterName
		return azureMachinePool
	}
}

func DataDisks(dataDisks []capz.DataDisk) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Template.DataDisks = dataDisks
		return azureMachinePool
	}
}

func Location(location string) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Location = location
		return azureMachinePool
	}
}

func Organization(org string) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Labels[label.Organization] = org
		azureMachinePool.Namespace = fmt.Sprintf("org-%s", org)
		return azureMachinePool
	}
}

func Name(name string) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.ObjectMeta.Name = name
		azureMachinePool.Labels[label.MachinePool] = name
		return azureMachinePool
	}
}

func SpotVMOptions(opts *capz.SpotVMOptions) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Template.SpotVMOptions = opts
		return azureMachinePool
	}
}

func StorageAccountType(storageAccountType compute.StorageAccountTypes) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Template.OSDisk.ManagedDisk.StorageAccountType = string(storageAccountType)
		return azureMachinePool
	}
}

func VMSize(vmsize string) BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		azureMachinePool.Spec.Template.VMSize = vmsize
		return azureMachinePool
	}
}

func WithDeletionTimestamp() BuilderOption {
	return func(azureMachinePool *capzexp.AzureMachinePool) *capzexp.AzureMachinePool {
		now := metav1.Now()
		azureMachinePool.ObjectMeta.SetDeletionTimestamp(&now)
		return azureMachinePool
	}
}

func BuildAzureMachinePool(opts ...BuilderOption) *capzexp.AzureMachinePool {
	nodepoolName := test.GenerateName()
	azureMachinePool := &capzexp.AzureMachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodepoolName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion: "5.0.0",
				label.Cluster:              "ab123",
				capi.ClusterLabelName:      "ab123",
				label.MachinePool:          nodepoolName,
				label.Organization:         "giantswarm",
				label.ReleaseVersion:       "13.0.0",
			},
		},
		Spec: capzexp.AzureMachinePoolSpec{
			Location: "westeurope",
			Template: capzexp.AzureMachineTemplate{
				VMSize: "Standard_D4_v3",
				OSDisk: capz.OSDisk{
					ManagedDisk: capz.ManagedDisk{
						StorageAccountType: "Standard_LRS",
					},
				},
				AcceleratedNetworking: to.BoolPtr(true),
				DataDisks: []capz.DataDisk{
					{
						NameSuffix: "docker",
						DiskSizeGB: 100,
						Lun:        to.Int32Ptr(21),
					},
					{
						NameSuffix: "kubelet",
						DiskSizeGB: 100,
						Lun:        to.Int32Ptr(22),
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(azureMachinePool)
	}

	return azureMachinePool
}

func BuildAzureMachinePoolAsJson(opts ...BuilderOption) []byte {
	azureMachinePool := BuildAzureMachinePool(opts...)

	byt, _ := json.Marshal(azureMachinePool)

	return byt
}
