//go:build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	azureclusterpkg "github.com/giantswarm/azure-admission-controller/pkg/azurecluster"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestAzureClusterFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var azureClusterList capz.AzureClusterList
	err := ctrlClient.List(ctx, &azureClusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureCluster := range azureClusterList.Items {
		if !azureCluster.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, ctrlClient, objectMeta)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &azureCluster, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", azureCluster.Namespace, azureCluster.Name)
			t.Errorf("Expected AzureCluster '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestAzureClusterWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var azureClusterWebhookHandler *azureclusterpkg.WebhookHandler
	{
		c := azureclusterpkg.WebhookHandlerConfig{
			BaseDomain: env.BaseDomain(),
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			Location:   env.Location(),
			CtrlReader: ctrlClient,
			Logger:     logger,
		}
		azureClusterWebhookHandler, err = azureclusterpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var azureClusterList capz.AzureClusterList
	err = ctrlClient.List(ctx, &azureClusterList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureCluster := range azureClusterList.Items {
		if !azureCluster.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		var patches []mutator.PatchOperation

		// Test mutating webhook, on create. Here we are passing the pointer to a copy of the
		// object, because the OnCreateMutate func can change it.
		patches, err = azureClusterWebhookHandler.OnCreateMutate(ctx, azureCluster.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on create, " +
				"because they should already have all fields set correctly.")
		}

		// Test validating webhook, on create.
		err = azureClusterWebhookHandler.OnCreateValidate(ctx, &azureCluster)
		if err != nil {
			t.Fatal(err)
		}

		updatedAzureCluster := azureCluster.DeepCopy()
		updatedAzureCluster.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test mutating webhook, on update. Here we are passing the pointer to a copy of the
		// object, because the OnUpdateMutate func can change it.
		patches, err = azureClusterWebhookHandler.OnUpdateMutate(ctx, &azureCluster, updatedAzureCluster.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on update, " +
				"because they should already have all fields set correctly.")
		}

		// Test validating webhook, on update.
		err = azureClusterWebhookHandler.OnUpdateValidate(ctx, &azureCluster, updatedAzureCluster)
		if err != nil {
			t.Fatal(err)
		}
	}
}
