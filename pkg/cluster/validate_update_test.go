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
)

func TestClusterUpdateValidate(t *testing.T) {
	type testCase struct {
		name         string
		oldCluster   []byte
		newCluster   []byte
		errorMatcher func(err error) bool
	}

	clusterNetwork := &v1alpha3.ClusterNetwork{
		APIServerPort: to.Int32Ptr(443),
		ServiceDomain: "cluster.local",
		Services: &v1alpha3.NetworkRanges{
			CIDRBlocks: []string{
				"172.31.0.0/16",
			},
		},
	}

	var testCases = []testCase{
		{
			name:         "case 0: unchanged ControlPlaneEndpoint",
			oldCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: nil,
		},
		{
			name:         "case 1: host changed",
			oldCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterRawObject("ab123", clusterNetwork, "api.azure.gigantic.io", 443, nil),
			errorMatcher: IsControlPlaneEndpointWasChangedError,
		},
		{
			name:         "case 2: port changed",
			oldCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 80, nil),
			errorMatcher: IsControlPlaneEndpointWasChangedError,
		},
		{
			name:         "case 3: clusterNetwork deleted",
			oldCluster:   clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterRawObject("ab123", nil, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 4: clusterNetwork.APIServerPort changed",
			oldCluster: clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(80),
					ServiceDomain: "cluster.local",
					Services: &v1alpha3.NetworkRanges{
						CIDRBlocks: []string{
							"192.168.0.0/24",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 5: clusterNetwork.ServiceDomain changed",
			oldCluster: clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "api.gigantic.io",
					Services: &v1alpha3.NetworkRanges{
						CIDRBlocks: []string{
							"192.168.0.0/24",
						},
					},
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 6: clusterNetwork.Services deleted",
			oldCluster: clusterRawObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterRawObject(
				"ab123",
				&v1alpha3.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "cluster.local",
					Services:      nil,
				},
				"api.ab123.k8s.test.westeurope.azure.gigantic.io",
				443,
				nil,
			),
			errorMatcher: IsClusterNetworkWasChangedError,
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
			err = admit.Validate(context.Background(), getUpdateAdmissionRequest(tc.oldCluster, tc.newCluster))

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
			Version:  "cluster.x-k8s.io/v1alpha3",
			Resource: "cluster",
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
