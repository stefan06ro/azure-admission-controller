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

func TestAzureClusterUpdateValidate(t *testing.T) {
	type testCase struct {
		name            string
		oldAzureCluster []byte
		newAzureCluster []byte
		errorMatcher    func(err error) bool
	}

	testCases := []testCase{
		{
			name:            "case 0: unchanged ControlPlaneEndpoint",
			oldAzureCluster: azureClusterRawObject("ab123", "api.ab123.test.westeurope.azure.gigantic.io", 443, "westeurope"),
			newAzureCluster: azureClusterRawObject("ab123", "api.ab123.test.westeurope.azure.gigantic.io", 443, "westeurope"),
			errorMatcher:    nil,
		},
		{
			name:            "case 1: host changed",
			oldAzureCluster: azureClusterRawObject("ab123", "api.ab123.test.westeurope.azure.gigantic.io", 443, "westeurope"),
			newAzureCluster: azureClusterRawObject("ab123", "api.azure.gigantic.io", 443, "westeurope"),
			errorMatcher:    errors.IsInvalidOperationError,
		},
		{
			name:            "case 2: port changed",
			oldAzureCluster: azureClusterRawObject("ab123", "api.ab123.test.westeurope.azure.gigantic.io", 443, "westeurope"),
			newAzureCluster: azureClusterRawObject("ab123", "api.ab123.test.westeurope.azure.gigantic.io", 80, "westeurope"),
			errorMatcher:    errors.IsInvalidOperationError,
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
			err = admit.Validate(context.Background(), getUpdateAdmissionRequest(tc.oldAzureCluster, tc.newAzureCluster))

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

func getUpdateAdmissionRequest(oldCR []byte, newCR []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azurecluster",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    newCR,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    oldCR,
			Object: nil,
		},
	}

	return req
}
