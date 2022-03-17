package azuremachine

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
)

func azureMachineObject(sshKey string, failureDomain *string, labels map[string]string) *capz.AzureMachine {
	mergedLabels := map[string]string{
		"azure-operator.giantswarm.io/version": "5.0.0",
		"giantswarm.io/cluster":                "ab123",
		"cluster.x-k8s.io/cluster-name":        "ab123",
		"cluster.x-k8s.io/control-plane":       "true",
		"giantswarm.io/machine-pool":           "ab123",
		"giantswarm.io/organization":           "giantswarm",
		"release.giantswarm.io/version":        "13.0.0",
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}
	am := capz.AzureMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureMachine",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels:    mergedLabels,
		},
		Spec: capz.AzureMachineSpec{
			FailureDomain: failureDomain,
			Image: &capz.Image{
				Marketplace: &capz.AzureMarketplaceImage{
					Publisher:       "kinvolk",
					Offer:           "flatcar-container-linux-free",
					SKU:             "stable",
					Version:         "2345.3.1",
					ThirdPartyImage: false,
				},
			},
			OSDisk: capz.OSDisk{
				OSType:      "Linux",
				CachingType: "ReadWrite",
				DiskSizeGB:  50,
				ManagedDisk: capz.ManagedDisk{
					StorageAccountType: "Premium_LRS",
				},
			},
			SSHPublicKey: sshKey,
			VMSize:       "Standard_D4s_v3",
		},
	}

	return &am
}
