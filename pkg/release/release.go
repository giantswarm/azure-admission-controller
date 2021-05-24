package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	gocache "github.com/patrickmn/go-cache"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// Cache of release components
	// key: release version
	// value: map "component name": "component version"
	releaseComponentsCache = gocache.New(24*time.Hour, 24*time.Hour)
)

func getComponentVersionsFromReleaseFromAPI(ctx context.Context, ctrlClient client.Client, releaseVersion string) (map[string]string, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}

	// Retrieve the `Release` CR.
	release := &releasev1alpha1.Release{}
	{
		err := ctrlClient.Get(ctx, client.ObjectKey{Name: releaseVersion, Namespace: "default"}, release)
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

// GetComponentVersionsFromRelease returns a map that contains all release components mapped to
// their respective versions which are included in that release.
//
// Since this function is called very often and it is calling Kubernetes API Server in the
// management cluster, the release components are cached in memory for 24h. This should not be a
// problem because versions of components in a release are idempotent, i.e. they do not change.
func GetComponentVersionsFromRelease(ctx context.Context, ctrlClient client.Client, releaseVersion string) (map[string]string, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}
	var err error

	var components map[string]string
	{
		cachedComponents, ok := releaseComponentsCache.Get(releaseVersion)

		if ok {
			// Release components are found in cache
			components = cachedComponents.(map[string]string)
		} else {
			components, err = getComponentVersionsFromReleaseFromAPI(ctx, ctrlClient, releaseVersion)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			// Save in cache
			releaseComponentsCache.Set(releaseVersion, components, gocache.DefaultExpiration)
		}
	}

	return components, nil
}

// ContainsAzureOperator checks if the specified release contains azure-operator.
//
// In order to perform the check, this function is calling GetComponentVersionsFromRelease function,
// which is caching obtained components in memory. See GetComponentVersionsFromRelease docs for
// more info about the caching.
func ContainsAzureOperator(ctx context.Context, ctrlClient client.Client, releaseVersion string) (bool, error) {
	componentVersions, err := GetComponentVersionsFromRelease(ctx, ctrlClient, releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if componentVersions["azure-operator"] == "" {
		return false, nil
	}

	return true, nil
}
