//go:build k8srequired

package env

import (
	"fmt"
	"os"
)

const (
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"

	EnvVarAzureClientID       = "AZURE_CLIENTID"
	EnvVarAzureClientSecret   = "AZURE_CLIENTSECRET" // #nosec
	EnvVarAzureSubscriptionID = "AZURE_SUBSCRIPTIONID"
	EnvVarAzureTenantID       = "AZURE_TENANTID"
)

var (
	circleSHA           string
	kubeconfig          string
	azureClientID       string
	azureClientSecret   string
	azureTenantID       string
	azureSubscriptionID string
)

func init() {
	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarE2EKubeconfig))
	}

	azureClientID = os.Getenv(EnvVarAzureClientID)
	if azureClientID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureClientID))
	}

	azureClientSecret = os.Getenv(EnvVarAzureClientSecret)
	if azureClientSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureClientSecret))
	}

	azureSubscriptionID = os.Getenv(EnvVarAzureSubscriptionID)
	if azureSubscriptionID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureSubscriptionID))
	}

	azureTenantID = os.Getenv(EnvVarAzureTenantID)
	if azureTenantID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAzureTenantID))
	}
}

func CircleSHA() string {
	return circleSHA
}

func KubeConfig() string {
	return kubeconfig
}

func AzureClientID() string {
	return azureClientID
}

func AzureClientSecret() string {
	return azureClientSecret
}

func AzureSubscriptionID() string {
	return azureSubscriptionID
}

func AzureTenantID() string {
	return azureTenantID
}
