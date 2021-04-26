package mutator

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/giantswarm/microerror"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"
)

func GenerateFrom(originalJSON []byte, current runtime.Object) ([]PatchOperation, error) {
	currentJSON, err := json.Marshal(current)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var patches []PatchOperation
	{
		var jsonPatches []jsonpatch.JsonPatchOperation
		jsonPatches, err = jsonpatch.CreatePatch(originalJSON, currentJSON)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		patches = make([]PatchOperation, 0, len(jsonPatches))
		for _, patch := range jsonPatches {
			patches = append(patches, PatchOperation(patch))
		}

		sort.SliceStable(patches, func(i, j int) bool {
			return patches[i].Path < patches[j].Path
		})
	}

	return patches, nil
}

func GenerateFromObjectDiff(original runtime.Object, current runtime.Object) ([]PatchOperation, error) {
	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentJSON, err := json.Marshal(current)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var patches []PatchOperation
	{
		var jsonPatches []jsonpatch.JsonPatchOperation
		jsonPatches, err = jsonpatch.CreatePatch(originalJSON, currentJSON)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		patches = make([]PatchOperation, 0, len(jsonPatches))
		for _, patch := range jsonPatches {
			patches = append(patches, PatchOperation(patch))
		}

		sort.SliceStable(patches, func(i, j int) bool {
			return patches[i].Path < patches[j].Path
		})
	}

	return patches, nil
}

func SkipForPath(path string, patches []PatchOperation) []PatchOperation {
	var modifiedPatches []PatchOperation
	{
		for _, patch := range patches {
			if !strings.HasPrefix(patch.Path, path) {
				modifiedPatches = append(modifiedPatches, patch)
			}
		}
	}

	return modifiedPatches
}
