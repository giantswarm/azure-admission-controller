package machinepool

import (
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

// Ensure the needed escapes are in place. See https://tools.ietf.org/html/rfc6901#section-3 .
func escapeJSONPatchString(input string) string {
	input = strings.ReplaceAll(input, "~", "~0")
	input = strings.ReplaceAll(input, "/", "~1")

	return input
}

// ensureAutoscalingAnnotations ensures the custom annotations used to determine the min and max replicas for
// the cluster autoscaler are set in the Machinepool CR.
func ensureAutoscalingAnnotations(m mutator.Mutator, machinePool *capiexp.MachinePool) []mutator.PatchOperation {
	var patches []mutator.PatchOperation

	// The replicas field could not be set, we default to 1.
	clusterReplicas := int32(defaultReplicas)
	if machinePool.Spec.Replicas != nil {
		clusterReplicas = *machinePool.Spec.Replicas
	}

	if machinePool.Annotations[annotation.NodePoolMinSize] == "" {
		m.Log("level", "debug", "message", fmt.Sprintf("setting MachinePool Annotation %s to %d", annotation.NodePoolMinSize, clusterReplicas))
		patches = append(patches, *mutator.PatchAdd(fmt.Sprintf("/metadata/annotations/%s", escapeJSONPatchString(annotation.NodePoolMinSize)), clusterReplicas))
	}

	if machinePool.Annotations[annotation.NodePoolMaxSize] == "" {
		m.Log("level", "debug", "message", fmt.Sprintf("setting MachinePool Annotation %s to %d", annotation.NodePoolMaxSize, clusterReplicas))
		patches = append(patches, *mutator.PatchAdd(fmt.Sprintf("/metadata/annotations/%s", escapeJSONPatchString(annotation.NodePoolMaxSize)), clusterReplicas))
	}

	return patches
}
