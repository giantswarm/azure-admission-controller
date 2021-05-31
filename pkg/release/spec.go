package release

import releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"

type Getter interface {
	Get(ver string) (releasev1alpha1.Release, error)
}
