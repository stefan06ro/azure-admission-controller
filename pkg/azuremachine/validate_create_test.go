package azuremachine

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachineCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		azureMachine []byte
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "Case 0 - empty ssh key",
			azureMachine: azureMachineRawObject("", "westeurope", nil, nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 1 - not empty ssh key",
			azureMachine: azureMachineRawObject("ssh-rsa 12345 giantswarm", "westeurope", nil, nil),
			errorMatcher: IsSSHFieldIsSetError,
		},
		{
			name:         "Case 2 - invalid location",
			azureMachine: azureMachineRawObject("", "westpoland", nil, nil),
			errorMatcher: IsUnexpectedLocationError,
		},
		{
			name:         "Case 3 - invalid failure domain",
			azureMachine: azureMachineRawObject("", "westeurope", to.StringPtr("2"), nil),
			errorMatcher: IsUnsupportedFailureDomainError,
		},
		{
			name:         "Case 4 - valid failure domain",
			azureMachine: azureMachineRawObject("", "westeurope", to.StringPtr("1"), nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 5 - empty failure domain",
			azureMachine: azureMachineRawObject("", "westeurope", to.StringPtr(""), nil),
			errorMatcher: nil,
		},
		{
			name:         "Case 6 - nil failure domain",
			azureMachine: azureMachineRawObject("", "westeurope", nil, nil),
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
				panic(microerror.JSON(err))
			}

			admit := &CreateValidator{
				ctrlClient: ctrlClient,
				location:   "westeurope",
				logger:     newLogger,
				vmcaps:     vmcaps,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(ctx, getCreateAdmissionRequest(tc.azureMachine))

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
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
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
