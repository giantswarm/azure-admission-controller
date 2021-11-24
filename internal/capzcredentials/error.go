package capzcredentials

import "github.com/giantswarm/microerror"

var invalidObjectMetaError = &microerror.Error{
	Kind: "invalidObjectMetaError",
}

var missingIdentityRefError = &microerror.Error{
	Kind: "missingIdentityRefError",
}
