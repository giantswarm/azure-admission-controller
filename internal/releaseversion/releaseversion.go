package releaseversion

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func Validate(ctx context.Context, g8sclient versioned.Interface, oldVersion semver.Version, newVersion semver.Version) (bool, error) {
	if oldVersion.Equals(newVersion) {
		return true, nil
	}

	availableReleases, err := availableReleases(ctx, g8sclient)
	if err != nil {
		return false, err
	}

	// Check if old and new versions are valid.
	if !included(availableReleases, newVersion) {
		return false, microerror.Maskf(errors.InvalidReleaseError, "release %s was not found in this installation", newVersion)
	}

	// Downgrades are not allowed.
	if newVersion.LT(oldVersion) {
		return false, microerror.Maskf(errors.InvalidOperationError, "downgrading is not allowed (attempted to downgrade from %s to %s)", oldVersion, newVersion)
	}

	// Check if either version is an alpha one.
	if isAlphaRelease(oldVersion.String()) || isAlphaRelease(newVersion.String()) {
		return false, microerror.Maskf(errors.InvalidOperationError, "It is not possible to upgrade to or from an alpha release")
	}

	if oldVersion.Major != newVersion.Major || oldVersion.Minor != newVersion.Minor {
		// The major or minor version is changed. We support this only for sequential minor releases (no skip allowed).
		for _, release := range availableReleases {
			if release.EQ(oldVersion) || release.EQ(newVersion) {
				continue
			}
			// Look for a release with higher major or higher minor than the oldVersion and is LT the newVersion
			if release.GT(oldVersion) && release.LT(newVersion) &&
				(oldVersion.Major != release.Major || oldVersion.Minor != release.Minor) &&
				(newVersion.Major != release.Major || newVersion.Minor != release.Minor) {
				// Skipped one major or minor release.
				return false, microerror.Maskf(errors.InvalidOperationError, "Upgrading from %s to %s is not allowed (skipped %s)", oldVersion, newVersion, release)
			}
		}
	}

	return true, nil
}

func availableReleases(ctx context.Context, g8sclient versioned.Interface) ([]*semver.Version, error) {
	releaseList, err := g8sclient.ReleaseV1alpha1().Releases().List(ctx, v1.ListOptions{})
	if err != nil {
		return []*semver.Version{}, microerror.Mask(err)
	}

	var ret []*semver.Version
	for _, release := range releaseList.Items {
		parsed, err := semver.ParseTolerant(release.Name)
		if err != nil {
			return []*semver.Version{}, microerror.Maskf(errors.InvalidReleaseError, "Unable to parse release %s to a semver.Release", release.Name)
		}
		ret = append(ret, &parsed)
	}

	return ret, nil
}

func included(releases []*semver.Version, release semver.Version) bool {
	for _, r := range releases {
		if r.EQ(release) {
			return true
		}
	}

	return false
}

func isAlphaRelease(release string) bool {
	return strings.Contains(release, "alpha")
}
