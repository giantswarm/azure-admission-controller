package filter

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

// IsObjectReconciledByLegacyRelease checks if the object is reconciled by an operator which is the
// part of a legacy Giant Swarm release (a release that does not have Cluster API controllers).
func IsObjectReconciledByLegacyRelease(ctx context.Context, logger micrologger.Logger, ctrlReader client.Reader, object metav1.ObjectMetaAccessor, ownerClusterGetter generic.OwnerClusterGetter) (bool, error) {
	// Try to get release from the CR.
	releaseCR, ok, err := release.TryFindReleaseForObject(ctx, ctrlReader, object, ownerClusterGetter)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !ok {
		// Release not found for the object.
		logger.Debugf(ctx, "Object not reconciled by a legacy release (release not found).")
		return false, nil
	}

	// Now when we have release CR, let's check if this is a legacy release.
	isLegacy := release.IsLegacy(releaseCR)
	if isLegacy {
		logger.Debugf(ctx, "Object is reconciled by a legacy release.")
	} else {
		logger.Debugf(ctx, "Object not reconciled by a legacy release.")
	}

	return isLegacy, nil
}
