package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FirstCAPIRelease is the first GS release that runs on CAPI controllers
	FirstCAPIRelease = "20.0.0-v1alpha3"
)

func IsCAPIRelease(meta metav1.Object) (bool, error) {
	if meta.GetLabels()[label.ReleaseVersion] == "" {
		return false, nil
	}
	releaseVersion, err := ReleaseVersion(meta)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse release version from object")
	}
	return IsCAPIVersion(releaseVersion)
}

// IsCAPIVersion returns whether a given releaseVersion is using CAPI controllers
func IsCAPIVersion(releaseVersion *semver.Version) (bool, error) {
	CAPIVersion, err := semver.New(FirstCAPIRelease)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to get first CAPI release version")
	}
	return releaseVersion.GE(*CAPIVersion), nil
}

func ReleaseVersion(meta metav1.Object) (*semver.Version, error) {
	version, ok := meta.GetLabels()[label.ReleaseVersion]
	if !ok {
		return nil, microerror.Maskf(parsingFailedError, "unable to get release version from Object %s", meta.GetName())
	}
	version = strings.TrimPrefix(version, "v")
	return semver.New(version)
}

func ClusterExists(ctx context.Context, ctrlClient client.Client, object metav1.Object) error {
	clusters := &capi.ClusterList{}
	err := ctrlClient.List(ctx, clusters)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(clusters.Items) > 0 {
		for _, cluster := range clusters.Items {
			if object.GetName() == cluster.Name {
				return microerror.Maskf(notAllowedError, fmt.Sprintf("Cluster %s/%s already exists", cluster.Namespace, cluster.Name))
			}
		}
	}
	return nil
}

func AzureClusterExists(ctx context.Context, ctrlClient client.Client, object metav1.Object) error {
	azureClusters := &capz.AzureClusterList{}
	err := ctrlClient.List(ctx, azureClusters)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(azureClusters.Items) > 0 {
		for _, cluster := range azureClusters.Items {
			if object.GetName() == cluster.Name {
				return microerror.Maskf(notAllowedError, fmt.Sprintf("AzureCluster %s/%s already exists", cluster.Namespace, cluster.Name))
			}
		}
	}
	return nil
}
