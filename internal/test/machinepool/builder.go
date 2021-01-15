package azuremachinepool

import (
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool

func AzureMachinePool(azureMachinePoolName string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Spec.Template.Spec.InfrastructureRef.Name = azureMachinePoolName
		return machinePool
	}
}

func FailureDomains(failureDomains []string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Spec.FailureDomains = failureDomains
		return machinePool
	}
}

func Name(name string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.ObjectMeta.Name = name
		machinePool.Labels[label.MachinePool] = name
		return machinePool
	}
}

func Organization(org string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Labels[label.Organization] = org
		machinePool.Namespace = fmt.Sprintf("org-%s", org)
		return machinePool
	}
}

func Replicas(replicas int32) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Spec.Replicas = &replicas
		machinePool.Annotations[annotation.NodePoolMinSize] = fmt.Sprintf("%d", replicas)
		machinePool.Annotations[annotation.NodePoolMaxSize] = fmt.Sprintf("%d", replicas)
		return machinePool
	}
}

func Annotation(name, val string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Annotations[name] = val
		return machinePool
	}
}

func BuildMachinePool(opts ...BuilderOption) *expcapiv1alpha3.MachinePool {
	nodepoolName := test.GenerateName()
	machinePool := &expcapiv1alpha3.MachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodepoolName,
			Namespace: "org-giantswarm",
			Annotations: map[string]string{
				annotation.NodePoolMinSize: "1",
				annotation.NodePoolMaxSize: "1",
			},
			Labels: map[string]string{
				label.AzureOperatorVersion:    "5.0.0",
				label.Cluster:                 "ab123",
				capiv1alpha3.ClusterLabelName: "ab123",
				label.MachinePool:             nodepoolName,
				label.Organization:            "giantswarm",
				label.ReleaseVersion:          "13.0.0",
			},
		},
		Spec: expcapiv1alpha3.MachinePoolSpec{
			FailureDomains: []string{},
			Template: v1alpha3.MachineTemplateSpec{
				Spec: v1alpha3.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Namespace: "org-giantswarm",
						Name:      "ab123",
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(machinePool)
	}

	return machinePool
}

func BuildMachinePoolAsJson(opts ...BuilderOption) []byte {
	machinePool := BuildMachinePool(opts...)

	byt, _ := json.Marshal(machinePool)

	return byt
}
