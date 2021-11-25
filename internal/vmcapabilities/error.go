package vmcapabilities

import "github.com/giantswarm/microerror"

var credentialsNotFoundError = &microerror.Error{
	Kind: "credentialsNotFoundError",
}

// IsCredentialsNotFoundError asserts credentialsNotFoundError.
func IsCredentialsNotFoundError(err error) bool {
	return microerror.Cause(err) == credentialsNotFoundError
}

var invalidRequestError = &microerror.Error{
	Kind: "invalidRequestError",
}

// IsInvalidRequest asserts invalidRequestError.
func IsInvalidRequest(err error) bool {
	return microerror.Cause(err) == invalidRequestError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidUpstreamResponseError = &microerror.Error{
	Kind: "invalidUpstreamResponseError",
}

// IsInvalidUpstreamResponse asserts invalidUpstreamResponseError.
func IsInvalidUpstreamResponse(err error) bool {
	return microerror.Cause(err) == invalidUpstreamResponseError
}

var skuNotFoundError = &microerror.Error{
	Kind: "skuNotFoundError",
}

// IsSkuNotFoundError asserts skuNotFoundError.
func IsSkuNotFoundError(err error) bool {
	return microerror.Cause(err) == skuNotFoundError
}

var missingValueError = &microerror.Error{
	Kind: "missingValueError",
}

var tooManyCredentialsError = &microerror.Error{
	Kind: "tooManyCredentialsError",
}

// IsTooManyCredentials asserts tooManyCredentialsError.
func IsTooManyCredentials(err error) bool {
	return microerror.Cause(err) == tooManyCredentialsError
}
