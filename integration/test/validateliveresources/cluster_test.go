// +build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	clusterpkg "github.com/giantswarm/azure-admission-controller/pkg/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
)

func TestClusterFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewCtrlClient(t)

	var clusterList capi.ClusterList
	err := ctrlClient.List(ctx, &clusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, cluster := range clusterList.Items {
		ownerClusterGetter := func(metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			return capi.Cluster{}, false, nil
		}

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &cluster, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			clusterName := fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name)
			t.Errorf("Expected 'true' (Cluster %s is reconciled by legacy release), got 'false' "+
				"(Cluster %s is not reconciled by a legacy release).",
				clusterName,
				clusterName)
		}
	}
}

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
		// Test mutating webhook, on create
		_, err = clusterWebhookHandler.OnCreateMutate(ctx, &cluster)
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on create
		err = clusterWebhookHandler.OnCreateValidate(ctx, &cluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster := cluster.DeepCopy()
		updatedCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test mutating webhook, on update
		_, err = clusterWebhookHandler.OnUpdateMutate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on update
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}
	}
}
