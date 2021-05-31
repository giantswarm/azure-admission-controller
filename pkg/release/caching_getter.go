package release

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type cachingGetter struct {
	cache.Cache
}

func NewCachingGetter(cfg *rest.Config) (Getter, error) {
	scheme := runtime.NewScheme()
	err := releasev1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	opts := cache.Options{
		Scheme: scheme,
	}

	c, err := cache.New(cfg, opts)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &cachingGetter{Cache: c}, nil
}

func (cg *cachingGetter) Get(ver string) (releasev1alpha1.Release, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(ver, "v") {
		ver = fmt.Sprintf("v%s", ver)
	}

	r := releasev1alpha1.Release{}
	err := cg.Cache.Get(context.Background(), client.ObjectKey{Name: ver}, &r)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	return r, nil
}
