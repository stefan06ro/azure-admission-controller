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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/dyson/certman"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/azure-admission-controller/config"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachine"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/machinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/project"
	"github.com/giantswarm/azure-admission-controller/pkg/spark"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
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
			// XXX: This orphaned throw-away stop channel is very ugly, but
			// will go away once `controller-runtime` library is updated. In
			// 0.8.x it's `context.Context` instead of channel.
			err = ctrlCache.Start(make(<-chan struct{}))
			if err != nil {
				// XXX: Due to asynchronous nature, there's no reasonable way
				// to return error from here, hence panic().
				panic(err)
			}
		}()

		ok := ctrlCache.WaitForCacheSync(make(<-chan struct{}))
		if !ok {
			return microerror.Mask(errors.New("couldn't wait for cache sync"))
		}
	}

	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	universalDeserializer := codecs.UniversalDeserializer()

	var resourceSkusClient compute.ResourceSkusClient
	{
		settings, err := auth.GetSettingsFromEnvironment()
		if err != nil {
			return microerror.Mask(err)
		}
		authorizer, err := settings.GetAuthorizer()
		if err != nil {
			return microerror.Mask(err)
		}
		resourceSkusClient = compute.NewResourceSkusClient(settings.GetSubscriptionID())
		resourceSkusClient.Client.Authorizer = authorizer
	}

	var vmcaps *vmcapabilities.VMSKU
	{
		vmcaps, err = vmcapabilities.New(vmcapabilities.Config{
			Logger: newLogger,
			Azure:  vmcapabilities.NewAzureAPI(vmcapabilities.AzureConfig{ResourceSkuClient: &resourceSkusClient}),
		})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureConfigValidator *azureupdate.AzureConfigValidator
	{
		azureConfigValidatorConfig := azureupdate.AzureConfigValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		azureConfigValidator, err = azureupdate.NewAzureConfigValidator(azureConfigValidatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureClusterWebhookHandler *azurecluster.WebhookHandler
	{
		c := azurecluster.WebhookHandlerConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlCache:  ctrlCache,
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureClusterWebhookHandler, err = azurecluster.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureClusterConfigValidator *azureupdate.AzureClusterConfigValidator
	{
		azureClusterConfigValidatorConfig := azureupdate.AzureClusterConfigValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		azureClusterConfigValidator, err = azureupdate.NewAzureClusterConfigValidator(azureClusterConfigValidatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachineWebhookHandler *azuremachine.WebhookHandler
	{
		c := azuremachine.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachineWebhookHandler, err = azuremachine.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolWebhookHandler *azuremachinepool.WebhookHandler
	{
		c := azuremachinepool.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolWebhookHandler, err = azuremachinepool.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterWebhookHandler *cluster.WebhookHandler
	{
		c := cluster.WebhookHandlerConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			CtrlReader: ctrlCache,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		clusterWebhookHandler, err = cluster.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolWebhookHandler *machinepool.WebhookHandler
	{
		c := machinepool.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		machinePoolWebhookHandler, err = machinepool.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var sparkWebhookHandler *spark.WebhookHandler
	{
		c := spark.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		sparkWebhookHandler, err = spark.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var validatorHttpHandlerFactory *validator.HttpHandlerFactory
	{
		c := validator.HttpHandlerFactoryConfig{
			CtrlClient: ctrlClient,
			CtrlCache:  ctrlCache,
		}
		validatorHttpHandlerFactory, err = validator.NewHttpHandlerFactory(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var mutatorHttpHandlerFactory *mutator.HttpHandlerFactory
	{
		c := mutator.HttpHandlerFactoryConfig{
			CtrlClient: ctrlClient,
			CtrlCache:  ctrlCache,
		}
		mutatorHttpHandlerFactory, err = mutator.NewHttpHandlerFactory(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	// Mutators.
	handler.Handle("/mutate/azuremachine/create", mutatorHttpHandlerFactory.NewCreateHandler(azureMachineWebhookHandler))
	handler.Handle("/mutate/azuremachine/update", mutatorHttpHandlerFactory.NewUpdateHandler(azureMachineWebhookHandler))
	handler.Handle("/mutate/azuremachinepool/create", mutatorHttpHandlerFactory.NewCreateHandler(azureMachinePoolWebhookHandler))
	handler.Handle("/mutate/azuremachinepool/update", mutatorHttpHandlerFactory.NewUpdateHandler(azureMachinePoolWebhookHandler))
	handler.Handle("/mutate/azurecluster/create", mutatorHttpHandlerFactory.NewCreateHandler(azureClusterWebhookHandler))
	handler.Handle("/mutate/azurecluster/update", mutatorHttpHandlerFactory.NewUpdateHandler(azureClusterWebhookHandler))
	handler.Handle("/mutate/cluster/create", mutatorHttpHandlerFactory.NewCreateHandler(clusterWebhookHandler))
	handler.Handle("/mutate/cluster/update", mutatorHttpHandlerFactory.NewUpdateHandler(clusterWebhookHandler))
	handler.Handle("/mutate/machinepool/create", mutatorHttpHandlerFactory.NewCreateHandler(machinePoolWebhookHandler))
	handler.Handle("/mutate/machinepool/update", mutatorHttpHandlerFactory.NewUpdateHandler(machinePoolWebhookHandler))
	handler.Handle("/mutate/spark/create", mutatorHttpHandlerFactory.NewCreateHandler(sparkWebhookHandler))

	// Validators.
	handler.Handle("/validate/azureconfig/update", validator.Handler(azureConfigValidator))
	handler.Handle("/validate/azureclusterconfig/update", validator.Handler(azureClusterConfigValidator))
	handler.Handle("/validate/azurecluster/create", validatorHttpHandlerFactory.NewCreateHandler(azureClusterWebhookHandler))
	handler.Handle("/validate/azurecluster/update", validatorHttpHandlerFactory.NewUpdateHandler(azureClusterWebhookHandler))
	handler.Handle("/validate/azuremachine/create", validatorHttpHandlerFactory.NewCreateHandler(azureMachineWebhookHandler))
	handler.Handle("/validate/azuremachine/update", validatorHttpHandlerFactory.NewUpdateHandler(azureMachineWebhookHandler))
	handler.Handle("/validate/azuremachinepool/create", validatorHttpHandlerFactory.NewCreateHandler(azureMachinePoolWebhookHandler))
	handler.Handle("/validate/azuremachinepool/update", validatorHttpHandlerFactory.NewUpdateHandler(azureMachinePoolWebhookHandler))
	handler.Handle("/validate/cluster/create", validatorHttpHandlerFactory.NewCreateHandler(clusterWebhookHandler))
	handler.Handle("/validate/cluster/update", validatorHttpHandlerFactory.NewUpdateHandler(clusterWebhookHandler))
	handler.Handle("/validate/machinepool/create", validatorHttpHandlerFactory.NewCreateHandler(machinePoolWebhookHandler))
	handler.Handle("/validate/machinepool/update", validatorHttpHandlerFactory.NewUpdateHandler(machinePoolWebhookHandler))
	handler.HandleFunc("/healthz", healthCheck)

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
