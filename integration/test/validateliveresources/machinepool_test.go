// +build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	machinepoolpkg "github.com/giantswarm/azure-admission-controller/pkg/machinepool"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func TestMachinePoolFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var machinePoolList capiexp.MachinePoolList
	err := ctrlClient.List(ctx, &machinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	for _, machinePool := range machinePoolList.Items {
		if !machinePool.GetDeletionTimestamp().IsZero() {
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

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &machinePool, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", machinePool.Namespace, machinePool.Name)
			t.Errorf("Expected MachinePool '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestMachinePoolWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var machinePoolWebhookHandler *machinepoolpkg.WebhookHandler
	{
		c := machinepoolpkg.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Decoder:    NewDecoder(),
			Logger:     logger,
			VMcaps:     NewVMCapabilities(t, logger),
		}
		machinePoolWebhookHandler, err = machinepoolpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var machinePoolList capiexp.MachinePoolList
	err = ctrlClient.List(ctx, &machinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	for _, machinePool := range machinePoolList.Items {
		if !machinePool.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		var patches []mutator.PatchOperation

		// Test mutating webhook, on create. Here we are passing the pointer to a copy of the
		// object, because the OnCreateMutate func can change it.
		patches, err = machinePoolWebhookHandler.OnCreateMutate(ctx, machinePool.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on create, " +
				"because they should already have all fields set correctly.")
		}

		// Test validating webhook, on create.
		err = machinePoolWebhookHandler.OnCreateValidate(ctx, &machinePool)
		if err != nil {
			t.Fatal(err)
		}

		updatedMachinePool := machinePool.DeepCopy()
		updatedMachinePool.Labels["test.giantswarm.io/dummy"] = "this is not really saved"

		// Test mutating webhook, on update. Here we are passing the pointer to a copy of the
		// object, because the OnUpdateMutate func can change it.
		patches, err = machinePoolWebhookHandler.OnUpdateMutate(ctx, &machinePool, updatedMachinePool.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on update, " +
				"because they should already have all fields set correctly.")
		}

		// Test validating webhook, on update.
		err = machinePoolWebhookHandler.OnUpdateValidate(ctx, &machinePool, updatedMachinePool)
		if err != nil {
			t.Fatal(err)
		}
	}
}
