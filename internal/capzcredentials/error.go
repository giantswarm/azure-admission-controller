package capzcredentials

import "github.com/giantswarm/microerror"

var invalidObjectMetaError = &microerror.Error{
	Kind: "invalidObjectMetaError",
}

var missingIdentityRefError = &microerror.Error{
	Kind: "missingIdentityRefError",
}

// IsMissingIdentityRef asserts missingIdentityRefError.
func IsMissingIdentityRef(err error) bool {
	return microerror.Cause(err) == missingIdentityRefError
}
