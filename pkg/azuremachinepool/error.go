package azuremachinepool

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

var invalidOperationError = &microerror.Error{
	Kind: "invalidOperationError",
}

// IsInvalidOperationError asserts invalidOperationError.
func IsInvalidOperationError(err error) bool {
	return microerror.Cause(err) == invalidOperationError
}

var parsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == parsingFailedError
}
