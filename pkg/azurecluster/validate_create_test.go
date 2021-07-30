package azurecluster

import (
	"context"
	"testing"

	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureClusterCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		azureCluster *capz.AzureCluster
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: empty ControlPlaneEndpoint",
			azureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("", 443)),
			errorMatcher: IsInvalidControlPlaneEndpointHostError,
		},
		{
			name:         "case 1: Invalid Port",
			azureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 80)),
			errorMatcher: IsInvalidControlPlaneEndpointPortError,
		},
		{
			name:         "case 2: Invalid Host",
			azureCluster: builder.BuildAzureCluster(builder.ControlPlaneEndpoint("api.gigantic.io", 443), builder.Location("westeurope")),
			errorMatcher: IsInvalidControlPlaneEndpointHostError,
		},
		{
			name:         "case 3: Valid values",
			azureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 443), builder.Location("westeurope")),
			errorMatcher: nil,
		},
		{
			name:         "case 4: Invalid region",
			azureCluster: builder.BuildAzureCluster(builder.Location("westpoland")),
			errorMatcher: IsUnexpectedLocationError,
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

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				BaseDomain: "k8s.test.westeurope.azure.gigantic.io",
				CtrlReader: ctrlClient,
				CtrlClient: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Location:   "westeurope",
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on AzureCluster creation.
			err = handler.OnCreateValidate(ctx, tc.azureCluster)

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
