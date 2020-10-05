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
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/azure-admission-controller/config"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

func main() {
	err := mainError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainError() error {
	config, err := config.Parse()
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

	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return microerror.Mask(err)
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
				releasev1alpha1.AddToScheme,
			},
			Logger: newLogger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}
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

	azureConfigValidatorConfig := azureupdate.AzureConfigValidatorConfig{
		K8sClient: k8sClient,
		Logger:    newLogger,
	}
	azureConfigValidator, err := azureupdate.NewAzureConfigValidator(azureConfigValidatorConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	azureClusterConfigValidatorConfig := azureupdate.AzureClusterConfigValidatorConfig{
		Logger: newLogger,
	}
	azureClusterConfigValidator, err := azureupdate.NewAzureClusterConfigValidator(azureClusterConfigValidatorConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	createValidatorConfig := azuremachinepool.CreateValidatorConfig{
		Logger: newLogger,
		VMcaps: vmcaps,
	}
	azureMachinePoolCreateValidator, err := azuremachinepool.NewCreateValidator(createValidatorConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	updateValidatorConfig := azuremachinepool.UpdateValidatorConfig{
		Logger: newLogger,
		VMcaps: vmcaps,
	}
	azureMachinePoolUpdateValidator, err := azuremachinepool.NewUpdateValidator(updateValidatorConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.Handle("/azureconfig", validator.Handler(azureConfigValidator))
	handler.Handle("/azureclusterconfig", validator.Handler(azureClusterConfigValidator))
	handler.Handle("/azuremachinepoolcreate", validator.Handler(azureMachinePoolCreateValidator))
	handler.Handle("/azuremachinepoolupdate", validator.Handler(azureMachinePoolUpdateValidator))
	handler.HandleFunc("/healthz", healthCheck)

	newLogger.LogCtx(context.Background(), "level", "debug", "message", fmt.Sprintf("Listening on port %s", config.Address))
	serve(config, handler)

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
