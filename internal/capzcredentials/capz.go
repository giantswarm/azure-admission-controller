package capzcredentials

import (
	"context"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetAzureCredentialsFromMetadata(ctx context.Context, ctrlClient client.Client, obj metav1.ObjectMeta) (string, string, string, string, error) {
	azureCluster, err := getAzureClusterFromMetadata(ctx, ctrlClient, obj)
	if err != nil {
		return "", "", "", "", microerror.Mask(err)
	}

	if azureCluster.Spec.IdentityRef == nil {
		return "", "", "", "", microerror.Maskf(missingIdentityRefError, "IdentiyRef was nil in AzureCluster %s/%s", azureCluster.Namespace, azureCluster.Name)
	}

	identity := capz.AzureClusterIdentity{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Namespace: azureCluster.Spec.IdentityRef.Namespace, Name: azureCluster.Spec.IdentityRef.Name}, &identity)
	if err != nil {
		return "", "", "", "", microerror.Mask(err)
	}

	secret := v1.Secret{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Namespace: identity.Spec.ClientSecret.Namespace, Name: identity.Spec.ClientSecret.Name}, &secret)
	if err != nil {
		return "", "", "", "", microerror.Mask(err)
	}

	return azureCluster.Spec.SubscriptionID, identity.Spec.ClientID, string(secret.Data["clientSecret"]), identity.Spec.TenantID, nil
}

func getAzureClusterFromMetadata(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*capz.AzureCluster, error) {
	// Check if "cluster.x-k8s.io/cluster-name" label is set.
	if obj.Labels[capi.ClusterLabelName] == "" {
		err := microerror.Maskf(invalidObjectMetaError, "Label %q must not be empty for object %q", capi.ClusterLabelName, obj.GetSelfLink())
		return nil, microerror.Mask(err)
	}

	return getAzureClusterByName(ctx, c, obj.Namespace, obj.Labels[capi.ClusterLabelName])
}

func getAzureClusterByName(ctx context.Context, c client.Client, namespace, name string) (*capz.AzureCluster, error) {
	azureCluster := &capz.AzureCluster{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, azureCluster); err != nil {
		return nil, microerror.Mask(err)
	}

	return azureCluster, nil
}
