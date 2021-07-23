// +build liveinstallation

package validateliveresources

import (
	"context"
	"testing"

	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	clusterpkg "github.com/giantswarm/azure-admission-controller/pkg/cluster"
)

func TestClusterWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewCtrlClient(t)

	var clusterWebhookHandler *clusterpkg.WebhookHandler
	{
		c := clusterpkg.WebhookHandlerConfig{
			BaseDomain: env.BaseDomain(),
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			CtrlReader: ctrlClient,
			Logger:     logger,
		}
		clusterWebhookHandler, err = clusterpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var clusterList capi.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, cluster := range clusterList.Items {
		err = clusterWebhookHandler.OnCreateValidate(ctx, &cluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster := cluster.DeepCopy()

		updatedCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster.Annotations["test.giantswarm.io/dummy"] = "this is not really saved"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster.Labels["release.giantswarm.io/version"] = "123456789.123456789.123456789"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if !releaseversion.IsReleaseNotFoundError(err) {
			t.Fatalf("expected releaseNotFoundError, got error: %#v", err)
		}
	}
}
