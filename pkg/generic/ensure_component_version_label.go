package generic

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func EnsureComponentVersionLabel(ctx context.Context, ctrlClient client.Client, meta metav1.Object, componentName string, labelName string) (*mutator.PatchOperation, error) {
	var err error
	if meta.GetLabels()[labelName] == "" {
		releaseVersion := meta.GetLabels()[label.ReleaseVersion]
		if releaseVersion == "" {
			// CR has no release version label set. Try to get the cluster version from the `Cluster` CR.
			releaseVersion, err = getReleaseVersionFromCluster(ctx, ctrlClient, meta)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		// Release CR always starts with a "v".
		if !strings.HasPrefix(releaseVersion, "v") {
			releaseVersion = fmt.Sprintf("v%s", releaseVersion)
		}

		// Retrieve the `Release` CR.
		release := &releasev1alpha1.Release{}
		{
			err := ctrlClient.Get(ctx, client.ObjectKey{Name: releaseVersion, Namespace: "default"}, release)
			if apierrors.IsNotFound(err) {
				return nil, microerror.Maskf(errors.InvalidOperationError, "Looking for Release %s but it was not found. Can't continue.", releaseVersion)
			} else if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		// Search the desired component.
		var componentVersion string
		{
			for _, component := range release.Spec.Components {
				if component.Name == componentName {
					componentVersion = component.Version
					break
				}
			}
		}

		if componentVersion == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Cannot find component %q in Release %s. Can't continue.", componentName, releaseVersion)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(labelName)), componentVersion), nil
	}

	return nil, nil
}
