// +build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	azureclusterpkg "github.com/giantswarm/azure-admission-controller/pkg/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

func TestAzureClusterFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var azureClusterList capz.AzureClusterList
	err := ctrlClient.List(ctx, &azureClusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureCluster := range azureClusterList.Items {
		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, ctrlClient, objectMeta)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &azureCluster, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			clusterName := fmt.Sprintf("%s/%s", azureCluster.Namespace, azureCluster.Name)
			t.Errorf("Expected 'true' (AzureCluster %s is reconciled by legacy release), got 'false' "+
				"(AzureCluster %s is not reconciled by a legacy release).",
				clusterName,
				clusterName)
		}
	}
}

func TestAzureClusterWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var azureClusterWebhookHandler *azureclusterpkg.WebhookHandler
	{
		c := azureclusterpkg.WebhookHandlerConfig{
			BaseDomain: env.BaseDomain(),
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			Location:   "testregion", // TODO: get from installation base domain
			CtrlReader: ctrlClient,
			Logger:     logger,
		}
		azureClusterWebhookHandler, err = azureclusterpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var azureClusterList capz.AzureClusterList
	err = ctrlClient.List(ctx, &azureClusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureCluster := range azureClusterList.Items {
		// Test mutating webhook, on create
		_, err = azureClusterWebhookHandler.OnCreateMutate(ctx, &azureCluster)
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on create
		err = azureClusterWebhookHandler.OnCreateValidate(ctx, &azureCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster := azureCluster.DeepCopy()
		updatedCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test mutating webhook, on update
		_, err = azureClusterWebhookHandler.OnUpdateMutate(ctx, &azureCluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on update
		err = azureClusterWebhookHandler.OnUpdateValidate(ctx, &azureCluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}
	}
}
