package azuremachine

import (
	"strings"

	"github.com/giantswarm/microerror"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateFailureDomain(azureMachine capzv1alpha3.AzureMachine, supportedAZs []string) error {
	// No failure domain specified.
	if azureMachine.Spec.FailureDomain == nil {
		return nil
	}

	for _, az := range supportedAZs {
		if *azureMachine.Spec.FailureDomain == az {
			// Failure Domain is valid.
			return nil
		}
	}

	return microerror.Maskf(invalidOperationError, "AzureMachine.Spec.FailureDomain is invalid for machine type %s in location %s. Supported AZs are %s", azureMachine.Spec.VMSize, azureMachine.Spec.Location, strings.Join(supportedAZs, ", "))
}

func validateFailureDomainUnchanged(old capzv1alpha3.AzureMachine, new capzv1alpha3.AzureMachine) error {
	// Was unspecified, stays unspecified.
	if old.Spec.FailureDomain == nil && new.Spec.FailureDomain == nil {
		return nil
	}

	// Was set and got blanked, was blank and got set, was set and got changed.
	if old.Spec.FailureDomain == nil && new.Spec.FailureDomain != nil ||
		old.Spec.FailureDomain != nil && new.Spec.FailureDomain == nil ||
		*old.Spec.FailureDomain != *new.Spec.FailureDomain {
		return microerror.Maskf(invalidOperationError, "AzureMachine.Spec.FailureDomain can't be changed")
	}

	return nil
}
