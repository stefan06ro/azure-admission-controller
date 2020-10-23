package azurecluster

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestAzureClusterCreateMutate(t *testing.T) {
	type testCase struct {
		name         string
		azureCluster []byte
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         fmt.Sprintf("case 0: ControlPlaneEndpoint left empty"),
			azureCluster: azureClusterRawObject("ab132", "", 0),
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
			azureCluster: azureClusterRawObject("ab132", "api.giantswarm.io", 123),
			patches:      []mutator.PatchOperation{},
			errorMatcher: nil,
		},
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
			patches, err := admit.Mutate(context.Background(), getCreateMutateAdmissionRequest(tc.azureCluster))

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
			Version:  "infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azurecluster",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}

func azureClusterRawObject(clusterName string, controlPlaneEndpointHost string, controlPlaneEndpointPort int32) []byte {
	mp := capzv1alpha3.AzureCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"cluster.x-k8s.io/cluster-name":        clusterName,
				"giantswarm.io/cluster":                clusterName,
				"giantswarm.io/organization":           "org-giantswarm",
				"release.giantswarm.io/version":        "13.0.0-alpha3",
			},
		},
		Spec: capzv1alpha3.AzureClusterSpec{
			ResourceGroup: clusterName,
			Location:      "westeurope",
			ControlPlaneEndpoint: v1alpha3.APIEndpoint{
				Host: controlPlaneEndpointHost,
				Port: controlPlaneEndpointPort,
			},
		},
	}
	byt, _ := json.Marshal(mp)
	return byt
}
