package azuremachinepool

import (
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/test"
)

type BuilderOption func(machinePool *capiexp.MachinePool) *capiexp.MachinePool

func AzureMachinePool(azureMachinePoolName string) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		machinePool.Spec.Template.Spec.InfrastructureRef.Name = azureMachinePoolName
		return machinePool
	}
}

func FailureDomains(failureDomains []string) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		machinePool.Spec.FailureDomains = failureDomains
		return machinePool
	}
}

func Name(name string) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		machinePool.ObjectMeta.Name = name
		machinePool.Labels[label.MachinePool] = name
		return machinePool
	}
}

func Organization(org string) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		namespace := fmt.Sprintf("org-%s", org)
		machinePool.Labels[label.Organization] = org
		machinePool.Namespace = namespace
		machinePool.Spec.Template.Spec.InfrastructureRef.Namespace = namespace
		machinePool.Spec.Template.Spec.Bootstrap.ConfigRef.Namespace = namespace
		return machinePool
	}
}

func Replicas(replicas int32) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		machinePool.Spec.Replicas = &replicas
		machinePool.Annotations[annotation.NodePoolMinSize] = fmt.Sprintf("%d", replicas)
		machinePool.Annotations[annotation.NodePoolMaxSize] = fmt.Sprintf("%d", replicas)
		return machinePool
	}
}

func Annotation(name, val string) BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		machinePool.Annotations[name] = val
		return machinePool
	}
}

func WithDeletionTimestamp() BuilderOption {
	return func(machinePool *capiexp.MachinePool) *capiexp.MachinePool {
		now := metav1.Now()
		machinePool.ObjectMeta.SetDeletionTimestamp(&now)
		return machinePool
	}
}

func BuildMachinePool(opts ...BuilderOption) *capiexp.MachinePool {
	nodepoolName := test.GenerateName()
	machinePool := &capiexp.MachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodepoolName,
			Namespace: "org-giantswarm",
			Annotations: map[string]string{
				annotation.NodePoolMinSize: "1",
				annotation.NodePoolMaxSize: "1",
			},
			Labels: map[string]string{
				label.AzureOperatorVersion: "5.0.0",
				label.Cluster:              "ab123",
				capi.ClusterLabelName:      "ab123",
				label.MachinePool:          nodepoolName,
				label.Organization:         "giantswarm",
				label.ReleaseVersion:       "13.0.0",
			},
		},
		Spec: capiexp.MachinePoolSpec{
			FailureDomains: []string{},
			Template: capi.MachineTemplateSpec{
				Spec: capi.MachineSpec{
					Bootstrap: capi.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Namespace: "org-giantswarm",
							Name:      "ab123",
						},
					},
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
