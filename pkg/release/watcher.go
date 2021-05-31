package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	defaultResyncPeriod = 5 * time.Minute
)

const (
	DisableMetricsServing = "0"
)

type watchGetter struct {
	manager  manager.Manager
	releases map[string]releasev1alpha1.Release
}

// WatchGetter creates a Release Getter that is watching API events on Release
// objects. It provides a near-realtime up-to-date cache to Releases available.
func WatchGetter(cfg *rest.Config) (Getter, error) {
	var err error

	wg := &watchGetter{
		releases: make(map[string]releasev1alpha1.Release, 0),
	}

	var mgr manager.Manager
	{
		o := manager.Options{
			// MetricsBindAddress is set to 0 in order to disable it. We do this
			// ourselves.
			MetricsBindAddress: DisableMetricsServing,
			Namespace:          metav1.NamespaceAll,
			SyncPeriod:         to.DurationP(defaultResyncPeriod),
		}

		mgr, err = manager.New(cfg, o)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	_, err = controller.New("releases-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(wg.reconcile),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = mgr.Start(signals.SetupSignalHandler())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	wg.manager = mgr

	err = wg.warmUp()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return wg, nil
}

// Get returns cached Release or an error if it does not exist for the
// requested version.
func (wg *watchGetter) Get(ver string) (releasev1alpha1.Release, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(ver, "v") {
		ver = fmt.Sprintf("v%s", ver)
	}

	r, exists := wg.releases[ver]
	if exists {
		return r, nil
	}

	return releasev1alpha1.Release{}, microerror.Maskf(releaseNotFoundError, "Release for version %q doesn't exist", ver)
}

// reconcile implements event handling for Release. If referenced object has
// `deletionTimestamp` set, the internally cached object is also removed.
// Otherwise the latest version is updated to cache.
func (wg *watchGetter) reconcile(req reconcile.Request) (reconcile.Result, error) {
	c := wg.manager.GetClient()
	o := &releasev1alpha1.Release{}
	err := c.Get(context.Background(), req.NamespacedName, o)
	if err != nil {
		return reconcile.Result{}, microerror.Mask(err)
	}

	if o.DeletionTimestamp.IsZero() {
		wg.releases[o.Name] = *o
	} else {
		delete(wg.releases, o.Name)
	}

	return reconcile.Result{}, nil
}

// warmUp Lists all Releases and stores them in internal wg.releases.
func (wg *watchGetter) warmUp() error {
	rl := &releasev1alpha1.ReleaseList{}
	c := wg.manager.GetClient()
	err := c.List(context.Background(), rl, nil)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, r := range rl.Items {
		// Only cache releases that have not been deleted yet.
		if r.DeletionTimestamp.IsZero() {
			wg.releases[r.Name] = r
		}
	}

	return nil
}
