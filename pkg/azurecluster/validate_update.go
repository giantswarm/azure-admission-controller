package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/v2/internal/errors"
	"github.com/giantswarm/azure-admission-controller/v2/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/v2/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/v2/pkg/key"
)

func (h *WebhookHandler) OnUpdateValidate(ctx context.Context, oldObject interface{}, object interface{}) error {
	azureClusterNewCR, err := key.ToAzureClusterPtr(object)
	if err != nil {
		return microerror.Mask(err)
	}
	if !azureClusterNewCR.GetDeletionTimestamp().IsZero() {
		h.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
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

	return h.validateRelease(ctx, azureClusterOldCR, azureClusterNewCR)
}

func (h *WebhookHandler) validateRelease(ctx context.Context, azureClusterOldCR *capz.AzureCluster, azureClusterNewCR *capz.AzureCluster) error {
	oldClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from the AzureCluster being updated")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from applied AzureCluster")
	}

	return releaseversion.Validate(ctx, h.ctrlClient, oldClusterVersion, newClusterVersion)
}
