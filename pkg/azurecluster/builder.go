package azurecluster

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
)

type BuilderOption func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster

func Location(location string) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		azureCluster.Spec.Location = location
		return azureCluster
	}
}

func ControlPlaneEndpoint(controlPlaneEndpointHost string, controlPlaneEndpointPort int32) BuilderOption {
	return func(azureCluster *capzv1alpha3.AzureCluster) *capzv1alpha3.AzureCluster {
		azureCluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpointHost
		azureCluster.Spec.ControlPlaneEndpoint.Port = controlPlaneEndpointPort
		return azureCluster
	}
}

func BuildAzureCluster(clusterName string, opts ...BuilderOption) *capzv1alpha3.AzureCluster {
	azureCluster := &capzv1alpha3.AzureCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureCluster",
			APIVersion: capzv1alpha3.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"cluster.x-k8s.io/cluster-name":        clusterName,
				"giantswarm.io/cluster":                clusterName,
				"giantswarm.io/organization":           "giantswarm",
				"release.giantswarm.io/version":        "13.0.0-alpha3",
			},
		},
		Spec: capzv1alpha3.AzureClusterSpec{
			ResourceGroup: clusterName,
			Location:      "westeurope",
			ControlPlaneEndpoint: v1alpha3.APIEndpoint{
				Host: "api.gigantic.io",
				Port: 8080,
			},
		},
	}

	for _, opt := range opts {
		opt(azureCluster)
	}

	return azureCluster
}

func BuildAzureClusterAsJson(clusterName string, opts ...BuilderOption) []byte {
	azureCluster := BuildAzureCluster(clusterName, opts...)

	byt, _ := json.Marshal(azureCluster)

	return byt
}
