package generic

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OwnerClusterGetter func(metav1.ObjectMetaAccessor) (capi.Cluster, bool, error)

func TryGetClusterName(object metav1.ObjectMetaAccessor) (string, bool) {
	if object.GetObjectMeta() == nil || object.GetObjectMeta().GetLabels() == nil {
		return "", false
	}
	labels := object.GetObjectMeta().GetLabels()

	// First let's try to get CAPI cluster name label.
	clusterName := labels[capi.ClusterLabelName]
	if clusterName == "" {
		// CAPI cluster name label not found, now let's try GS cluster ID label, which is basically
		// the same thing.
		clusterID := labels[label.Cluster]
		if clusterID != "" {
			// We found GS cluster ID, this is our cluster name.
			clusterName = clusterID
		}
	}

	return clusterName, clusterName != ""
}

func GetOwnerCluster(ctx context.Context, ctrlClient client.Client, object metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
	clusterName, ok := TryGetClusterName(object)
	if !ok {
		return capi.Cluster{}, false, nil
	}

	if object.GetObjectMeta() == nil {
		return capi.Cluster{}, false, nil
	}

	var cluster capi.Cluster
	key := client.ObjectKey{
		Namespace: object.GetObjectMeta().GetNamespace(),
		Name:      clusterName,
	}
	err := ctrlClient.Get(ctx, key, &cluster)
	if apierrors.IsNotFound(err) {
		return capi.Cluster{}, false, nil
	} else if err != nil {
		return capi.Cluster{}, false, microerror.Mask(err)
	}

	return cluster, true, nil
}
