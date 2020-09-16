module github.com/giantswarm/azure-admission-controller

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions v0.4.20
	github.com/giantswarm/k8sclient/v3 v3.1.2
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/tools v0.0.0-20200706234117-b22de6825cf7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/cluster-api v0.3.8
	sigs.k8s.io/controller-runtime v0.6.3
)
