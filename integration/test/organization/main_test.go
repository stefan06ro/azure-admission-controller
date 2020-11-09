// +build k8srequired

package organization

import (
	"context"
	"fmt"
	"os"
	"testing"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/crd"
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/micrologger"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/integration/values"
)

const (
	prodCatalogName = "control-plane-catalog"
	testCatalogName = "control-plane-test-catalog"
	// API Groups for upstream Cluster API types.
	clusterAPIGroup                    = "cluster.x-k8s.io"
	infrastructureAPIGroup             = "infrastructure.cluster.x-k8s.io"
	experimentalClusterAPIGroup        = "exp.cluster.x-k8s.io"
	experimentalInfrastructureAPIGroup = "exp.infrastructure.cluster.x-k8s.io"
)

var (
	appTest apptest.Interface
	logger  micrologger.Logger
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	ctx := context.Background()

	{
		logger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			panic(err.Error())
		}
	}

	{
		appTest, err = apptest.New(apptest.Config{
			KubeConfigPath: env.KubeConfig(),
			Logger:         logger,
		})
		if err != nil {
			panic(err.Error())
		}
	}

	{
		err = appTest.EnsureCRDs(ctx, getRequiredCRDs())
		if err != nil {
			logger.LogCtx(ctx, "level", "error", "message", "failed ensuring crds", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
		}
	}

	{
		valuesYAML, err := values.YAML(env.AzureClientID(), env.AzureClientSecret(), env.AzureTenantID(), env.AzureSubscriptionID())
		if err != nil {
			logger.LogCtx(ctx, "level", "error", "message", "install apps failed", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
		}

		apps := []apptest.App{
			{
				CatalogName:   prodCatalogName,
				Name:          "cert-manager-app",
				Namespace:     metav1.NamespaceDefault,
				Version:       "2.3.1",
				WaitForDeploy: true,
			},
			{
				CatalogName:   testCatalogName,
				Name:          "azure-admission-controller",
				Namespace:     metav1.NamespaceDefault,
				SHA:           env.CircleSHA(),
				ValuesYAML:    valuesYAML,
				WaitForDeploy: true,
			},
		}
		err = appTest.InstallApps(ctx, apps)
		if err != nil {
			logger.LogCtx(ctx, "level", "error", "message", "install apps failed", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(-1)
		}
	}

	os.Exit(m.Run())
}

func getRequiredCRDs() []*apiextensionsv1.CustomResourceDefinition {
	clusterCRD := crd.LoadV1(clusterAPIGroup, "Cluster")
	azureClusterCRD := crd.LoadV1(infrastructureAPIGroup, "AzureCluster")
	azureMachineCRD := crd.LoadV1(infrastructureAPIGroup, "AzureMachine")
	machinePoolCRD := crd.LoadV1(experimentalClusterAPIGroup, "MachinePool")
	azureMachinePoolCRD := crd.LoadV1(experimentalInfrastructureAPIGroup, "AzureMachinePool")

	return []*apiextensionsv1.CustomResourceDefinition{
		corev1alpha1.NewAzureClusterConfigCRD(),
		corev1alpha1.NewSparkCRD(),
		providerv1alpha1.NewAzureConfigCRD(),

		clusterCRD,
		azureClusterCRD,
		azureMachineCRD,
		machinePoolCRD,
		azureMachinePoolCRD,
	}
}
