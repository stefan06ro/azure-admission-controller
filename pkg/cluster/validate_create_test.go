package cluster

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestClusterCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		cluster      []byte
		errorMatcher func(err error) bool
	}

	clusterNetwork := &capi.ClusterNetwork{
		APIServerPort: to.Int32Ptr(443),
		ServiceDomain: "cluster.local",
		Services: &capi.NetworkRanges{
			CIDRBlocks: []string{
				"172.31.0.0/16",
			},
		},
	}

	testCases := []testCase{
		{
			name:         "case 0: empty ControlPlaneEndpoint",
			cluster:      clusterRawObject("ab123", clusterNetwork, "", 0, nil),
			errorMatcher: IsInvalidControlPlaneEndpointHostError,
		},
		{
			name:         "case 1: Invalid Port",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 80, nil),
			errorMatcher: IsInvalidControlPlaneEndpointPortError,
		},
		{
			name:         "case 2: Invalid Host",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.gigantic.io", 443, nil),
			errorMatcher: IsInvalidControlPlaneEndpointHostError,
		},
		{
			name:         "case 3: Valid values",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: nil,
		},
		{
			name:         "case 4: ClusterNetwork null",
			cluster:      clusterRawObject("ab123", nil, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: IsEmptyClusterNetworkError,
		},
		{
			name: "case 5: ClusterNetwork.APIServerPort wrong",
			cluster: clusterRawObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(80),
					ServiceDomain: "cluster.local",
					Services: &capi.NetworkRanges{
						CIDRBlocks: []string{
							"172.31.0.0/16",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsUnexpectedAPIServerPortError,
		},
		{
			name: "case 6: ClusterNetwork.ServiceDomain wrong",
			cluster: clusterRawObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "api.gigantic.io",
					Services: &capi.NetworkRanges{
						CIDRBlocks: []string{
							"172.31.0.0/16",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsUnexpectedServiceDomainError,
		},
		{
			name: "case 7: ClusterNetwork.Services nil",
			cluster: clusterRawObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "cluster.local",
					Services:      nil,
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsEmptyClusterNetworkServicesError,
		},
		{
			name: "case 8: ClusterNetwork.CIDRBlocks wrong",
			cluster: clusterRawObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "cluster.local",
					Services: &capi.NetworkRanges{
						CIDRBlocks: []string{
							"192.168.0.0/24",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsUnexpectedCIDRBlocksError,
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
			err = admit.Validate(ctx, getCreateAdmissionRequest(tc.cluster))

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
			Version:  "cluster.x-k8s.io/v1alpha3",
			Resource: "cluster",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}
