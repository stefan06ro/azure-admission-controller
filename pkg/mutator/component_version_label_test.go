package mutator

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func Test_EnsureComponentVersionLabel(t *testing.T) {
	testCases := []struct {
		name         string
		meta         metav1.Object
		patch        *PatchOperation
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: azure operator label missing",
			meta: newObjectWithLabels(to.StringPtr("ab123"), nil),
			patch: &PatchOperation{
				Operation: "add",
				Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
				Value:     "5.0.0",
			},
			errorMatcher: nil,
		},
		{
			name:         "case 1: azure operator label present",
			meta:         newObjectWithLabels(to.StringPtr("ab123"), map[string]string{label.ReleaseVersion: "v13.0.0", label.AzureOperatorVersion: "5.0.0"}),
			patch:        nil,
			errorMatcher: nil,
		},
		{
			name:         "case 2: operator label missing, cluster not present",
			meta:         newObjectWithLabels(to.StringPtr("nf404"), map[string]string{}),
			patch:        nil,
			errorMatcher: errors.IsNotFoundError,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// AzureCluster with azure operator label.
			ab123 := &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"azure-operator.giantswarm.io/version": "5.0.0",
					},
				},
			}
			err := ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			// AzureCluster lacking any operator annotation.
			ef789 := &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ef789",
					Namespace: "default",
				},
			}
			err = ctrlClient.Create(ctx, ef789)
			if err != nil {
				t.Fatal(err)
			}

			patch, err := CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx, ctrlClient, tc.meta)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			// Check if the validation result is the expected one.
			if !reflect.DeepEqual(tc.patch, patch) {
				t.Fatalf("Patch mismatch: expected %v, got %v", tc.patch, patch)
			}
		})
	}
}

func newObjectWithLabels(clusterID *string, labels map[string]string) metav1.Object {
	mergedLabels := map[string]string{
		"cluster.x-k8s.io/cluster-name":  "ab123",
		"cluster.x-k8s.io/control-plane": "true",
		"giantswarm.io/machine-pool":     "ab123",
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}
	obj := &GenericObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Unknown",
			APIVersion: "unknown.generic.example/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels:    mergedLabels,
		},
	}

	if clusterID != nil {
		obj.Labels[label.Cluster] = *clusterID
	}

	return obj
}
