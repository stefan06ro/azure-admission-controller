package machinepool

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/machinepool"
)

func TestMachinePoolUpdateValidate(t *testing.T) {
	type testCase struct {
		name         string
		oldNodePool  []byte
		newNodePool  []byte
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: FailureDomains unchanged",
			oldNodePool:  builder.BuildMachinePoolAsJson(builder.FailureDomains([]string{"1", "2"})),
			newNodePool:  builder.BuildMachinePoolAsJson(builder.FailureDomains([]string{"1", "2"})),
			errorMatcher: nil,
		},
		{
			name:         "case 1: FailureDomains changed",
			oldNodePool:  builder.BuildMachinePoolAsJson(builder.FailureDomains([]string{"1"})),
			newNodePool:  builder.BuildMachinePoolAsJson(builder.FailureDomains([]string{"2"})),
			errorMatcher: IsFailureDomainWasChangedError,
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

			admit := &UpdateValidator{
				logger: newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(context.Background(), getUpdateAdmissionRequest(tc.oldNodePool, tc.newNodePool))

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
		})
	}
}

func getUpdateAdmissionRequest(oldMP []byte, newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    oldMP,
			Object: nil,
		},
	}

	return req
}
