// +build k8srequired

package createcluster

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/crd"
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	expcapiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/integration/values"
)

const (
	prodCatalogName = "control-plane-catalog"
	testCatalogName = "control-plane-test-catalog"
	// API Groups for upstream Cluster API types.
	giantswarmCoreAPIGroup             = "core.giantswarm.io"
	clusterAPIGroup                    = "cluster.x-k8s.io"
	infrastructureAPIGroup             = "infrastructure.cluster.x-k8s.io"
	experimentalClusterAPIGroup        = "exp.cluster.x-k8s.io"
	experimentalInfrastructureAPIGroup = "exp.infrastructure.cluster.x-k8s.io"
	securityAPIGroup                   = "security.giantswarm.io"
)

func TestCreateCluster(t *testing.T) {
	var err error

	ctx := context.Background()

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatal(err)
	}

	var appTest apptest.Interface
	{
		runtimeScheme := runtime.NewScheme()
		appSchemeBuilder := runtime.SchemeBuilder{
			applicationv1alpha1.AddToScheme,
			apiextensionsv1.AddToScheme,
			capiv1alpha3.AddToScheme,
			capzv1alpha3.AddToScheme,
			expcapiv1alpha3.AddToScheme,
			expcapzv1alpha3.AddToScheme,
			securityv1alpha1.AddToScheme,
			corev1.AddToScheme,
			corev1alpha1.AddToScheme,
		}
		err = appSchemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			t.Fatal(err)
		}
		appTest, err = apptest.New(apptest.Config{
			KubeConfigPath: env.KubeConfig(),
			Logger:         logger,
			Scheme:         runtimeScheme,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err = appTest.EnsureCRDs(ctx, getRequiredCRDs())
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		valuesYAML, err := values.YAML(env.AzureClientID(), env.AzureClientSecret(), env.AzureSubscriptionID(), env.AzureTenantID())
		if err != nil {
			t.Fatal(err)
		}

		apps := []apptest.App{
			{
				CatalogName:   prodCatalogName,
				Name:          "cert-manager-app",
				Namespace:     metav1.NamespaceDefault,
				Version:       "2.3.1",
				WaitForDeploy: true,
			},
			{
				CatalogName:   testCatalogName,
				Name:          "azure-admission-controller",
				Namespace:     metav1.NamespaceDefault,
				SHA:           env.CircleSHA(),
				ValuesYAML:    valuesYAML,
				WaitForDeploy: true,
			},
		}
		err = appTest.InstallApps(ctx, apps)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = ensureCRsExist(ctx, appTest.CtrlClient())
	_ = deleteCRs(ctx, appTest.CtrlClient())
	if err != nil {
		t.Log(microerror.JSON(err))
		t.Fatal(err)
	}
}

func getRequiredCRDs() []*apiextensionsv1.CustomResourceDefinition {
	return []*apiextensionsv1.CustomResourceDefinition{
		corev1alpha1.NewAzureClusterConfigCRD(),
		providerv1alpha1.NewAzureConfigCRD(),
		crd.LoadV1(infrastructureAPIGroup, "AzureCluster"),
		crd.LoadV1(infrastructureAPIGroup, "AzureMachine"),
		crd.LoadV1(experimentalInfrastructureAPIGroup, "AzureMachinePool"),
		crd.LoadV1(clusterAPIGroup, "Cluster"),
		crd.LoadV1(experimentalClusterAPIGroup, "MachinePool"),
		crd.LoadV1(securityAPIGroup, "Organization"),
		crd.LoadV1(giantswarmCoreAPIGroup, "Spark"),
	}
}

func ensureCRsExist(ctx context.Context, client client.Client) error {
	return filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			bs, err := ioutil.ReadFile(path)
			if err != nil {
				return microerror.Mask(err)
			}

			o, err := loadCR(bs)
			if err != nil {
				return microerror.Mask(err)
			}

			err = client.Create(ctx, o)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	})
}

func deleteCRs(ctx context.Context, client client.Client) error {
	return filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			bs, err := ioutil.ReadFile(path)
			if err != nil {
				return microerror.Mask(err)
			}

			o, err := loadCR(bs)
			if err != nil {
				return microerror.Mask(err)
			}

			err = client.Delete(ctx, o)
			if apierrors.IsNotFound(err) {
				// Ok
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	})
}

func loadCR(bs []byte) (runtime.Object, error) {
	var err error
	var obj runtime.Object

	// First parse kind.
	typeMeta := &metav1.TypeMeta{}
	err = yaml.Unmarshal(bs, typeMeta)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Then construct correct CR object.
	switch typeMeta.Kind {
	case "Namespace":
		obj = new(corev1.Namespace)
	case "NamespaceList":
		obj = new(corev1.NamespaceList)
	case "Organization":
		obj = new(securityv1alpha1.Organization)
	case "Cluster":
		obj = new(capiv1alpha3.Cluster)
	case "MachinePool":
		obj = new(expcapiv1alpha3.MachinePool)
	case "AzureCluster":
		obj = new(capzv1alpha3.AzureCluster)
	case "AzureMachine":
		obj = new(capzv1alpha3.AzureMachine)
	case "AzureMachinePool":
		obj = new(expcapzv1alpha3.AzureMachinePool)
	case "Spark":
		obj = new(corev1alpha1.Spark)
	default:
		return nil, microerror.Maskf(unknownKindError, "error while unmarshalling the CR read from file, kind: %s", typeMeta.Kind)
	}

	// ...and unmarshal the whole object.
	err = yaml.Unmarshal(bs, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return obj, nil
}
