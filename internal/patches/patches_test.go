package patches

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"

	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func TestGenerateFrom(t *testing.T) {
	testCases := []struct {
		name            string
		originalObject  runtime.Object
		currentObject   runtime.Object
		expectedPatches []mutator.PatchOperation
	}{
		{
			name:            "case 0 finds no difference between identical objects",
			originalObject:  &capiv1alpha3.Cluster{},
			currentObject:   &capiv1alpha3.Cluster{},
			expectedPatches: []mutator.PatchOperation{},
		},
		{
			name:           "case 1 finds the differences between two compatible objects",
			originalObject: &capiv1alpha3.Cluster{},
			currentObject: &capiv1alpha3.Cluster{
				Spec: capiv1alpha3.ClusterSpec{ClusterNetwork: &capiv1alpha3.ClusterNetwork{Services: &capiv1alpha3.NetworkRanges{CIDRBlocks: []string{"1", "2", "3"}}}},
			},
			expectedPatches: []mutator.PatchOperation{
				{
					Operation: "add",
					Path:      "/spec/clusterNetwork",
					Value: map[string]interface{}{
						"services": map[string]interface{}{
							"cidrBlocks": []interface{}{
								"1", "2", "3",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalObjectJSON, err := json.Marshal(tc.originalObject)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			patches, err := GenerateFrom(originalObjectJSON, tc.currentObject)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if len(tc.expectedPatches) == 0 && len(patches) == 0 {
				return
			}

			if !reflect.DeepEqual(tc.expectedPatches, patches) {
				t.Fatalf("patches mismatch: expected %v, got %v", tc.expectedPatches, patches)
			}
		})
	}
}

func TestSkipForPath(t *testing.T) {
	patches := []mutator.PatchOperation{
		{
			Operation: "add",
			Path:      "/spec/test",
		},
		{
			Operation: "remove",
			Path:      "/spec/test/someother/other",
		},
		{
			Operation: "add",
			Path:      "/justapath/test",
		},
		{
			Operation: "add",
			Path:      "/testPath",
		},
		{
			Operation: "add",
			Path:      "/spec/test/someother",
		},
	}
	expectedPatches := []mutator.PatchOperation{
		{
			Operation: "add",
			Path:      "/justapath/test",
		},
		{
			Operation: "add",
			Path:      "/testPath",
		},
	}

	filteredPatches := SkipForPath("/spec/test", patches)

	if !reflect.DeepEqual(expectedPatches, filteredPatches) {
		t.Fatalf("patches mismatch: expected %v, got %v", expectedPatches, filteredPatches)
	}
}
