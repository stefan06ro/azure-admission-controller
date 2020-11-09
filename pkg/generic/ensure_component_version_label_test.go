package generic

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func Test_EnsureComponentVersionLabel(t *testing.T) {
	testCases := []struct {
		name         string
		meta         metav1.Object
		patches      []mutator.PatchOperation
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: both operators missing",
			meta: newObjectWithLabels(to.StringPtr("ab123"), map[string]string{}),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
					Value:     "5.0.0",
				},
				{
					Operation: "add",
					Path:      "/metadata/labels/cluster-operator.giantswarm.io~1version",
					Value:     "0.23.18",
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: cluster operator missing",
			meta: newObjectWithLabels(to.StringPtr("ab123"), map[string]string{label.AzureOperatorVersion: "5.0.0"}),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/cluster-operator.giantswarm.io~1version",
					Value:     "0.23.18",
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: azure operator missing",
			meta: newObjectWithLabels(to.StringPtr("ab123"), map[string]string{label.ClusterOperatorVersion: "0.23.18"}),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
					Value:     "5.0.0",
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 3: both operators present",
			meta:         newObjectWithLabels(to.StringPtr("ab123"), map[string]string{label.ReleaseVersion: "v13.0.0", label.AzureOperatorVersion: "5.0.0", label.ClusterOperatorVersion: "0.23.18"}),
			patches:      nil,
			errorMatcher: nil,
		},
		{
			name: "case 4: both operators missing, cluster present but lacks one operator's label",
			meta: newObjectWithLabels(to.StringPtr("cd456"), map[string]string{}),
			patches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/labels/azure-operator.giantswarm.io~1version",
					Value:     "5.0.0",
				},
			},
			errorMatcher: errors.IsInvalidOperationError,
		},
		{
			name:         "case 5: both operators missing, cluster not present",
			meta:         newObjectWithLabels(to.StringPtr("nf404"), map[string]string{}),
			patches:      nil,
			errorMatcher: errors.IsInvalidOperationError,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// Cluster with both operator annotations.
			ab123 := &v1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"azure-operator.giantswarm.io/version":   "5.0.0",
						"cluster-operator.giantswarm.io/version": "0.23.18",
					},
				},
			}
			err := ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			// Cluster lacking cluster-operator annotation.
			cd456 := &v1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cd456",
					Namespace: "default",
					Labels: map[string]string{
						"release.giantswarm.io/version":        "13.0.0",
						"azure-operator.giantswarm.io/version": "5.0.0",
					},
				},
			}
			err = ctrlClient.Create(ctx, cd456)
			if err != nil {
				t.Fatal(err)
			}

			// Cluster lacking any operator annotation.
			ef789 := &v1alpha3.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ef789",
					Namespace: "default",
				},
			}
			err = ctrlClient.Create(ctx, ef789)
			if err != nil {
				t.Fatal(err)
			}

			var patches []mutator.PatchOperation
			var patch1, patch2 *mutator.PatchOperation
			patch1, err = EnsureComponentVersionLabel(ctx, ctrlClient, tc.meta, label.AzureOperatorVersion)
			if err == nil {
				patch2, err = EnsureComponentVersionLabel(ctx, ctrlClient, tc.meta, label.ClusterOperatorVersion)
			}

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

			if patch1 != nil {
				patches = append(patches, *patch1)
			}
			if patch2 != nil {
				patches = append(patches, *patch2)
			}

			// Check if the validation result is the expected one.
			if !reflect.DeepEqual(tc.patches, patches) {
				t.Fatalf("Patch mismatch: expected %v, got %v", tc.patches, patches)
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
