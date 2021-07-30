package app

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachine"
	"github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/config"
	"github.com/giantswarm/azure-admission-controller/pkg/machinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/spark"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

// RegisterWebhookHandlers first creates all required webhook handlers and then it registers
// them to the specified HttpRequestHandler with appropriate paths.
//
// Examples:
//
// - A webhook handler implementation that implements validator.WebhookCreateHandler will be
// registered to handle HTTP requests at path `/validate/<resource name>/create`.
//
// - A webhook handler implementation that implements validator.WebhookUpdateHandler will be
// registered to handle HTTP requests at path `/validate/<resource name>/update`.
//
// - A webhook handler implementation that implements mutator.WebhookCreateHandler will be
// registered to handle HTTP requests at path `/mutate/<resource name>/create`.
//
// - A webhook handler implementation that implements mutator.WebhookUpdateHandler will be
// registered to handle HTTP requests at path `/mutate/<resource name>/update`.
func RegisterWebhookHandlers(httpRequestHandler HttpRequestHandler, cfg config.Config, newLogger micrologger.Logger, ctrlClient client.Client, ctrlReader client.Reader, vmcaps *vmcapabilities.VMSKU) error {
	var err error

	var validatorHttpHandlerFactory *validator.HttpHandlerFactory
	{
		c := validator.HttpHandlerFactoryConfig{
			CtrlClient: ctrlClient,
			CtrlReader: ctrlReader,
			Logger:     newLogger,
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
			CtrlReader: ctrlReader,
			Logger:     newLogger,
		}
		mutatorHttpHandlerFactory, err = mutator.NewHttpHandlerFactory(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	handlers, err := getAllHandlers(cfg, newLogger, ctrlClient, ctrlReader, vmcaps)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, h := range handlers {
		// Check if the handler is implementing validator.WebhookCreateHandler, and if it does,
		// register a handler function for validating create requests.
		if webhookHandler, ok := h.(validator.WebhookCreateHandler); ok {
			pattern := fmt.Sprintf("/validate/%s/create", webhookHandler.Resource())
			httpHandlerFunc := validatorHttpHandlerFactory.NewCreateHandler(webhookHandler)
			httpRequestHandler.Handle(pattern, httpHandlerFunc)
		}

		// Check if the handler is implementing validator.WebhookUpdateHandler, and if it does,
		// register a handler function for validating update requests.
		if webhookHandler, ok := h.(validator.WebhookUpdateHandler); ok {
			pattern := fmt.Sprintf("/validate/%s/update", webhookHandler.Resource())
			httpHandlerFunc := validatorHttpHandlerFactory.NewUpdateHandler(webhookHandler)
			httpRequestHandler.Handle(pattern, httpHandlerFunc)
		}

		// Check if the handler is implementing mutator.WebhookCreateHandler, and if it does,
		// register a handler function for validating create requests.
		if webhookHandler, ok := h.(mutator.WebhookCreateHandler); ok {
			pattern := fmt.Sprintf("/mutate/%s/create", webhookHandler.Resource())
			httpHandlerFunc := mutatorHttpHandlerFactory.NewCreateHandler(webhookHandler)
			httpRequestHandler.Handle(pattern, httpHandlerFunc)
		}

		// Check if the handler is implementing mutator.WebhookUpdateHandler, and if it does,
		// register a handler function for validating update requests.
		if webhookHandler, ok := h.(mutator.WebhookUpdateHandler); ok {
			pattern := fmt.Sprintf("/mutate/%s/update", webhookHandler.Resource())
			httpHandlerFunc := mutatorHttpHandlerFactory.NewUpdateHandler(webhookHandler)
			httpRequestHandler.Handle(pattern, httpHandlerFunc)
		}
	}

	return nil
}

func getAllHandlers(cfg config.Config, newLogger micrologger.Logger, ctrlClient client.Client, ctrlReader client.Reader, vmcaps *vmcapabilities.VMSKU) ([]ResourceHandler, error) {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	universalDeserializer := codecs.UniversalDeserializer()
	var handlers []ResourceHandler

	{
		c := azureupdate.AzureConfigWebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		azureConfigWebhookHandler, err := azureupdate.NewAzureConfigWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, azureConfigWebhookHandler)
	}

	{
		c := azurecluster.WebhookHandlerConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlReader: ctrlReader,
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
		}
		azureClusterWebhookHandler, err := azurecluster.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, azureClusterWebhookHandler)
	}

	{
		c := azureupdate.AzureClusterConfigWebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		azureClusterConfigWebhookHandler, err := azureupdate.NewAzureClusterConfigWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, azureClusterConfigWebhookHandler)
	}

	{
		c := azuremachine.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachineWebhookHandler, err := azuremachine.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, azureMachineWebhookHandler)
	}

	{
		c := azuremachinepool.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Location:   cfg.Location,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		azureMachinePoolWebhookHandler, err := azuremachinepool.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, azureMachinePoolWebhookHandler)
	}

	{
		c := cluster.WebhookHandlerConfig{
			BaseDomain: cfg.BaseDomain,
			CtrlClient: ctrlClient,
			CtrlReader: ctrlReader,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		clusterWebhookHandler, err := cluster.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, clusterWebhookHandler)
	}

	{
		c := machinepool.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
			VMcaps:     vmcaps,
		}
		machinePoolWebhookHandler, err := machinepool.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, machinePoolWebhookHandler)
	}

	{
		c := spark.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    universalDeserializer,
			Logger:     newLogger,
		}
		sparkWebhookHandler, err := spark.NewWebhookHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		handlers = append(handlers, sparkWebhookHandler)
	}

	return handlers, nil
}
