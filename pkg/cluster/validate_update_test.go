package cluster

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestClusterUpdateValidate(t *testing.T) {
	type testCase struct {
		name         string
		oldCluster   *capi.Cluster
		newCluster   *capi.Cluster
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

	var testCases = []testCase{
		{
			name:         "case 0: unchanged ControlPlaneEndpoint",
			oldCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: nil,
		},
		{
			name:         "case 1: host changed",
			oldCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterObject("ab123", clusterNetwork, "api.azure.gigantic.io", 443, nil),
			errorMatcher: IsControlPlaneEndpointWasChangedError,
		},
		{
			name:         "case 2: port changed",
			oldCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 80, nil),
			errorMatcher: IsControlPlaneEndpointWasChangedError,
		},
		{
			name:         "case 3: clusterNetwork deleted",
			oldCluster:   clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster:   clusterObject("ab123", nil, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 4: clusterNetwork.APIServerPort changed",
			oldCluster: clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(80),
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
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 5: clusterNetwork.ServiceDomain changed",
			oldCluster: clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterObject(
				"ab123",
				&capi.ClusterNetwork{
					APIServerPort: to.Int32Ptr(443),
					ServiceDomain: "api.gigantic.io",
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
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:       "case 6: clusterNetwork.Services deleted",
			oldCluster: clusterObject("ab123", clusterNetwork, "api.ab123.test.westeurope.azure.gigantic.io", 443, nil),
			newCluster: clusterObject(
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
			errorMatcher: IsClusterNetworkWasChangedError,
		},
		{
			name:         "case 7: host changed but object is being deleted",
			oldCluster:   builder.BuildCluster(builder.Name("ab123"), builder.WithDeletionTimestamp(), builder.ControlPlaneEndpoint("api.ab123.test.westeurope.azure.gigantic.io", 443)),
			newCluster:   builder.BuildCluster(builder.Name("ab123"), builder.WithDeletionTimestamp(), builder.ControlPlaneEndpoint("api.azure.gigantic.io", 443)),
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

			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				BaseDomain: "k8s.test.westeurope.azure.gigantic.io",
				CtrlClient: ctrlClient,
				CtrlReader: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on Cluster update.
			err = handler.OnUpdateValidate(context.Background(), tc.oldCluster, tc.newCluster)

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
