// +build liveinstallation,validate

package env

import (
	"fmt"
	"os"
)

const (
	// EnvVarBaseDomain is the process environment variable representing the
	// E2E_BASE_DOMAIN env var which is set to k8s cluster base domain.
	// For example, if we have a cluster named "mycluster" with API Server
	// domain is api.mycluster.k8s.example.com, then the base domain is
	// k8s.example.com.
	envVarBaseDomain = "E2E_BASE_DOMAIN"

	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
)

var (
	e2eBaseDomain string
	kubeconfig    string
)

func init() {
	e2eBaseDomain = os.Getenv(envVarBaseDomain)
	if e2eBaseDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", envVarBaseDomain))
	}

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarE2EKubeconfig))
	}
}

func BaseDomain() string {
	return e2eBaseDomain
}

func KubeConfig() string {
	return kubeconfig
}
