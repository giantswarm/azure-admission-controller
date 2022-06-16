package releaseversion

import (
	"github.com/blang/semver"
	"github.com/giantswarm/release-operator/v3/api/v1alpha1"
)

type release struct {
	Version *semver.Version
	CR      *v1alpha1.Release
}
