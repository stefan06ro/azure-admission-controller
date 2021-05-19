// +build k8srequired

package createcluster

import (
	"context"
	"testing"
	"time"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/crd"
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/integration/util"
	"github.com/giantswarm/azure-admission-controller/integration/values"
)

const (
	crsFolder       = "testdata"
	prodCatalogName = "control-plane-catalog"
	testCatalogName = "control-plane-test-catalog"
	// API Groups for upstream Cluster API types.
	giantswarmCoreAPIGroup             = "core.giantswarm.io"
	clusterAPIGroup                    = "cluster.x-k8s.io"
	infrastructureAPIGroup             = "infrastructure.cluster.x-k8s.io"
	experimentalClusterAPIGroup        = "exp.cluster.x-k8s.io"
	experimentalInfrastructureAPIGroup = "exp.infrastructure.cluster.x-k8s.io"
	releaseAPIGroup                    = "release.giantswarm.io"
	securityAPIGroup                   = "security.giantswarm.io"
)

func TestCreateCluster(t *testing.T) {
	var err error

	ctx := context.Background()

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatal(err)
	}

	var appTest apptest.Interface
	{
		runtimeScheme := runtime.NewScheme()
		appSchemeBuilder := runtime.SchemeBuilder{
			applicationv1alpha1.AddToScheme,
			apiextensionsv1.AddToScheme,
			capi.AddToScheme,
			capz.AddToScheme,
			capiexp.AddToScheme,
			capzexp.AddToScheme,
			securityv1alpha1.AddToScheme,
			corev1.AddToScheme,
			corev1alpha1.AddToScheme,
			releasev1alpha1.AddToScheme,
		}
		err = appSchemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			t.Fatal(err)
		}
		appTest, err = apptest.New(apptest.Config{
			KubeConfigPath: env.KubeConfig(),
			Logger:         logger,
			Scheme:         runtimeScheme,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err = appTest.EnsureCRDs(ctx, getRequiredCRDs())
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		valuesYAML, err := values.YAML(env.AzureClientID(), env.AzureClientSecret(), env.AzureSubscriptionID(), env.AzureTenantID())
		if err != nil {
			t.Fatal(err)
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
			t.Fatal(err)
		}
	}

	o := func() error {
		err = util.CreateCRsInFolder(ctx, appTest.CtrlClient(), crsFolder)
		if err != nil {
			deleteErr := util.DeleteCRsInFolder(ctx, appTest.CtrlClient(), crsFolder)
			t.Log(microerror.JSON(deleteErr))
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewConstant(backoff.ShortMaxWait, 10*time.Second)
	n := backoff.NewNotifier(logger, ctx)

	err = backoff.RetryNotify(o, b, n)
	_ = util.DeleteCRsInFolder(ctx, appTest.CtrlClient(), crsFolder)
	if err != nil {
		t.Log(microerror.JSON(err))
		t.Fatal(err)
	}
}

func getRequiredCRDs() []*apiextensionsv1.CustomResourceDefinition {
	return []*apiextensionsv1.CustomResourceDefinition{
		corev1alpha1.NewAzureClusterConfigCRD(),
		providerv1alpha1.NewAzureConfigCRD(),
		crd.LoadV1(infrastructureAPIGroup, "AzureCluster"),
		crd.LoadV1(infrastructureAPIGroup, "AzureMachine"),
		crd.LoadV1(experimentalInfrastructureAPIGroup, "AzureMachinePool"),
		crd.LoadV1(clusterAPIGroup, "Cluster"),
		crd.LoadV1(experimentalClusterAPIGroup, "MachinePool"),
		crd.LoadV1(securityAPIGroup, "Organization"),
		crd.LoadV1(releaseAPIGroup, "Release"),
		crd.LoadV1(giantswarmCoreAPIGroup, "Spark"),
	}
}
