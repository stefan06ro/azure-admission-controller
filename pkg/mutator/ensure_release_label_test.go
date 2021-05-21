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
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func Test_EnsureReleaseLabel(t *testing.T) {
	testCases := []struct {
		name         string
		meta         metav1.Object
		patch        *PatchOperation
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: release already set",
			meta:         newObjectWithRelease(to.StringPtr("ab123"), to.StringPtr("v13.0.0")),
			patch:        nil,
			errorMatcher: nil,
		},
		{
			name:         "case 1: release wasn't set but cluster ID wasn't set either",
			meta:         newObjectWithRelease(nil, nil),
			patch:        nil,
			errorMatcher: IsClusterLabelNotFoundError,
		},
		{
			name:         "case 2: release wasn't set but cluster CR not found",
			meta:         newObjectWithRelease(to.StringPtr("non-existing"), nil),
			patch:        nil,
			errorMatcher: errors.IsNotFoundError,
		},
		{
			name:         "case 3: release wasn't set, cluster CR found but without a release label",
			meta:         newObjectWithRelease(to.StringPtr("cd456"), nil),
			patch:        nil,
			errorMatcher: IsReleaseLabelNotFoundError,
		},
		{
			name: "case 4: release wasn't set, cluster CR found, release label present",
			meta: newObjectWithRelease(to.StringPtr("ab123"), nil),
			patch: &PatchOperation{
				Operation: "add",
				Path:      "/metadata/labels/release.giantswarm.io~1version",
				Value:     "13.0.0",
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			ab123 := &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ab123",
					Namespace: "default",
					Labels: map[string]string{
						"release.giantswarm.io/version": "13.0.0",
					},
				},
			}
			err := ctrlClient.Create(ctx, ab123)
			if err != nil {
				t.Fatal(err)
			}

			azureClusterWithoutReleaseLabel := &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cd456",
					Namespace: "default",
				},
			}
			err = ctrlClient.Create(ctx, azureClusterWithoutReleaseLabel)
			if err != nil {
				t.Fatal(err)
			}

			ef789 := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ef789",
					Namespace: "default",
				},
			}
			err = ctrlClient.Create(ctx, ef789)
			if err != nil {
				t.Fatal(err)
			}

			patch, err := EnsureReleaseVersionLabel(ctx, ctrlClient, tc.meta)

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

func newObjectWithRelease(clusterID *string, release *string) metav1.Object {
	obj := &GenericObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"cluster.x-k8s.io/cluster-name":        "ab123",
				"cluster.x-k8s.io/control-plane":       "true",
				"giantswarm.io/machine-pool":           "ab123",
			},
		},
	}

	if clusterID != nil {
		obj.Labels[label.Cluster] = *clusterID
	}

	if release != nil {
		obj.Labels[label.ReleaseVersion] = *release
	}

	return obj
}
