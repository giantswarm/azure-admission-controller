package azurecluster

import (
	"context"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capiutil "sigs.k8s.io/cluster-api/util"
)

func (a *Validator) ValidateUpdate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureClusterNewCR, err := key.ToAzureClusterPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	azureClusterOldCR, err := key.ToAzureClusterPtr(oldObject)
	if err != nil {
		return microerror.Mask(err)
	}

	err = azureClusterNewCR.ValidateUpdate(azureClusterOldCR)
	err = errors.IgnoreCAPIErrorForField("metadata.Name", err)
	err = errors.IgnoreCAPIErrorForField("spec.networkSpec.subnets", err)
	// TODO(axbarsan): Remove this once all the older clusters have it.
	err = errors.IgnoreCAPIErrorForField("spec.networkSpec.apiServerLB", err)
	err = errors.IgnoreCAPIErrorForField("spec.SubscriptionID", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(azureClusterOldCR, azureClusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpointUnchanged(*azureClusterOldCR, *azureClusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return a.validateRelease(ctx, azureClusterOldCR, azureClusterNewCR)
}

func (a *Validator) validateRelease(ctx context.Context, azureClusterOldCR *capz.AzureCluster, azureClusterNewCR *capz.AzureCluster) error {
	oldClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from the AzureCluster being updated")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from applied AzureCluster")
	}

	if !newClusterVersion.Equals(oldClusterVersion) {
		cluster, err := capiutil.GetOwnerCluster(ctx, a.ctrlClient, azureClusterNewCR.ObjectMeta)
		if err != nil {
			return microerror.Mask(err)
		}

		clusterCRReleaseVersion, err := semverhelper.GetSemverFromLabels(cluster.Labels)
		if err != nil {
			return microerror.Mask(err)
		}

		if !newClusterVersion.Equals(clusterCRReleaseVersion) {
			return microerror.Maskf(errors.InvalidOperationError, "AzureCluster release version must be set to the same release version as Cluster CR release version label")
		}
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}
