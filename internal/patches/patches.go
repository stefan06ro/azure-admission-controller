package patches

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/giantswarm/microerror"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func GenerateFrom(originalJSON []byte, current runtime.Object) ([]mutator.PatchOperation, error) {
	currentJSON, err := json.Marshal(current)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var patches []mutator.PatchOperation
	{
		var jsonPatches []jsonpatch.JsonPatchOperation
		jsonPatches, err = jsonpatch.CreatePatch(originalJSON, currentJSON)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		patches = make([]mutator.PatchOperation, 0, len(jsonPatches))
		for _, patch := range jsonPatches {
			patches = append(patches, mutator.PatchOperation(patch))
		}

		sort.SliceStable(patches, func(i, j int) bool {
			return patches[i].Path < patches[j].Path
		})
	}

	return patches, nil
}

func SkipForPath(path string, patches []mutator.PatchOperation) []mutator.PatchOperation {
	var modifiedPatches []mutator.PatchOperation
	{
		for _, patch := range patches {
			if !strings.HasPrefix(patch.Path, path) {
				modifiedPatches = append(modifiedPatches, patch)
			}
		}
	}

	return modifiedPatches
}
