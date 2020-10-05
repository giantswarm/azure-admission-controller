package azureupdate

import (
	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
)

func clusterIsUpgrading(cr *v1alpha1.AzureConfig) (bool, string) {
	for _, cond := range cr.Status.Cluster.Conditions {
		if cond.Type == conditionUpdating {
			return true, conditionUpdating
		}
		if cond.Type == conditionCreating {
			return true, conditionCreating
		}
	}

	return false, ""
}
