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
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	restclient "k8s.io/client-go/rest"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/config"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachine"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
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
				providerv1alpha1.AddToScheme,
				corev1alpha1.AddToScheme,
				releasev1alpha1.AddToScheme,
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

	var azureMachinePoolCreateMutator *azuremachinepool.CreateMutator
	{
		createMutatorConfig := azuremachinepool.CreateMutatorConfig{
			Logger: newLogger,
			VMcaps: vmcaps,
		}
		azureMachinePoolCreateMutator, err = azuremachinepool.NewCreateMutator(createMutatorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var azureMachinePoolCreateValidator *azuremachinepool.CreateValidator
	{
		createValidatorConfig := azuremachinepool.CreateValidatorConfig{
			Logger: newLogger,
			VMcaps: vmcaps,
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
			Logger: newLogger,
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

	// Here we register our endpoints.
	handler := http.NewServeMux()
	// Mutators.
	handler.Handle("/mutate/azuremachinepool/create", mutator.Handler(azureMachinePoolCreateMutator))

	// Validators.
	handler.Handle("/validate/azureconfig/update", validator.Handler(azureConfigValidator))
	handler.Handle("/validate/azureclusterconfig/update", validator.Handler(azureClusterConfigValidator))
	handler.Handle("/validate/azurecluster/update", validator.Handler(azureClusterUpdateValidator))
	handler.Handle("/validate/azuremachine/create", validator.Handler(azureMachineCreateValidator))
	handler.Handle("/validate/azuremachine/update", validator.Handler(azureMachineUpdateValidator))
	handler.Handle("/validate/azuremachinepool/create", validator.Handler(azureMachinePoolCreateValidator))
	handler.Handle("/validate/azuremachinepool/update", validator.Handler(azureMachinePoolUpdateValidator))
	handler.Handle("/validate/cluster/update", validator.Handler(clusterUpdateValidator))
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
