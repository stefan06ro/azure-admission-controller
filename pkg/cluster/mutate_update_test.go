package cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestClusterUpdateMutate(t *testing.T) {
	type testCase struct {
		name         string
		cluster      *capi.Cluster
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:    "case 0: Wrong azure-operator component label",
			cluster: builder.BuildCluster(builder.Name("ab123"), builder.Labels(map[string]string{"release.giantswarm.io/version": "v13.1.0", "azure-operator.giantswarm.io/version": "4.2.0", "cluster-operator.giantswarm.io/version": "0.23.11"})),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
					Value:     "5.1.0",
				},
			},
			errorMatcher: nil,
		},
		{
			name:    "case 0: Wrong azure-operator component label",
			cluster: builder.BuildCluster(builder.Name("ab123"), builder.Labels(map[string]string{"release.giantswarm.io/version": "v13.1.0", "azure-operator.giantswarm.io/version": "5.1.0", "cluster-operator.giantswarm.io/version": "0.23.10"})),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/cluster-operator.giantswarm.io~1version",
					Value:     "0.23.11",
				},
			},
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

			release131 := &v1alpha1.Release{
				ObjectMeta: metav1.ObjectMeta{
					Name: "v13.1.0",
				},
				Spec: v1alpha1.ReleaseSpec{
					Components: []v1alpha1.ReleaseSpecComponent{
						{
							Name:    "azure-operator",
							Version: "5.1.0",
						},
						{
							Name:    "cluster-operator",
							Version: "0.23.11",
						},
					},
				},
			}
			err = ctrlClient.Create(ctx, release131)
			if err != nil {
				t.Fatal(err)
			}

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

			// Run mutation webhook handler on Cluster update.
			patches, err := handler.OnUpdateMutate(context.Background(), nil, tc.cluster)

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
			if len(tc.patches) != 0 || len(patches) != 0 {
				if !reflect.DeepEqual(tc.patches, patches) {
					t.Fatalf("Patches mismatch: expected %v, got %v", tc.patches, patches)
				}
			}
		})
	}
}
