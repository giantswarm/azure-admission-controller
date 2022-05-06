package azuremachine

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func validateFailureDomain(azureMachine capz.AzureMachine, supportedAZs []string, location string) error {
	// No failure domain specified.
	if azureMachine.Spec.FailureDomain == nil || *azureMachine.Spec.FailureDomain == "" {
		return nil
	}

	for _, az := range supportedAZs {
		if *azureMachine.Spec.FailureDomain == az {
			// Failure Domain is valid.
			return nil
		}
	}

	supportedAZsMsg := fmt.Sprintf("Location %#q supports Failure Domains %s for VM size %#q but got %#q", location, strings.Join(supportedAZs, ", "), azureMachine.Spec.VMSize, *azureMachine.Spec.FailureDomain)
	if len(supportedAZs) == 0 {
		supportedAZsMsg = fmt.Sprintf("Location %#q does not support specifying a Failure Domain for VM size %#q and the Failure Domain %#q was selected", location, azureMachine.Spec.VMSize, *azureMachine.Spec.FailureDomain)
		return microerror.Maskf(locationWithNoFailureDomainSupportError, supportedAZsMsg)
	}

	return microerror.Maskf(unsupportedFailureDomainError, supportedAZsMsg)
}

func validateFailureDomainUnchanged(old capz.AzureMachine, new capz.AzureMachine) error {
	// Was unspecified, stays unspecified.
	if old.Spec.FailureDomain == nil && new.Spec.FailureDomain == nil {
		return nil
	}

	// Allow changing from nil to "" and from "" to nil. They are synonyms.
	if old.Spec.FailureDomain == nil && new.Spec.FailureDomain != nil && *new.Spec.FailureDomain == "" ||
		old.Spec.FailureDomain != nil && *old.Spec.FailureDomain == "" && new.Spec.FailureDomain == nil {
		return nil
	}

	// Was set and got blanked, was blank and got set, was set and got changed.
	if old.Spec.FailureDomain == nil && new.Spec.FailureDomain != nil ||
		old.Spec.FailureDomain != nil && new.Spec.FailureDomain == nil ||
		*old.Spec.FailureDomain != *new.Spec.FailureDomain {
		return microerror.Maskf(failureDomainWasChangedError, "AzureMachine.Spec.FailureDomain can't be changed")
	}

	return nil
}
