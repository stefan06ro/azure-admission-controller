//go:build liveinstallation

package validateliveresources

import (
	"context"
	"fmt"
	"testing"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	sparkpkg "github.com/giantswarm/azure-admission-controller/pkg/spark"
)

func TestSparkFiltering(t *testing.T) {
	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)

	var sparkList corev1alpha1.SparkList
	err := ctrlClient.List(ctx, &sparkList)
	if err != nil {
		t.Fatal(err)
	}

	for _, spark := range sparkList.Items {
		if !spark.GetDeletionTimestamp().IsZero() {
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

		result, err := filter.IsObjectReconciledByLegacyRelease(ctx, logger, ctrlClient, &spark, ownerClusterGetter)
		if err != nil {
			t.Fatal(err)
		}

		if result == false {
			objectName := fmt.Sprintf("%s/%s", spark.Namespace, spark.Name)
			t.Errorf("Expected Spark '%s' to be reconciled by a legacy release, but it's not.", objectName)
		}
	}
}

func TestSparkWebhookHandler(t *testing.T) {
	var err error

	ctx := context.Background()
	logger, _ := micrologger.New(micrologger.Config{})
	ctrlClient := NewReadOnlyCtrlClient(t)
	SetAzureEnvironmentVariables(t, ctx, ctrlClient)

	var sparkWebhookHandler *sparkpkg.WebhookHandler
	{
		c := sparkpkg.WebhookHandlerConfig{
			Decoder:    NewDecoder(),
			CtrlClient: ctrlClient,
			Logger:     logger,
		}
		sparkWebhookHandler, err = sparkpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var sparkList corev1alpha1.SparkList
	err = ctrlClient.List(ctx, &sparkList)
	if err != nil {
		t.Fatal(err)
	}

	for _, spark := range sparkList.Items {
		if !spark.GetDeletionTimestamp().IsZero() {
			// Skip CRs that are being deleted.
			continue
		}

		var patches []mutator.PatchOperation

		// Test mutating webhook, on create. Here we are passing the pointer to a copy of the
		// object, because the OnCreateMutate func can change it.
		patches, err = sparkWebhookHandler.OnCreateMutate(ctx, spark.DeepCopy())
		if err != nil {
			t.Fatal(err)
		}
		if len(patches) > 0 {
			t.Fatalf("CRs from a real management cluster must not require patches on create, " +
				"because they should already have all fields set correctly.")
		}
	}
}
