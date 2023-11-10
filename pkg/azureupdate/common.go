package azureupdate

import (
	corev1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func getFakeCtrlClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	err := v1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = corev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capiexp.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capzexp.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capi.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = capz.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = providerv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = releasev1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	builder := fake.NewClientBuilder().WithScheme(scheme)

	return builder.Build(), nil
}
