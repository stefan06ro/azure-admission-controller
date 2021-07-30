package azuremachine

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachineUpdateValidate(t *testing.T) {
	type testCase struct {
		name         string
		oldAM        *capz.AzureMachine
		newAM        *capz.AzureMachine
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "Case 0 - empty ssh key",
			oldAM:        azureMachineObject("", "westeurope", nil, nil),
			newAM:        azureMachineObject("", "westeurope", nil, nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 1 - not empty ssh key",
			oldAM:        azureMachineObject("", "westeurope", nil, nil),
			newAM:        azureMachineObject("ssh-rsa 12345 giantswarm", "westeurope", nil, nil),
			errorMatcher: IsSSHFieldIsSetError,
		},
		{
			name:         "Case 2 - location changed",
			oldAM:        azureMachineObject("", "westeurope", nil, nil),
			newAM:        azureMachineObject("", "westpoland", nil, nil),
			errorMatcher: IsLocationWasChangedError,
		},
		{
			name:         "Case 3 - failure domain changed",
			oldAM:        azureMachineObject("", "westpoland", to.StringPtr("1"), nil),
			newAM:        azureMachineObject("", "westpoland", to.StringPtr("2"), nil),
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

			stubAPI := unittest.NewEmptyResourceSkuStubAPI()
			vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
				Azure:  stubAPI,
				Logger: newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				CtrlClient: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Location:   "westeurope",
				Logger:     newLogger,
				VMcaps:     vmcaps,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run admission request to validate AzureConfig updates.
			err = handler.OnUpdateValidate(ctx, tc.oldAM, tc.newAM)

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
