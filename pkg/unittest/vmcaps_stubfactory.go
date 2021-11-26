package unittest

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

type VMCapsStubFactory struct {
	logger      micrologger.Logger
	stubbedSKUs map[string]compute.ResourceSku
}

func NewVMCapsStubFactory(stubbedSKUs map[string]compute.ResourceSku, logger micrologger.Logger) vmcapabilities.Factory {
	return &VMCapsStubFactory{
		logger:      logger,
		stubbedSKUs: stubbedSKUs,
	}
}

func (s *VMCapsStubFactory) GetClient(_ context.Context, _ client.Client, _ v1.ObjectMeta) (*vmcapabilities.VMSKU, error) {
	stubAPI := NewResourceSkuStubAPI(s.stubbedSKUs)

	vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
		Azure:  stubAPI,
		Logger: s.logger,
	})
	if err != nil {
		return nil, err
	}
	return vmcaps, nil
}
