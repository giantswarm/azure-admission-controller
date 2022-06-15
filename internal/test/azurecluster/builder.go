package azurecluster

import (
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v6/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/azure-admission-controller/internal/test"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

type BuilderOption func(azureCluster *capz.AzureCluster) *capz.AzureCluster

func Name(name string) BuilderOption {
	return func(azureCluster *capz.AzureCluster) *capz.AzureCluster {
		azureCluster.ObjectMeta.Name = name
		azureCluster.Labels[capi.ClusterLabelName] = name
		azureCluster.Labels[label.Cluster] = name
		azureCluster.Spec.ResourceGroup = name
		azureCluster.Spec.ControlPlaneEndpoint.Host = fmt.Sprintf("api.%s.k8s.test.westeurope.azure.gigantic.io", name)
		azureCluster.Spec.NetworkSpec.APIServerLB.Name = key.APIServerLBName(name)
		azureCluster.Spec.NetworkSpec.APIServerLB.FrontendIPs = []capz.FrontendIP{
			{
				Name: key.APIServerLBFrontendIPName(name),
			},
		}
		return azureCluster
	}
}

func Labels(labels map[string]string) BuilderOption {
	return func(azureCluster *capz.AzureCluster) *capz.AzureCluster {
		for k, v := range labels {
			azureCluster.Labels[k] = v
		}
		return azureCluster
	}
}

func Location(location string) BuilderOption {
	return func(azureCluster *capz.AzureCluster) *capz.AzureCluster {
		azureCluster.Spec.Location = location
		return azureCluster
	}
}

func ControlPlaneEndpoint(controlPlaneEndpointHost string, controlPlaneEndpointPort int32) BuilderOption {
	return func(azureCluster *capz.AzureCluster) *capz.AzureCluster {
		azureCluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpointHost
		azureCluster.Spec.ControlPlaneEndpoint.Port = controlPlaneEndpointPort
		return azureCluster
	}
}

func WithDeletionTimestamp() BuilderOption {
	return func(azureCluster *capz.AzureCluster) *capz.AzureCluster {
		now := metav1.Now()
		azureCluster.ObjectMeta.SetDeletionTimestamp(&now)
		return azureCluster
	}
}

func BuildAzureCluster(opts ...BuilderOption) *capz.AzureCluster {
	clusterName := test.GenerateName()
	azureCluster := &capz.AzureCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureCluster",
			APIVersion: capz.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "org-giantswarm",
			Labels: map[string]string{
				label.AzureOperatorVersion: "5.0.0",
				capi.ClusterLabelName:      clusterName,
				label.Cluster:              clusterName,
				label.Organization:         "giantswarm",
				label.ReleaseVersion:       "13.0.0-alpha4",
			},
		},
		Spec: capz.AzureClusterSpec{
			AzureClusterClassSpec: capz.AzureClusterClassSpec{
				Location:         "westeurope",
				AzureEnvironment: "AzurePublicCloud",
			},
			NetworkSpec: capz.NetworkSpec{
				Subnets: capz.Subnets{
					capz.SubnetSpec{
						SubnetClassSpec: capz.SubnetClassSpec{
							Role: "control-plane",
						},
						Name: key.MasterSubnetName(clusterName),
					},
					capz.SubnetSpec{
						SubnetClassSpec: capz.SubnetClassSpec{
							Role: "node",
						},
						Name: clusterName,
					},
				},
				APIServerLB: capz.LoadBalancerSpec{
					Name: key.APIServerLBName(clusterName),
					LoadBalancerClassSpec: capz.LoadBalancerClassSpec{
						SKU:  capz.SKU(key.APIServerLBSKU()),
						Type: capz.LBType(key.APIServerLBType()),
					},
					FrontendIPs: []capz.FrontendIP{
						{
							Name: key.APIServerLBFrontendIPName(clusterName),
						},
					},
				},
				NodeOutboundLB: &capz.LoadBalancerSpec{},
			},
			ResourceGroup: clusterName,
			ControlPlaneEndpoint: capi.APIEndpoint{
				Host: fmt.Sprintf("api.%s.k8s.test.westeurope.azure.gigantic.io", clusterName),
				Port: 443,
			},
		},
	}

	for _, opt := range opts {
		opt(azureCluster)
	}

	return azureCluster
}

func BuildAzureClusterAsJson(opts ...BuilderOption) []byte {
	azureCluster := BuildAzureCluster(opts...)

	byt, _ := json.Marshal(azureCluster)

	return byt
}
