package vmcapabilities

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/capzcredentials"
)

type FactoryImpl struct {
	cache  map[string]*VMSKU
	logger micrologger.Logger
}

func NewFactory(logger micrologger.Logger) (*FactoryImpl, error) {
	return &FactoryImpl{
		cache:  make(map[string]*VMSKU),
		logger: logger,
	}, nil
}

func (f *FactoryImpl) GetClient(ctx context.Context, ctrlClient client.Client, objectMeta v1.ObjectMeta) (*VMSKU, error) {
	subscriptionID, clientID, clientSecret, tenantID, err := capzcredentials.GetAzureCredentialsFromMetadata(ctx, ctrlClient, objectMeta)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if vmsku, hit := f.cache[subscriptionID]; hit {
		f.logger.Debugf(ctx, "VMSKU client found in cache for subscription %q", subscriptionID)
		return vmsku, nil
	}

	f.logger.Debugf(ctx, "Initializing VMSKU client for subscription %q", subscriptionID)

	var resourceSkusClient compute.ResourceSkusClient
	{
		settings := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)
		authorizer, err := settings.Authorizer()
		if err != nil {
			return nil, microerror.Mask(err)
		}
		resourceSkusClient = compute.NewResourceSkusClient(subscriptionID)
		resourceSkusClient.Client.Authorizer = authorizer
	}

	vmsku, err := New(Config{
		Azure:  NewAzureAPI(AzureConfig{ResourceSkuClient: &resourceSkusClient}),
		Logger: f.logger,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	f.cache[subscriptionID] = vmsku

	return vmsku, nil
}
