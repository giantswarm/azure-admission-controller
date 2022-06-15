module github.com/giantswarm/azure-admission-controller

go 1.16

require (
	github.com/Azure/azure-sdk-for-go v58.1.0+incompatible
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/dyson/certman v0.2.1
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions-application v0.4.0
	github.com/giantswarm/apiextensions/v6 v6.0.0
	github.com/giantswarm/app/v5 v5.4.0
	github.com/giantswarm/apptest v0.10.3
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/release-operator/v3 v3.2.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.5.7
	github.com/stretchr/testify v1.7.1
	gomodules.xyz/jsonpatch/v2 v2.2.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.22.5
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	sigs.k8s.io/cluster-api v1.0.5
	sigs.k8s.io/cluster-api-provider-azure v1.0.2
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.3.0
)

require github.com/go-logr/logr v1.2.3 // indirect

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.34
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.6
	github.com/containerd/imgcrypt => github.com/containerd/imgcrypt v1.1.6
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.31
	github.com/caddyserver/caddy => github.com/caddyserver/caddy v1.0.5
	// Required to replace version with vulnerabilities.
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/stdr => github.com/go-logr/stdr v0.4.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	go.uber.org/goleak => go.uber.org/goleak v1.1.10
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.9.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.5
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.10.3
)
