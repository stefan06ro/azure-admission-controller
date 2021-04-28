package azurecluster

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureClusterUpdateValidate(t *testing.T) {
	type testCase struct {
		name            string
		oldAzureCluster *capz.AzureCluster
		newAzureCluster *capz.AzureCluster
		errorMatcher    func(err error) bool
	}

	testCases := []testCase{
		{
			name:            "case 0: unchanged ControlPlaneEndpoint",
			oldAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 443)),
			newAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 443)),
			errorMatcher:    nil,
		},
		{
			name:            "case 1: host changed",
			oldAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 443)),
			newAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.azure.gigantic.io", 443)),
			errorMatcher:    IsControlPlaneEndpointWasChangedError,
		},
		{
			name:            "case 2: port changed",
			oldAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 443)),
			newAzureCluster: builder.BuildAzureCluster(builder.Name("ab123"), builder.ControlPlaneEndpoint("api.ab123.k8s.test.westeurope.azure.gigantic.io", 80)),
			errorMatcher:    IsControlPlaneEndpointWasChangedError,
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

			admit := &Validator{
				baseDomain: "k8s.test.westeurope.azure.gigantic.io",
				ctrlClient: ctrlClient,
				location:   "westeurope",
				logger:     newLogger,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.OnUpdateValidate(ctx, tc.oldAzureCluster, tc.newAzureCluster)

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
