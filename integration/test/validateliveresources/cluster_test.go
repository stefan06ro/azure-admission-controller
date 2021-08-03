// +build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	clusterpkg "github.com/giantswarm/azure-admission-controller/pkg/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestClusterFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var clusterList capi.ClusterList
	err := ctrlClient.List(ctx, &clusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, cluster := range clusterList.Items {
		if !cluster.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		ownerClusterGetter := func(metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			return capi.Cluster{}, false, nil
		}

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &cluster, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name)
			t.Errorf("Expected Cluster '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestClusterWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var clusterWebhookHandler *clusterpkg.WebhookHandler
	{
		c := clusterpkg.WebhookHandlerConfig{
			BaseDomain: env.BaseDomain(),
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			CtrlReader: ctrlClient,
			Logger:     logger,
		}
		clusterWebhookHandler, err = clusterpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var clusterList capi.ClusterList
	err = ctrlClient.List(ctx, &clusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, cluster := range clusterList.Items {
		if !cluster.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		var patches []mutator.PatchOperation

		// Test mutating webhook, on create. Here we are passing the pointer to a copy of the
		// object, because the OnCreateMutate func can change it.
		patches, err = clusterWebhookHandler.OnCreateMutate(ctx, cluster.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on create, " +
				"because they should already have all fields set correctly.")
		}

		// Test validating webhook, on create.
		err = clusterWebhookHandler.OnCreateValidate(ctx, &cluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedCluster := cluster.DeepCopy()
		updatedCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test validating webhook, on update.
		// Test mutating webhook, on update. Here we are passing the pointer to a copy of the
		// object, because the OnUpdateMutate func can change it.
		patches, err = clusterWebhookHandler.OnUpdateMutate(ctx, &cluster, updatedCluster.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on update, " +
				"because they should already have all fields set correctly.")
		}

		err = clusterWebhookHandler.OnUpdateValidate(ctx, &cluster, updatedCluster)
		if err != nil {
			t.Fatal(err)
		}
	}
}
