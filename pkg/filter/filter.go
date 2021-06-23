package filter

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

// IsObjectReconciledByLegacyRelease checks if the object is reconciled by an operator which is the
// part of a legacy Giant Swarm release (a release that does not have Cluster API controllers).
func IsObjectReconciledByLegacyRelease(ctx context.Context, ctrlCache client.Reader, objectMeta metav1.Object) (bool, error) {
	// Try to get release label from the CR.
	releaseVersionLabel, err := release.GetReleaseVersionLabel(ctx, ctrlCache, objectMeta)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if releaseVersionLabel == "" {
		// We cannot find out which release this object cluster belongs to.
		return false, nil
	}

	// Now when we have release version for the CR, let's check if this is a legacy release.
	return release.IsLegacyRelease(ctx, ctrlCache, releaseVersionLabel)
}
