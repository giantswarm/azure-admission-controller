package config

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"gopkg.in/alecthomas/kingpin.v2"

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

	AzureCluster azureupdate.AzureClusterConfigValidatorConfig
	AzureConfig  azureupdate.AzureConfigValidatorConfig
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

	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&result.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&result.KeyFile)
	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&result.Address)

	// add logger to each admission handler
	result.AzureCluster.Logger = newLogger
	result.AzureConfig.Logger = newLogger

	kingpin.Parse()
	return result, nil
}
