package release

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck
	"sigs.k8s.io/yaml"
)

func Test_GetComponentVersionsFromRelease(t *testing.T) {
	testCases := []struct {
		name               string
		inputRelease       string
		expectedComponents map[string]string
	}{
		{
			name:         "Release v14.1.4",
			inputRelease: "14.1.4",
			expectedComponents: map[string]string{
				"app-operator":     "3.2.1",
				"azure-operator":   "5.5.2",
				"cert-operator":    "0.1.0",
				"cluster-operator": "0.23.22",
				"kubernetes":       "1.19.8",
				"containerlinux":   "2605.12.0",
				"calico":           "3.15.3",
				"etcd":             "3.4.14",
			},
		},
		{
			name:         "Release v20.0.0-v1alpha3",
			inputRelease: "20.0.0-v1alpha3",
			expectedComponents: map[string]string{
				"app-operator":                           "3.2.0",
				"cluster-api-bootstrap-provider-kubeadm": "0.0.0",
				"cluster-api-control-plane":              "0.0.0",
				"cluster-api-core":                       "0.0.1",
				"cluster-api-provider-azure":             "0.0.0",
				"kubernetes":                             "1.19.8",
				"containerlinux":                         "2605.12.0",
				"calico":                                 "3.15.3",
				"etcd":                                   "3.4.14",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)
			ctx := context.Background()
			ctrlClient := newFakeClient()
			loadReleases(t, ctrlClient, tc.inputRelease)

			result, err := GetComponentVersionsFromRelease(ctx, ctrlClient, tc.inputRelease)
			if err != nil {
				t.Fatalf("Error while calling GetComponentVersionsFromRelease: %#v", err)
			}

			// Assert number of components
			if len(result) != len(tc.expectedComponents) {
				t.Errorf("Expected %d components in a release, got %d", len(tc.expectedComponents), len(result))
			}

			// Assert all returned components have been expected with correct versions
			for resultComponent, resultComponentVersion := range result {
				expectedComponentVersion, ok := tc.expectedComponents[resultComponent]
				if !ok {
					t.Errorf("Component %q was not expected in the release", resultComponent)
					continue
				}

				if resultComponentVersion != expectedComponentVersion {
					t.Errorf("Expected comonent %q to have version %s, got %s instead", resultComponent, expectedComponentVersion, resultComponentVersion)
				}
			}

			// Assert all expected components have been returned
			for expectedComponent := range tc.expectedComponents {
				_, ok := result[expectedComponent]
				if !ok {
					t.Errorf("Expected component %q was not found in the release", expectedComponent)
					continue
				}
			}
		})
	}
}

func Test_IsLegacy(t *testing.T) {
	testCases := []struct {
		name           string
		inputRelease   string
		expectedResult bool
	}{
		{
			name:           "Release v14.1.4",
			inputRelease:   "14.1.4",
			expectedResult: true,
		},
		{
			name:           "Release v20.0.0-v1alpha3",
			inputRelease:   "20.0.0-v1alpha3",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)
			ctx := context.Background()
			ctrlClient := newFakeClient()
			loadReleases(t, ctrlClient, tc.inputRelease)

			release, err := FindRelease(ctx, ctrlClient, tc.inputRelease)
			if err != nil {
				t.Fatalf("Error while calling FindRelease: %#v", err)
			}

			result := IsLegacy(release)

			if result != tc.expectedResult {
				var expectedMsg string
				var gotMsg string
				if tc.expectedResult == true {
					expectedMsg = "release is legacy"
					gotMsg = "release is not legacy"
				} else {
					expectedMsg = "release is not legacy"
					gotMsg = "release is legacy"
				}

				t.Errorf("Expected %t (%s), got %t instead (%s)", tc.expectedResult, expectedMsg, result, gotMsg)
			}
		})
	}
}

func Test_TryFindReleaseForObject(t *testing.T) {
	testCases := []struct {
		name           string
		inputRelease   string
		object         metav1.Object
		ownerCluster   capi.Cluster
		expectedResult bool
	}{
		{
			name:         "Cluster with release label",
			inputRelease: "14.1.4",
			object: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "14.1.4",
					},
				},
			},
			expectedResult: true,
		},
		{
			name:           "Cluster without release label",
			inputRelease:   "14.1.4",
			object:         &capi.Cluster{},
			expectedResult: false,
		},
		{
			name:         "AzureCluster with release label",
			inputRelease: "14.1.4",
			object: &capz.AzureCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "14.1.4",
					},
				},
			},
			expectedResult: true,
		},
		{
			name:         "AzureCluster without release label, Cluster with release label",
			inputRelease: "14.1.4",
			object:       &capz.AzureCluster{},
			ownerCluster: capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "14.1.4",
					},
				},
			},
			expectedResult: true,
		},
		{
			name:           "AzureCluster without release label, Cluster without release label",
			inputRelease:   "14.1.4",
			object:         &capz.AzureCluster{},
			ownerCluster:   capi.Cluster{},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)
			ctx := context.Background()
			ctrlClient := newFakeClient()
			loadReleases(t, ctrlClient, tc.inputRelease)

			clusterGetter := func(metav1.Object) capi.Cluster {
				return tc.ownerCluster
			}

			_, ok, err := TryFindReleaseForObject(ctx, ctrlClient, tc.object, clusterGetter)
			if err != nil {
				t.Fatalf("Error while calling TryFindReleaseForObject: %#v", err)
			}

			if ok != tc.expectedResult {
				var expectedMsg string
				var gotMsg string
				if tc.expectedResult == true {
					expectedMsg = "release found"
					gotMsg = "release is not found"
				} else {
					expectedMsg = "release is not found"
					gotMsg = "release is found"
				}

				t.Errorf("Expected %t (%s), got %t instead (%s)", tc.expectedResult, expectedMsg, ok, gotMsg)
			}
		})
	}
}

func loadReleases(t *testing.T, client client.Client, testReleasesToLoad ...string) {
	for _, releaseVersion := range testReleasesToLoad {
		t.Logf("Loading release %s to fake client...", releaseVersion)
		var err error
		fileName := fmt.Sprintf("release-v%s.yaml", releaseVersion)

		var bs []byte
		{
			bs, err = ioutil.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
			}
		}

		// First parse kind.
		typeMeta := &metav1.TypeMeta{}
		err = yaml.Unmarshal(bs, typeMeta)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		if typeMeta.Kind != "Release" {
			t.Fatalf("failed to create object from input file %s, the file does not contain Release CR", fileName)
		}

		var release releasev1alpha1.Release
		err = yaml.Unmarshal(bs, &release)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		err = client.Create(context.Background(), &release)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}
	}
}

func newFakeClient() client.Client {
	scheme := runtime.NewScheme()

	err := capi.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = capz.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = corev1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = providerv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = releasev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	return fake.NewFakeClientWithScheme(scheme)
}
