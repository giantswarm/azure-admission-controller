package config

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
)

const (
	defaultAddress = ":8080"
)

type Config struct {
	CertFile          string
	KeyFile           string
	Address           string
	AvailabilityZones string
	Location          string

	AzureCluster azureupdate.AzureClusterConfigValidatorConfig
	AzureConfig  azureupdate.AzureConfigValidatorConfig

	AzureMachinePoolCreate azuremachinepool.CreateValidatorConfig
	AzureMachinePoolUpdate azuremachinepool.UpdateValidatorConfig
}

func Parse() (Config, error) {
	var err error
	var result Config

	// Create a new logger that is used by all admitters.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var resourceSkusClient *compute.ResourceSkusClient
	{
		//client := compute.NewResourceSkusClient(subscriptionID)
		//client.Authorizer = authorizer
		//_ = client.AddToUserAgent(partnerID)
		//senddecorator.ConfigureClient(&backpressure.Backpressure{}, client)
	}

	var vmcaps *vmcapabilities.VMSKU
	{
		vmcaps, err = vmcapabilities.New(vmcapabilities.Config{
			Logger:             newLogger,
			ResourceSkusClient: resourceSkusClient,
		})
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&result.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&result.KeyFile)
	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&result.Address)

	// add logger to each admission handler
	result.AzureCluster.Logger = newLogger
	result.AzureConfig.Logger = newLogger
	result.AzureMachinePoolCreate.Logger = newLogger
	result.AzureMachinePoolUpdate.Logger = newLogger

	// Add the VM capabilities helper to the handlers that need it.
	result.AzureMachinePoolCreate.VMcaps = vmcaps

	kingpin.Parse()
	return result, nil
}
