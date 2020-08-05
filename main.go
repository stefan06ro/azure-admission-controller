package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/azure-admission-controller/config"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

func main() {
	config, err := config.Parse()
	if err != nil {
		panic(microerror.JSON(err))
	}

	azureConfigValidator, err := azureupdate.NewAzureConfigValidator(config.AzureConfig)
	if err != nil {
		panic(microerror.JSON(err))
	}

	azureClusterConfigValidator, err := azureupdate.NewAzureClusterConfigValidator(config.AzureCluster)
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.Handle("/azureconfig", validator.Handler(azureConfigValidator))
	handler.Handle("/azureclusterconfig", validator.Handler(azureClusterConfigValidator))
	handler.HandleFunc("/healthz", healthCheck)

	serve(config, handler)
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
