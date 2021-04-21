package key

import (
	"fmt"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/microerror"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
)

const (
	ControlPlaneEndpointPort  = 443
	ClusterNetworkServiceCIDR = "172.31.0.0/16"
)

func GetControlPlaneEndpointHost(clusterName string, baseDomain string) string {
	return fmt.Sprintf("api.%s.%s", clusterName, baseDomain)
}

func ServiceDomain() string {
	return "cluster.local"
}

func APIServerLBName(clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, "API", "PublicLoadBalancer")
}

func APIServerLBSKU() string {
	return "Standard"
}

func APIServerLBType() string {
	return "Public"
}

func APIServerLBFrontendIPName(clusterName string) string {
	return fmt.Sprintf("%s-%s", APIServerLBName(clusterName), "Frontend")
}

func OSDiskCachingType() string {
	return "ReadWrite"
}

func MasterSubnetName(clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, "VirtualNetwork", "MasterSubnet")
}

func ToAzureClusterPtr(v interface{}) (*capz.AzureCluster, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capz.AzureCluster{}, v)
	}

	customObjectPointer, ok := v.(*capz.AzureCluster)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capz.AzureCluster{}, v)
	}

	return customObjectPointer, nil
}

func ToAzureMachinePtr(v interface{}) (*capz.AzureMachine, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capz.AzureMachine{}, v)
	}

	customObjectPointer, ok := v.(*capz.AzureMachine)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capz.AzureMachine{}, v)
	}

	return customObjectPointer, nil
}

func ToAzureMachinePoolPtr(v interface{}) (*capzexp.AzureMachinePool, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capzexp.AzureMachinePool{}, v)
	}

	customObjectPointer, ok := v.(*capzexp.AzureMachinePool)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capzexp.AzureMachinePool{}, v)
	}

	return customObjectPointer, nil
}
