package app

import (
	"net/http"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/azure-admission-controller/v2/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/config"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/unittest"
)

func Test_RegisterWebhookHandlers(t *testing.T) {
	var err error
	var logger micrologger.Logger
	{
		logger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Dummy config that we otherwise get from flags.
	cfg := config.Config{
		BaseDomain: "k8s.test.westeurope.azure.gigantic.io",
		Location:   "westeurope",
	}

	fakeK8sClient := unittest.FakeK8sClient()
	ctrlClient := fakeK8sClient.CtrlClient()

	vmcaps, err := vmcapabilities.NewFactory(logger)
	if err != nil {
		t.Fatal(microerror.JSON(err))
	}

	// Real *http.ServeMux, not that we gonna run it here.
	handler := http.NewServeMux()

	// Run webhook handlers registration.
	err = RegisterWebhookHandlers(handler, cfg, logger, ctrlClient, ctrlClient, vmcaps)
	if err != nil {
		t.Fatalf("Error while registering webhook handlers %#v", err)
	}
}
