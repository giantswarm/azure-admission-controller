package vmcapabilities

import (
	"github.com/giantswarm/microerror"
)

var unknownVMTypeError = &microerror.Error{
	Kind: "unknownVMTypeError",
}

// IsUnknownVMTypeError asserts unknownVMTypeError.
func IsUnknownVMTypeError(err error) bool {
	return microerror.Cause(err) == unknownVMTypeError
}
