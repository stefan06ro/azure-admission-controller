package azurecluster

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func TestAzureClusterCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		azureCluster []byte
		allowed      bool
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: empty ControlPlaneEndpoint",
			azureCluster: azureClusterRawObject("ab123", "", 0),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 1: Invalid Port",
			azureCluster: azureClusterRawObject("ab123", "api.ab123.k8s.test.westeurope.azure.gigantic.io", 80),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 2: Invalid Host",
			azureCluster: azureClusterRawObject("ab123", "api.gigantic.io", 443),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 3: Valid values",
			azureCluster: azureClusterRawObject("ab123", "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443),
			allowed:      true,
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

			admit := &CreateValidator{
				baseDomain: "k8s.test.westeurope.azure.gigantic.io",
				logger:     newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			allowed, err := admit.Validate(context.Background(), getCreateAdmissionRequest(tc.azureCluster))

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
			if tc.allowed != allowed {
				t.Fatalf("expected %v to be equal to %v", tc.allowed, allowed)
			}
		})
	}
}

func getCreateAdmissionRequest(newMP []byte) *v1beta1.AdmissionRequest {
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
