// +build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	azuremachinepoolpkg "github.com/giantswarm/azure-admission-controller/pkg/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

func TestAzureMachinePoolFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var azureMachinePoolList capzexp.AzureMachinePoolList
	err := ctrlClient.List(ctx, &azureMachinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureMachinePool := range azureMachinePoolList.Items {
		if !azureMachinePool.GetDeletionTimestamp().IsZero() {
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

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &azureMachinePool, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", azureMachinePool.Namespace, azureMachinePool.Name)
			t.Errorf("Expected AzureMachinePool '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestAzureMachinePoolWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var azureMachinePoolWebhookHandler *azuremachinepoolpkg.WebhookHandler
	{
		c := azuremachinepoolpkg.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    NewDecoder(),
			Location:   env.Location(),
			Logger:     logger,
			VMcaps:     NewVMCapabilities(t, logger),
		}
		azureMachinePoolWebhookHandler, err = azuremachinepoolpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var azureMachinePoolList capzexp.AzureMachinePoolList
	err = ctrlClient.List(ctx, &azureMachinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureMachinePool := range azureMachinePoolList.Items {
		if !azureMachinePool.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		// Test mutating webhook, on create. Here we are passing the pointer to a copy of the
		// object, because the OnCreateMutate func can change it.
		_, err = azureMachinePoolWebhookHandler.OnCreateMutate(ctx, azureMachinePool.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on create.
		err = azureMachinePoolWebhookHandler.OnCreateValidate(ctx, &azureMachinePool)
		if err != nil {
			t.Fatal(err)
		}

		updatedAzureMachinePool := azureMachinePool.DeepCopy()
		updatedAzureMachinePool.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test mutating webhook, on update. Here we are passing the pointer to a copy of the
		// object, because the OnUpdateMutate func can change it.
		_, err = azureMachinePoolWebhookHandler.OnUpdateMutate(ctx, &azureMachinePool, updatedAzureMachinePool.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}

		// Test validating webhook, on update.
		err = azureMachinePoolWebhookHandler.OnUpdateValidate(ctx, &azureMachinePool, updatedAzureMachinePool)
		if err != nil {
			t.Fatal(err)
		}
	}
}
