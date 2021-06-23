package release

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetReleaseVersionLabel gets release version string for the specified object.
//
// If the object itself does not have the release version label, the function will look for cluster
// name in the object labels, and then get the Cluster CR from which it will try to get the release
// version label.
func GetReleaseVersionLabel(ctx context.Context, ctrlCache client.Reader, objectMeta metav1.Object) (string, error) {
	// Try to get release label from the CR
	releaseVersionLabel := objectMeta.GetLabels()[label.ReleaseVersion]
	if releaseVersionLabel != "" {
		return releaseVersionLabel, nil
	}

	// Release label is not found on the CR, let's try to get it from owner Cluster CR.

	// First let's try to get CAPI cluster name label.
	clusterName := objectMeta.GetLabels()[capi.ClusterLabelName]
	if clusterName == "" {
		// CAPI cluster name label not found, now let's try GS cluster ID label, which is basically
		// the same thing.
		clusterID := objectMeta.GetLabels()[label.Cluster]
		if clusterID == "" {
			// We can't find out which cluster and release this CR belongs to.
			return "", nil
		}

		// We found GS cluster ID, this is our cluster name.
		clusterName = clusterID
	}

	// Now get the owner cluster by name, we will try to check if it has release label.
	cluster := &capi.Cluster{}
	key := client.ObjectKey{
		Namespace: objectMeta.GetNamespace(),
		Name:      clusterName,
	}
	err := ctrlCache.Get(ctx, key, cluster)
	if err != nil {
		return "", microerror.Mask(err)
	}

	releaseVersionLabel = cluster.Labels[label.ReleaseVersion]
	return releaseVersionLabel, nil
}

// IsLegacyRelease checks if the specified release is a legacy release, i.e. a release without
// Cluster API controllers.
func IsLegacyRelease(ctx context.Context, ctrlCache client.Reader, releaseVersion string) (bool, error) {
	releaseContainsAzureOperator, err := ContainsAzureOperator(ctx, ctrlCache, releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return releaseContainsAzureOperator, nil
}

func GetComponentVersionsFromRelease(ctx context.Context, ctrlReader client.Reader, releaseVersion string) (map[string]string, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}

	// Retrieve the `Release` CR.
	release := &releasev1alpha1.Release{}
	{
		err := ctrlReader.Get(ctx, client.ObjectKey{Name: releaseVersion}, release)
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(releaseNotFoundError, "Looking for Release %s but it was not found. Can't continue.", releaseVersion)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	ret := map[string]string{}
	// Search the desired component.
	for _, component := range release.Spec.Components {
		ret[component.Name] = component.Version
	}

	return ret, nil
}

// ContainsAzureOperator checks if the specified release contains azure-operator.
func ContainsAzureOperator(ctx context.Context, ctrlReader client.Reader, releaseVersion string) (bool, error) {
	componentVersions, err := GetComponentVersionsFromRelease(ctx, ctrlReader, releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if componentVersions["azure-operator"] == "" {
		return false, nil
	}

	return true, nil
}
