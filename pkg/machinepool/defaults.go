package machinepool

import (
	"context"
	"fmt"
	"strconv"

	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

// setDefaultSpecValues checks if some optional field is not set, and sets
// default values defined by upstream Cluster API.
func (m *CreateMutator) setDefaultSpecValues(ctx context.Context, machinePool *capiexp.MachinePool) []mutator.PatchOperation {
	var patches []mutator.PatchOperation
	m.logger.LogCtx(ctx, "level", "debug", "message", "setting default MachinePool.Spec values")

	defaultSpecReplicas := m.setDefaultReplicaValue(ctx, machinePool)
	if defaultSpecReplicas != nil {
		patches = append(patches, *defaultSpecReplicas)
	}

	return patches
}

// setDefaultReplicaValue checks if Spec.Replicas has been set, and if it is
// not, it sets its value to 1.
func (m *CreateMutator) setDefaultReplicaValue(ctx context.Context, machinePool *capiexp.MachinePool) *mutator.PatchOperation {
	currentValue := ""
	if machinePool.Spec.Replicas == nil {
		currentValue = "nil"
		const defaultReplicas = "1"
		m.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting default MachinePool.Spec.Replicas to %s", defaultReplicas))
		return mutator.PatchAdd("/spec/replicas", defaultReplicas)
	} else {
		currentValue = strconv.Itoa(int(*machinePool.Spec.Replicas))
	}

	m.logger.LogCtx(ctx,
		"level", "debug",
		"message", fmt.Sprintf("setting default MachinePool.Spec.Replica value, value was %s", currentValue))

	return nil
}
