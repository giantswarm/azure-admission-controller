package azureupdate

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var cantChangeMasterCIDRError = &microerror.Error{
	Kind: "cantChangeMasterCIDRError",
}

// IsCantChangeMasterCIDR asserts cantChangeMasterCIDRError.
func IsCantChangeMasterCIDR(err error) bool {
	return microerror.Cause(err) == cantChangeMasterCIDRError
}
