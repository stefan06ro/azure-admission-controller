//go:build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/azureupdate"
	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

func TestAzureConfigFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var azureConfigList v1alpha1.AzureConfigList
	err := ctrlClient.List(ctx, &azureConfigList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureConfig := range azureConfigList.Items {
		if !azureConfig.GetDeletionTimestamp().IsZero() {
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

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &azureConfig, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", azureConfig.Namespace, azureConfig.Name)
			t.Errorf("Expected AzureConfig '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestAzureConfigWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var azureConfigWebhookHandler *azureupdate.AzureConfigWebhookHandler
	{
		c := azureupdate.AzureConfigWebhookHandlerConfig{
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			Logger:     logger,
		}
		azureConfigWebhookHandler, err = azureupdate.NewAzureConfigWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var azureConfigList v1alpha1.AzureConfigList
	err = ctrlClient.List(ctx, &azureConfigList)
	if err != nil {
		t.Fatal(err)
	}

	for _, azureConfig := range azureConfigList.Items {
		if !azureConfig.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		updatedAzureConfig := azureConfig.DeepCopy()
		updatedAzureConfig.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test validating webhook, on update.
		err = azureConfigWebhookHandler.OnUpdateValidate(ctx, &azureConfig, updatedAzureConfig)
		if err != nil {
			t.Fatal(err)
		}
	}
}
