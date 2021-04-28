package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/dyson/certman"
	corev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	restclient "k8s.io/client-go/rest"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return microerror.Mask(err)
		}

		restConfig.UserAgent = fmt.Sprintf("%s/%s", project.Name(), project.Version())

		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				capiv1alpha3.AddToScheme,
				capzv1alpha3.AddToScheme,
				providerv1alpha1.AddToScheme,
				corev1alpha1.AddToScheme,
				releasev1alpha1.AddToScheme,
				expcapzv1alpha3.AddToScheme,
				securityv1alpha1.AddToScheme,
			},
			Logger: newLogger,

			RestConfig: restConfig,
		}

		k8sClient, err := k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}

		ctrlClient = k8sClient.CtrlClient()
	}

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

	var azureMachineMutator *azuremachine.Mutator
	{
		azureMachineMutatorConfig := azuremachine.MutatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureMachineMutator, err = azuremachine.NewMutator(azureMachineMutatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolMutator *azuremachinepool.Mutator
	{
		azureMachinePoolMutatorConfig := azuremachinepool.MutatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolMutator, err = azuremachinepool.NewMutator(azureMachinePoolMutatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolValidator *azuremachinepool.Validator
	{
		azureMachinePoolValidatorConfig := azuremachinepool.ValidatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolValidator, err = azuremachinepool.NewValidator(azureMachinePoolValidatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureClusterWebhookHandler *azurecluster.WebhookHandler
	{
		c := azurecluster.WebhookHandlerConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureClusterWebhookHandler, err = azurecluster.NewWebhookHandler(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachineValidator *azuremachine.Validator
	{
		c := azuremachine.ValidatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachineValidator, err = azuremachine.NewValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterMutator *cluster.Mutator
	{
		conf := cluster.MutatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		clusterMutator, err = cluster.NewMutator(conf)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterValidator *cluster.Validator
	{
		c := cluster.ValidatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		clusterValidator, err = cluster.NewValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolMutator *machinepool.Mutator
	{
		c := machinepool.MutatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		machinePoolMutator, err = machinepool.NewMutator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolValidator *machinepool.Validator
	{
		c := machinepool.ValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		machinePoolValidator, err = machinepool.NewValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var sparkCreateMutator *spark.Mutator
	{
		c := spark.MutatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		sparkCreateMutator, err = spark.NewMutator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var validatorHandlerFactory *validator.HandlerFactory
	{
		c := validator.HandlerFactoryConfig{
			CtrlClient: ctrlClient,
		}
		validatorHandlerFactory, err = validator.NewHandlerFactory(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var mutatorHandlerFactory *mutator.HandlerFactory
	{
		c := mutator.HandlerFactoryConfig{
			CtrlClient: ctrlClient,
		}
		mutatorHandlerFactory, err = mutator.NewHandlerFactory(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	// Mutators.
	handler.Handle("/mutate/azuremachine/create", mutatorHandlerFactory.NewCreateHandler(azureMachineMutator))
	handler.Handle("/mutate/azuremachine/update", mutatorHandlerFactory.NewUpdateHandler(azureMachineMutator))
	handler.Handle("/mutate/azuremachinepool/create", mutatorHandlerFactory.NewCreateHandler(azureMachinePoolMutator))
	handler.Handle("/mutate/azuremachinepool/update", mutatorHandlerFactory.NewUpdateHandler(azureMachinePoolMutator))
	handler.Handle("/mutate/azurecluster/create", mutatorHandlerFactory.NewCreateHandler(azureClusterWebhookHandler))
	handler.Handle("/mutate/azurecluster/update", mutatorHandlerFactory.NewUpdateHandler(azureClusterWebhookHandler))
	handler.Handle("/mutate/cluster/create", mutatorHandlerFactory.NewCreateHandler(clusterMutator))
	handler.Handle("/mutate/cluster/update", mutatorHandlerFactory.NewUpdateHandler(clusterMutator))
	handler.Handle("/mutate/machinepool/create", mutatorHandlerFactory.NewCreateHandler(machinePoolMutator))
	handler.Handle("/mutate/machinepool/update", mutatorHandlerFactory.NewUpdateHandler(machinePoolMutator))
	handler.Handle("/mutate/spark/create", mutatorHandlerFactory.NewCreateHandler(sparkCreateMutator))

	// Validators.
	handler.Handle("/validate/azureconfig/update", validatorHandlerFactory.NewUpdateHandler(azureConfigValidator))
	handler.Handle("/validate/azureclusterconfig/update", validatorHandlerFactory.NewUpdateHandler(azureClusterConfigValidator))
	handler.Handle("/validate/azurecluster/create", validatorHandlerFactory.NewCreateHandler(azureClusterWebhookHandler))
	handler.Handle("/validate/azurecluster/update", validatorHandlerFactory.NewUpdateHandler(azureClusterWebhookHandler))
	handler.Handle("/validate/azuremachine/create", validatorHandlerFactory.NewCreateHandler(azureMachineValidator))
	handler.Handle("/validate/azuremachine/update", validatorHandlerFactory.NewUpdateHandler(azureMachineValidator))
	handler.Handle("/validate/azuremachinepool/create", validatorHandlerFactory.NewCreateHandler(azureMachinePoolValidator))
	handler.Handle("/validate/azuremachinepool/update", validatorHandlerFactory.NewUpdateHandler(azureMachinePoolValidator))
	handler.Handle("/validate/cluster/create", validatorHandlerFactory.NewCreateHandler(clusterValidator))
	handler.Handle("/validate/cluster/update", validatorHandlerFactory.NewUpdateHandler(clusterValidator))
	handler.Handle("/validate/machinepool/create", validatorHandlerFactory.NewCreateHandler(machinePoolValidator))
	handler.Handle("/validate/machinepool/update", validatorHandlerFactory.NewUpdateHandler(machinePoolValidator))
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
