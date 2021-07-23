// +build liveinstallation

package validateliveresources

import (
	"context"
	"testing"

	"github.com/giantswarm/micrologger"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	machinepoolpkg "github.com/giantswarm/azure-admission-controller/pkg/machinepool"
)

func TestMachinePools(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var machinePoolWebhookHandler *machinepoolpkg.WebhookHandler
	{
		c := machinepoolpkg.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    NewDecoder(),
			Logger:     logger,
			VMcaps:     NewVMCapabilities(t, logger),
		}
		machinePoolWebhookHandler, err = machinepoolpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var machinePoolList capiexp.MachinePoolList
	err = ctrlClient.List(ctx, &machinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	for _, machinePool := range machinePoolList.Items {
		err = machinePoolWebhookHandler.OnCreateValidate(ctx, &machinePool)
		if err != nil {
			t.Fatal(err)
		}

		updatedMachinePool := machinePool.DeepCopy()

		updatedMachinePool.Labels["test.giantswarm.io/dummy"] = "this is not really saved"
		err = machinePoolWebhookHandler.OnUpdateValidate(ctx, &machinePool, updatedMachinePool)
		if err != nil {
			t.Fatal(err)
		}
	}
}
