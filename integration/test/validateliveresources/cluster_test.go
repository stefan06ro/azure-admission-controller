// +build liveinstallation,validate

package validateliveresources

import (
	"context"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	clusterpkg "github.com/giantswarm/azure-admission-controller/pkg/cluster"
)

func TestClusters(t *testing.T) {
	var err error

	ctx := context.Background()

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatal(err)
	}

	schemeBuilder := runtime.SchemeBuilder{
		capi.AddToScheme,
		capiexp.AddToScheme,
		capz.AddToScheme,
		capzexp.AddToScheme,
		releasev1alpha1.AddToScheme,
		securityv1alpha1.AddToScheme,
	}

	ctrlClient, err := NewCtrlClient(schemeBuilder)
	if err != nil {
		t.Fatal(err)
	}

	var clusterList capi.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		t.Fatal(err)
	}

	baseDomain := env.BaseDomain()
	var clusterWebhookHandler *clusterpkg.WebhookHandler
	{
		c := clusterpkg.WebhookHandlerConfig{
			BaseDomain: baseDomain,
			CtrlClient: ctrlClient,
			Logger:     logger,
		}
		clusterWebhookHandler, err = clusterpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, cluster := range clusterList.Items {
		err = clusterWebhookHandler.OnCreateValidate(ctx, &cluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster := cluster.DeepCopy()

		updatedCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster.Annotations["test.giantswarm.io/dummy"] = "this is not really saved"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster.Labels["release.giantswarm.io/version"] = "123456789.123456789.123456789"
		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if !releaseversion.IsReleaseNotFoundError(err) {
			t.Fatalf("expected releaseNotFoundError, got error: %#v", err)
		}
	}
}
