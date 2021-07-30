package azuremachine

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachineCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		azureMachine *capz.AzureMachine
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "Case 0 - empty ssh key",
			azureMachine: azureMachineObject("", "westeurope", nil, nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 1 - not empty ssh key",
			azureMachine: azureMachineObject("ssh-rsa 12345 giantswarm", "westeurope", nil, nil),
			errorMatcher: IsSSHFieldIsSetError,
		},
		{
			name:         "Case 2 - invalid location",
			azureMachine: azureMachineObject("", "westpoland", nil, nil),
			errorMatcher: IsUnexpectedLocationError,
		},
		{
			name:         "Case 3 - invalid failure domain",
			azureMachine: azureMachineObject("", "westeurope", to.StringPtr("2"), nil),
			errorMatcher: IsUnsupportedFailureDomainError,
		},
		{
			name:         "Case 4 - valid failure domain",
			azureMachine: azureMachineObject("", "westeurope", to.StringPtr("1"), nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 5 - empty failure domain",
			azureMachine: azureMachineObject("", "westeurope", to.StringPtr(""), nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 6 - nil failure domain",
			azureMachine: azureMachineObject("", "westeurope", nil, nil),
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

			stubbedSKUs := map[string]compute.ResourceSku{
				"Standard_D4s_v3": {
					Name: to.StringPtr("Standard_D4s_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("False"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("True"),
						},
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Zones: &[]string{
								"1",
							},
						},
					},
				},
			}
			stubAPI := NewStubAPI(stubbedSKUs)
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

			// Run validating webhook handler on AzureMachine creation.
			err = handler.OnCreateValidate(ctx, tc.azureMachine)

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

type StubAPI struct {
	stubbedSKUs map[string]compute.ResourceSku
}

func NewStubAPI(stubbedSKUs map[string]compute.ResourceSku) vmcapabilities.API {
	return &StubAPI{stubbedSKUs: stubbedSKUs}
}

func (s *StubAPI) List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error) {
	return s.stubbedSKUs, nil
}
