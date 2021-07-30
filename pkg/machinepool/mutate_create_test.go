package machinepool

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/machinepool"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestMachinePoolCreateMutate(t *testing.T) {
	type testCase struct {
		name         string
		nodePool     *capiexp.MachinePool
		patches      []mutator.PatchOperation
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:     "case 0: set default number of replicas",
			nodePool: builder.BuildMachinePool(),
			patches: []mutator.PatchOperation{
				{
					Operation: "replace",
					Path:      "/metadata/labels/cluster.x-k8s.io~1cluster-name",
					Value:     "",
				},
				{
					Operation: "add",
					Path:      "/spec/minReadySeconds",
					Value:     float64(0),
				},
				{
					Operation: "add",
					Path:      "/spec/replicas",
					Value:     float64(1),
				},
			},
			errorMatcher: nil,
		},
		{
			name:     "case 1: set min replicas annotation when replicas field is set",
			nodePool: builder.BuildMachinePool(builder.Replicas(7), builder.Annotation(annotation.NodePoolMinSize, "")),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-min-size",
					Value:     "7",
				},
				{
					Operation: "replace",
					Path:      "/metadata/labels/cluster.x-k8s.io~1cluster-name",
					Value:     "",
				},
				{
					Operation: "add",
					Path:      "/spec/minReadySeconds",
					Value:     float64(0),
				},
			},
			errorMatcher: nil,
		},
		{
			name:     "case 2: set max replicas annotation when replicas field is set",
			nodePool: builder.BuildMachinePool(builder.Replicas(7), builder.Annotation(annotation.NodePoolMaxSize, "")),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-max-size",
					Value:     "7",
				},
				{
					Operation: "replace",
					Path:      "/metadata/labels/cluster.x-k8s.io~1cluster-name",
					Value:     "",
				},
				{
					Operation: "add",
					Path:      "/spec/minReadySeconds",
					Value:     float64(0),
				},
			},
			errorMatcher: nil,
		},
		{
			name:     "case 3: set min and max replicas annotation when replicas field is set",
			nodePool: builder.BuildMachinePool(builder.Replicas(7), builder.Annotation(annotation.NodePoolMinSize, ""), builder.Annotation(annotation.NodePoolMaxSize, "")),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-min-size",
					Value:     "7",
				},
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-max-size",
					Value:     "7",
				},
				{
					Operation: "replace",
					Path:      "/metadata/labels/cluster.x-k8s.io~1cluster-name",
					Value:     "",
				},
				{
					Operation: "add",
					Path:      "/spec/minReadySeconds",
					Value:     float64(0),
				},
			},
			errorMatcher: nil,
		},
		{
			name:     "case 4: set min and max replicas annotation when replicas field is not set",
			nodePool: builder.BuildMachinePool(builder.Annotation(annotation.NodePoolMinSize, ""), builder.Annotation(annotation.NodePoolMaxSize, "")),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-min-size",
					Value:     "1",
				},
				{
					Operation: "add",
					Path:      "/metadata/annotations/cluster.k8s.io~1cluster-api-autoscaler-node-group-max-size",
					Value:     "1",
				},
				{
					Operation: "replace",
					Path:      "/metadata/labels/cluster.x-k8s.io~1cluster-name",
					Value:     "",
				},
				{
					Operation: "add",
					Path:      "/spec/minReadySeconds",
					Value:     float64(0),
				},
				{
					Operation: "add",
					Path:      "/spec/replicas",
					Value:     float64(1),
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

			release13 := &v1alpha1.Release{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "v13.0.0-alpha4",
					Namespace: "default",
				},
				Spec: v1alpha1.ReleaseSpec{
					Components: []v1alpha1.ReleaseSpecComponent{
						{
							Name:    "azure-operator",
							Version: "5.0.0",
						},
					},
				},
			}
			err = ctrlClient.Create(ctx, release13)
			if err != nil {
				t.Fatal(err)
			}

			// Cluster with both operator annotations.
			ab123 := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"azure-operator.giantswarm.io/version": "5.0.0",
					},
				},
			}
			err = ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			stubbedSKUs := map[string]compute.ResourceSku{}
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
				Logger:     newLogger,
				VMcaps:     vmcaps,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run mutating webhook handler on MachinePool creation.
			patches, err := handler.OnCreateMutate(ctx, tc.nodePool)

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
			if !reflect.DeepEqual(tc.patches, patches) {
				t.Fatalf("Patches mismatch: expected %v, got %v", tc.patches, patches)
			}
		})
	}
}
