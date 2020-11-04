package azuremachine

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestAzureMachineCreateMutate(t *testing.T) {
	type testCase struct {
		name         string
		azureMachine []byte
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         fmt.Sprintf("case 0: Location empty"),
			azureMachine: azureMachineRawObject("ab132", ""),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/location",
					Value:     "westeurope",
				},
			},
			errorMatcher: nil,
		},
		{
			name:         fmt.Sprintf("case 1: Location has value"),
			azureMachine: azureMachineRawObject("ab132", "westeurope"),
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
				location: "westeurope",
				logger:   newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			patches, err := admit.Mutate(context.Background(), getCreateMutateAdmissionRequest(tc.azureMachine))

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
			Resource: "azuremachine",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}
