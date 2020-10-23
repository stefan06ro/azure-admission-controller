package cluster

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func TestClusterCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		cluster      []byte
		allowed      bool
		errorMatcher func(err error) bool
	}

	clusterNetwork := &v1alpha3.ClusterNetwork{
		APIServerPort: to.Int32Ptr(443),
		ServiceDomain: "ab123.k8s.test.westeurope.azure.gigantic.io",
		Services: &v1alpha3.NetworkRanges{
			CIDRBlocks: []string{
				"172.31.0.0/16",
			},
		},
	}

	testCases := []testCase{
		{
			name:         "case 0: empty ControlPlaneEndpoint",
			cluster:      clusterRawObject("ab123", clusterNetwork, "", 0),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 1: Invalid Port",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 80),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 2: Invalid Host",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.gigantic.io", 443),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 3: Valid values",
			cluster:      clusterRawObject("ab123", clusterNetwork, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443),
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name:         "case 4: ClusterNetwork null",
			cluster:      clusterRawObject("ab123", nil, "api.ab123.k8s.test.westeurope.azure.gigantic.io", 443),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name: "case 5: ClusterNetwork.APIServerPort wrong",
			cluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(80),
					ServiceDomain: "ab123.k8s.test.westeurope.azure.gigantic.io",
					Services: &v1alpha3.NetworkRanges{
						CIDRBlocks: []string{
							"172.31.0.0/16",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
			),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name: "case 6: ClusterNetwork.ServiceDomain wrong",
			cluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "api.gigantic.io",
					Services: &v1alpha3.NetworkRanges{
						CIDRBlocks: []string{
							"172.31.0.0/16",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
			),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name: "case 7: ClusterNetwork.Services nil",
			cluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "ab123.k8s.test.westeurope.azure.gigantic.io",
					Services:      nil,
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
			),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name: "case 8: ClusterNetwork.CIDRBlocks wrong",
			cluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "ab123.k8s.test.westeurope.azure.gigantic.io",
					Services: &v1alpha3.NetworkRanges{
						CIDRBlocks: []string{
							"192.168.0.0/24",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
			),
			allowed:      false,
			errorMatcher: errors.IsInvalidOperationError,
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
			allowed, err := admit.Validate(context.Background(), getCreateAdmissionRequest(tc.cluster))

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
