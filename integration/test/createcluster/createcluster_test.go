// +build k8srequired

package createcluster

import (
	"context"
	"fmt"
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
	clusterAPIGroup                    = "cluster.x-k8s.io"
	infrastructureAPIGroup             = "infrastructure.cluster.x-k8s.io"
	experimentalClusterAPIGroup        = "exp.cluster.x-k8s.io"
	experimentalInfrastructureAPIGroup = "exp.infrastructure.cluster.x-k8s.io"
	securityAPIGroup                   = "security.giantswarm.io"
)

var (
	appTest apptest.Interface
	logger  micrologger.Logger
)

func TestCreateCluster(t *testing.T) {
	var err error

	ctx := context.Background()

	{
		logger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			panic(err.Error())
		}
	}

	{
		runtimeScheme := runtime.NewScheme()
		appSchemeBuilder := runtime.SchemeBuilder{
			applicationv1alpha1.AddToScheme,
			apiextensionsv1.AddToScheme,
			capiv1alpha3.AddToScheme,
			capzv1alpha3.AddToScheme,
			securityv1alpha1.AddToScheme,
			corev1.AddToScheme,
		}
		err = appSchemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			panic(err)
		}
		appTest, err = apptest.New(apptest.Config{
			KubeConfigPath: env.KubeConfig(),
			Logger:         logger,
			Scheme:         runtimeScheme,
		})
		if err != nil {
			panic(err.Error())
		}
	}

	{
		err = appTest.EnsureCRDs(ctx, getRequiredCRDs())
		if err != nil {
			logger.LogCtx(ctx, "level", "error", "message", "failed ensuring crds", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
		}
	}

	{
		valuesYAML, err := values.YAML(env.AzureClientID(), env.AzureClientSecret(), env.AzureTenantID(), env.AzureSubscriptionID())
		if err != nil {
			logger.LogCtx(ctx, "level", "error", "message", "install apps failed", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
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
			logger.LogCtx(ctx, "level", "error", "message", "install apps failed", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
		}
	}

	err = ensureCRsExist(appTest.CtrlClient(), []string{"namespaces.yaml", "organization.yaml", "azurecluster.yaml", "cluster.yaml"})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	defer deleteCRs(appTest.CtrlClient(), []string{"namespaces.yaml", "organization.yaml", "azurecluster.yaml", "cluster.yaml"})
}

func getRequiredCRDs() []*apiextensionsv1.CustomResourceDefinition {
	clusterCRD := crd.LoadV1(clusterAPIGroup, "Cluster")
	azureClusterCRD := crd.LoadV1(infrastructureAPIGroup, "AzureCluster")
	azureMachineCRD := crd.LoadV1(infrastructureAPIGroup, "AzureMachine")
	machinePoolCRD := crd.LoadV1(experimentalClusterAPIGroup, "MachinePool")
	azureMachinePoolCRD := crd.LoadV1(experimentalInfrastructureAPIGroup, "AzureMachinePool")
	organizationCRD := crd.LoadV1(securityAPIGroup, "Organization")

	return []*apiextensionsv1.CustomResourceDefinition{
		corev1alpha1.NewAzureClusterConfigCRD(),
		corev1alpha1.NewSparkCRD(),
		providerv1alpha1.NewAzureConfigCRD(),

		clusterCRD,
		azureClusterCRD,
		azureMachineCRD,
		machinePoolCRD,
		azureMachinePoolCRD,
		organizationCRD,
	}
}

func ensureCRsExist(client client.Client, inputFiles []string) error {
	for _, f := range inputFiles {
		o, err := loadCR(f)
		if err != nil {
			return err
		}

		err = client.Create(context.Background(), o)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteCRs(client client.Client, inputFiles []string) {
	for _, f := range inputFiles {
		o, err := loadCR(f)
		if err != nil {
			panic(err.Error())
		}

		err = client.Delete(context.Background(), o)
		if err != nil {
			panic(err.Error())
		}
	}
}

func loadCR(fName string) (runtime.Object, error) {
	var err error
	var obj runtime.Object

	var bs []byte
	{
		bs, err = ioutil.ReadFile(filepath.Join("testdata", fName))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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
	default:
		return nil, microerror.Maskf(unknownKindError, "kind: %s", typeMeta.Kind)
	}

	// ...and unmarshal the whole object.
	err = yaml.Unmarshal(bs, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return obj, nil
}
