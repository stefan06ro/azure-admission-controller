package filter

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

func IsCRProcessed(ctx context.Context, ctrlClient client.Client, objectMeta metav1.Object) (bool, error) {
	// Try to get release label from the CR
	releaseVersionLabel := objectMeta.GetLabels()[label.ReleaseVersion]
	if releaseVersionLabel == "" {
		// release label is not found on the CR, let's try to get it from owner
		// Cluster CR

		// first let's try CAPI cluster name label
		clusterName := objectMeta.GetLabels()[capi.ClusterLabelName]
		if clusterName == "" {
			// CAPI cluster name label not found, now let's try GS cluster ID label
			clusterID := objectMeta.GetLabels()[label.Cluster]
			if clusterID == "" {
				// We can't find out which cluster and release this CR belongs to
				return false, nil
			}

			// we found GS cluster ID, this is our cluster name
			clusterName = clusterID
		}

		// now get the owner cluster by name, we will try to check if it has
		// release label
		cluster, err := capiutil.GetClusterByName(ctx, ctrlClient, objectMeta.GetNamespace(), clusterName)
		if err != nil {
			return false, microerror.Mask(err)
		}

		releaseVersionLabel = cluster.Labels[label.ReleaseVersion]
		if releaseVersionLabel == "" {
			// We found the cluster, but we cannot find out which release the
			// cluster belongs to.
			return false, nil
		}
	}

	// Now when we have release version for the CR, let's check if the release
	// contains azure-operator.
	return release.ContainsAzureOperator(ctx, ctrlClient, releaseVersionLabel)
}
