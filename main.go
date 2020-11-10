package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
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

	var azureClusterCreateMutator *azurecluster.CreateMutator
	{
		conf := azurecluster.CreateMutatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureClusterCreateMutator, err = azurecluster.NewCreateMutator(conf)
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

	var azureMachineCreateMutator *azuremachine.CreateMutator
	{
		createMutatorConfig := azuremachine.CreateMutatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureMachineCreateMutator, err = azuremachine.NewCreateMutator(createMutatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolCreateMutator *azuremachinepool.CreateMutator
	{
		createMutatorConfig := azuremachinepool.CreateMutatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolCreateMutator, err = azuremachinepool.NewCreateMutator(createMutatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolCreateValidator *azuremachinepool.CreateValidator
	{
		createValidatorConfig := azuremachinepool.CreateValidatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolCreateValidator, err = azuremachinepool.NewCreateValidator(createValidatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolUpdateValidator *azuremachinepool.UpdateValidator
	{
		updateValidatorConfig := azuremachinepool.UpdateValidatorConfig{
			Logger: newLogger,
			VMcaps: vmcaps,
		}
		azureMachinePoolUpdateValidator, err = azuremachinepool.NewUpdateValidator(updateValidatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureClusterCreateValidator *azurecluster.CreateValidator
	{
		c := azurecluster.CreateValidatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureClusterCreateValidator, err = azurecluster.NewCreateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureClusterUpdateValidator *azurecluster.UpdateValidator
	{
		c := azurecluster.UpdateValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		azureClusterUpdateValidator, err = azurecluster.NewUpdateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachineCreateValidator *azuremachine.CreateValidator
	{
		c := azuremachine.CreateValidatorConfig{
			CtrlClient: ctrlClient,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachineCreateValidator, err = azuremachine.NewCreateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachineUpdateValidator *azuremachine.UpdateValidator
	{
		c := azuremachine.UpdateValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		azureMachineUpdateValidator, err = azuremachine.NewUpdateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterCreateMutator *cluster.CreateMutator
	{
		conf := cluster.CreateMutatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		clusterCreateMutator, err = cluster.NewCreateMutator(conf)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterCreateValidator *cluster.CreateValidator
	{
		c := cluster.CreateValidatorConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		clusterCreateValidator, err = cluster.NewCreateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusterUpdateValidator *cluster.UpdateValidator
	{
		c := cluster.UpdateValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		clusterUpdateValidator, err = cluster.NewUpdateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolCreateMutator *machinepool.CreateMutator
	{
		c := machinepool.CreateMutatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		machinePoolCreateMutator, err = machinepool.NewCreateMutator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolUpdateMutator *machinepool.UpdateMutator
	{
		c := machinepool.UpdateMutatorConfig{
			Logger: newLogger,
		}
		machinePoolUpdateMutator, err = machinepool.NewUpdateMutator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolCreateValidator *machinepool.CreateValidator
	{
		c := machinepool.CreateValidatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		machinePoolCreateValidator, err = machinepool.NewCreateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var machinePoolUpdateValidator *machinepool.UpdateValidator
	{
		c := machinepool.UpdateValidatorConfig{
			Logger: newLogger,
		}
		machinePoolUpdateValidator, err = machinepool.NewUpdateValidator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var sparkCreateMutator *spark.CreateMutator
	{
		c := spark.CreateMutatorConfig{
			CtrlClient: ctrlClient,
			Logger:     newLogger,
		}
		sparkCreateMutator, err = spark.NewCreateMutator(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	// Mutators.
	handler.Handle("/mutate/azuremachine/create", mutator.Handler(azureMachineCreateMutator))
	handler.Handle("/mutate/azuremachinepool/create", mutator.Handler(azureMachinePoolCreateMutator))
	handler.Handle("/mutate/azurecluster/create", mutator.Handler(azureClusterCreateMutator))
	handler.Handle("/mutate/cluster/create", mutator.Handler(clusterCreateMutator))
	handler.Handle("/mutate/machinepool/create", mutator.Handler(machinePoolCreateMutator))
	handler.Handle("/mutate/machinepool/update", mutator.Handler(machinePoolUpdateMutator))
	handler.Handle("/mutate/spark/create", mutator.Handler(sparkCreateMutator))

	// Validators.
	handler.Handle("/validate/azureconfig/update", validator.Handler(azureConfigValidator))
	handler.Handle("/validate/azureclusterconfig/update", validator.Handler(azureClusterConfigValidator))
	handler.Handle("/validate/azurecluster/create", validator.Handler(azureClusterCreateValidator))
	handler.Handle("/validate/azurecluster/update", validator.Handler(azureClusterUpdateValidator))
	handler.Handle("/validate/azuremachine/create", validator.Handler(azureMachineCreateValidator))
	handler.Handle("/validate/azuremachine/update", validator.Handler(azureMachineUpdateValidator))
	handler.Handle("/validate/azuremachinepool/create", validator.Handler(azureMachinePoolCreateValidator))
	handler.Handle("/validate/azuremachinepool/update", validator.Handler(azureMachinePoolUpdateValidator))
	handler.Handle("/validate/cluster/create", validator.Handler(clusterCreateValidator))
	handler.Handle("/validate/cluster/update", validator.Handler(clusterUpdateValidator))
	handler.Handle("/validate/machinepool/create", validator.Handler(machinePoolCreateValidator))
	handler.Handle("/validate/machinepool/update", validator.Handler(machinePoolUpdateValidator))
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
	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
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

	err := server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
