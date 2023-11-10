package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyson/certman"
	expcapz "github.com/giantswarm/apiextensions/v6/pkg/apis/capzexp/v1alpha3"
	corev1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/app"
	"github.com/giantswarm/azure-admission-controller/pkg/config"
	"github.com/giantswarm/azure-admission-controller/pkg/project"
)

func main() {
	err := mainError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainError() error {
	cfg, err := config.Parse()
	if err != nil {
		return microerror.Mask(err)
	}

	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var ctrlClient client.Client
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return microerror.Mask(err)
		}

		restConfig.UserAgent = fmt.Sprintf("%s/%s", project.Name(), project.Version())

		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				capi.AddToScheme,
				capz.AddToScheme,
				expcapz.AddToScheme,
				providerv1alpha1.AddToScheme,
				corev1alpha1.AddToScheme,
				releasev1alpha1.AddToScheme,
				capzexp.AddToScheme,
				securityv1alpha1.AddToScheme,
			},
			Logger: newLogger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}

		ctrlClient = k8sClient.CtrlClient()
	}

	var ctrlCache cache.Cache
	{
		mapper, err := apiutil.NewDynamicRESTMapper(rest.CopyConfig(k8sClient.RESTConfig()))
		if err != nil {
			return microerror.Mask(err)
		}

		o := cache.Options{
			Scheme: k8sClient.Scheme(),
			Mapper: mapper,
		}

		ctrlCache, err = cache.New(k8sClient.RESTConfig(), o)
		if err != nil {
			return microerror.Mask(err)
		}

		go func() {
			err = ctrlCache.Start(context.Background())
			if err != nil {
				// XXX: Due to asynchronous nature, there's no reasonable way
				// to return error from here, hence panic().
				panic(err)
			}
		}()

		ok := ctrlCache.WaitForCacheSync(context.Background())
		if !ok {
			return microerror.Mask(errors.New("couldn't wait for cache sync"))
		}
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.HandleFunc("/healthz", healthCheck)

	vmcapsFactory, err := vmcapabilities.NewFactory(newLogger)
	if err != nil {
		return microerror.Mask(err)
	}

	// Register all webhook handlers
	err = app.RegisterWebhookHandlers(handler, cfg, newLogger, ctrlClient, ctrlCache, vmcapsFactory)
	if err != nil {
		return microerror.Mask(err)
	}

	newLogger.LogCtx(context.Background(), "level", "debug", "message", fmt.Sprintf("Listening on port %s", cfg.Address))
	serve(cfg, handler)

	return nil
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func serve(config config.Config, handler http.Handler) {
	cm, err := certman.New(config.CertFile, config.KeyFile)
	if err != nil {
		panic(microerror.JSON(err))
	}
	if err := cm.Watch(); err != nil {
		panic(microerror.JSON(err))
	}

	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: cm.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
