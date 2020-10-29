package azurecluster

import (
	"context"
	"testing"

	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureClusterCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		azureCluster []byte
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: empty ControlPlaneEndpoint",
			azureCluster: azureClusterRawObject("ab123", "", 0),
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 1: Invalid Port",
			azureCluster: azureClusterRawObject("ab123", "api.ab123.k8s.test.westeurope.azure.gigantic.io", 80),
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 2: Invalid Host",
			azureCluster: azureClusterRawObject("ab123", "api.gigantic.io", 443),
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 3: Valid values",
			azureCluster: azureClusterRawObject("ab123", "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443),
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

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// Create default GiantSwarm organization.
			organization := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
				Spec: securityv1alpha1.OrganizationSpec{},
			}
			err = ctrlClient.Create(ctx, organization)
			if err != nil {
				t.Fatal(err)
			}

			admit := &CreateValidator{
				baseDomain: "k8s.test.westeurope.azure.gigantic.io",
				ctrlClient: ctrlClient,
				logger:     newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(ctx, getCreateAdmissionRequest(tc.azureCluster))

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
