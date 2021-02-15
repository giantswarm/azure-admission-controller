package patches

import (
	"encoding/json"
	"strings"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/microerror"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"
)

func GeneratePatchesFrom(originalJSON []byte, current runtime.Object) ([]mutator.PatchOperation, error) {
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
	}

	return patches, nil
}

func SkipPatchesForPath(path string, patches []mutator.PatchOperation) []mutator.PatchOperation {
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
