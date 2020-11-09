package azuremachinepool

import (
	"encoding/json"
	"math/rand"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
)

type BuilderOption func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func AzureMachinePool(azureMachinePoolName string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Spec.Template.Spec.InfrastructureRef.Name = azureMachinePoolName
		return machinePool
	}
}

func FailureDomains(failureDomains []string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.Spec.FailureDomains = failureDomains
		return machinePool
	}
}

func Name(name string) BuilderOption {
	return func(machinePool *expcapiv1alpha3.MachinePool) *expcapiv1alpha3.MachinePool {
		machinePool.ObjectMeta.Name = name
		machinePool.Labels[label.MachinePool] = name
		return machinePool
	}
}

func BuildMachinePool(opts ...BuilderOption) *expcapiv1alpha3.MachinePool {
	nodepoolName := generateName()
	machinePool := &expcapiv1alpha3.MachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodepoolName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion: "5.0.0",
				label.Cluster:              "ab123",
				label.MachinePool:          nodepoolName,
				label.Organization:         "giantswarm",
				label.ReleaseVersion:       "13.0.0",
			},
		},
		Spec: expcapiv1alpha3.MachinePoolSpec{
			FailureDomains: []string{},
			Template: v1alpha3.MachineTemplateSpec{
				Spec: v1alpha3.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Namespace: "org-giantswarm",
						Name:      "ab123",
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(machinePool)
	}

	return machinePool
}

func BuildMachinePoolAsJson(opts ...BuilderOption) []byte {
	machinePool := BuildMachinePool(opts...)

	byt, _ := json.Marshal(machinePool)

	return byt
}

func generateName() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
