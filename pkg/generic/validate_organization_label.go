package generic

import (
	"context"

	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func ValidateOrganizationLabelUnchanged(old, new metav1.Object) error {
	if _, exists := old.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(errors.NotFoundError, "old CR doesn't contain Organization label %#q", label.Organization)
	}

	if _, exists := new.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(errors.NotFoundError, "new CR doesn't contain Organization label %#q", label.Organization)
	}

	if old.GetLabels()[label.Organization] != new.GetLabels()[label.Organization] {
		return microerror.Maskf(errors.InvalidOperationError, "Organization label %#q must not be changed", label.Organization)
	}

	return nil
}

func ValidateOrganizationLabelContainsExistingOrganization(ctx context.Context, obj metav1.Object, ctrlClient client.Client) error {
	organizationName, ok := obj.GetLabels()[label.Organization]
	if !ok {
		return microerror.Maskf(errors.NotFoundError, "CR doesn't contain Organization label %#q", label.Organization)
	}

	organization := &securityv1alpha1.Organization{}
	err := ctrlClient.Get(ctx, client.ObjectKey{Name: organizationName}, organization)
	if err != nil {
		return microerror.Maskf(errors.InvalidOperationError, "Organization label %#q must contain an existing organization, got %#q", label.Organization, organizationName)
	}

	return nil
}
