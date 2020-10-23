package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestClusterCreateMutate(t *testing.T) {
	type testCase struct {
		name         string
		cluster      []byte
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	clusterNetwork := &v1alpha3.ClusterNetwork{
		APIServerPort: to.Int32Ptr(443),
		ServiceDomain: "ab123.k8s.test.westeurope.azure.gigantic.io",
		Services: &v1alpha3.NetworkRanges{
			CIDRBlocks: []string{
				"172.31.0.0/16",
			},
		},
	}

	testCases := []testCase{
		{
			name:    fmt.Sprintf("case 0: ControlPlaneEndpoint left empty"),
			cluster: clusterRawObject("ab132", clusterNetwork, "", 0),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/controlPlaneEndpoint/host",
					Value:     "api.ab132.k8s.test.westeurope.azure.gigantic.io",
				},
				{
					Operation: "add",
					Path:      "/spec/controlPlaneEndpoint/port",
					Value:     443,
				},
			},
			errorMatcher: nil,
		},
		{
			name:         fmt.Sprintf("case 1: ControlPlaneEndpoint has a value"),
			cluster:      clusterRawObject("ab132", clusterNetwork, "api.giantswarm.io", 123),
			patches:      []mutator.PatchOperation{},
			errorMatcher: nil,
		},
		// This test doesn't work because the clusterNetwork struct has pointers in it and we can't compare them.
		//{
		//	name:         fmt.Sprintf("case 2: ClusterNetwork empty"),
		//	cluster:      clusterRawObject("ab132", nil, "api.giantswarm.io", 123),
		//	patches:      []mutator.PatchOperation{
		//		{
		//			Operation: "add",
		//			Path:      "/spec/clusterNetwork",
		//			Value:     *clusterNetwork,
		//		},
		//	},
		//	errorMatcher: nil,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}

			admit := &CreateMutator{
				baseDomain: "k8s.test.westeurope.azure.gigantic.io",
				logger:     newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			patches, err := admit.Mutate(context.Background(), getCreateMutateAdmissionRequest(tc.cluster))

			// Check if the error is the expected one.
			switch {
			case err == nil && tc.errorMatcher == nil:
				// fall through
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("expected %#v got %#v", nil, err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("expected %#v got %#v", "error", nil)
			case !tc.errorMatcher(err):
				t.Fatalf("unexpected error: %#v", err)
			}

			// Check if the validation result is the expected one.
			if len(tc.patches) != 0 || len(patches) != 0 {
				if !reflect.DeepEqual(tc.patches, patches) {
					t.Fatalf("Patches mismatch: expected %v, got %v", tc.patches, patches)
				}
			}
		})
	}
}

func getCreateMutateAdmissionRequest(newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "cluster.x-k8s.io/v1alpha3",
			Resource: "cluster",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}

func clusterRawObject(clusterName string, clusterNetwork *v1alpha3.ClusterNetwork, controlPlaneEndpointHost string, controlPlaneEndpointPort int32) []byte {
	mp := v1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version":   "5.0.0",
				"cluster-operator.giantswarm.io/version": "",
				"cluster.x-k8s.io/cluster-name":          clusterName,
				"giantswarm.io/cluster":                  clusterName,
				"giantswarm.io/organization":             "org-giantswarm",
				"release.giantswarm.io/version":          "13.0.0-alpha3",
			},
		},
		Spec: v1alpha3.ClusterSpec{
			ClusterNetwork: clusterNetwork,
			ControlPlaneEndpoint: v1alpha3.APIEndpoint{
				Host: controlPlaneEndpointHost,
				Port: controlPlaneEndpointPort,
			},
		},
	}
	byt, _ := json.Marshal(mp)
	return byt
}
