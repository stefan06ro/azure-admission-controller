package azuremachinepool

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool

func AcceleratedNetworking(acceleratedNetworking *bool) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Spec.Template.AcceleratedNetworking = acceleratedNetworking
		return azureMachinePool
	}
}

func Cluster(clusterName string) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Labels[capiv1alpha3.ClusterLabelName] = clusterName
		return azureMachinePool
	}
}

func DataDisks(dataDisks []capzv1alpha3.DataDisk) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Spec.Template.DataDisks = dataDisks
		return azureMachinePool
	}
}

func Location(location string) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Spec.Location = location
		return azureMachinePool
	}
}

func Organization(org string) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Labels[label.Organization] = org
		azureMachinePool.Namespace = fmt.Sprintf("org-%s", org)
		return azureMachinePool
	}
}

func Name(name string) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.ObjectMeta.Name = name
		azureMachinePool.Labels[label.MachinePool] = name
		return azureMachinePool
	}
}

func StorageAccountType(storageAccountType compute.StorageAccountTypes) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Spec.Template.OSDisk.ManagedDisk.StorageAccountType = string(storageAccountType)
		return azureMachinePool
	}
}

func VMSize(vmsize string) BuilderOption {
	return func(azureMachinePool *expcapzv1alpha3.AzureMachinePool) *expcapzv1alpha3.AzureMachinePool {
		azureMachinePool.Spec.Template.VMSize = vmsize
		return azureMachinePool
	}
}

func BuildAzureMachinePool(opts ...BuilderOption) *expcapzv1alpha3.AzureMachinePool {
	nodepoolName := test.GenerateName()
	azureMachinePool := &expcapzv1alpha3.AzureMachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodepoolName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion:    "5.0.0",
				label.Cluster:                 "ab123",
				capiv1alpha3.ClusterLabelName: "ab123",
				label.MachinePool:             nodepoolName,
				label.Organization:            "giantswarm",
				label.ReleaseVersion:          "13.0.0",
			},
		},
		Spec: expcapzv1alpha3.AzureMachinePoolSpec{
			Location: "westeurope",
			Template: expcapzv1alpha3.AzureMachineTemplate{
				VMSize: "Standard_D4_v3",
				OSDisk: capzv1alpha3.OSDisk{
					ManagedDisk: capzv1alpha3.ManagedDisk{
						StorageAccountType: "Standard_LRS",
					},
				},
				AcceleratedNetworking: to.BoolPtr(true),
				DataDisks: []capzv1alpha3.DataDisk{
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
