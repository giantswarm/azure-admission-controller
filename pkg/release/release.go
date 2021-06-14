package release

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
//
// In order to perform the check, this function is calling GetComponentVersionsFromRelease function,
// which is caching obtained components in memory. See GetComponentVersionsFromRelease docs for
// more info about the caching.
func ContainsAzureOperator(ctx context.Context, ctrlClient client.Client, logger micrologger.Logger, releaseVersion string) (bool, error) {
	componentVersions, err := GetComponentVersionsFromRelease(ctx, ctrlClient, logger, releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if componentVersions["azure-operator"] == "" {
		return false, nil
	}

	return true, nil
}
