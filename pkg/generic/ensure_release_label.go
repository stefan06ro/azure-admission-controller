package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func EnsureReleaseVersionLabel(ctx context.Context, ctrlClient client.Client, meta metav1.Object) (*mutator.PatchOperation, error) {
	if meta.GetLabels()[label.ReleaseVersion] == "" {
		release, err := getReleaseLabelValueFromAzureCluster(ctx, ctrlClient, meta)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if release == "" {
			return nil, microerror.Maskf(releaseLabelNotFoundError, "AzureCluster did not have the %#q label set. Can't continue.", label.ReleaseVersion)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(label.ReleaseVersion)), release), nil
	}

	return nil, nil
}

func getReleaseLabelValueFromAzureCluster(ctx context.Context, ctrlClient client.Client, meta metav1.Object) (string, error) {
	return getLabelValueFromAzureCluster(ctx, ctrlClient, meta, label.ReleaseVersion)
}

func getLabelValueFromAzureCluster(ctx context.Context, ctrlClient client.Client, meta metav1.Object, labelName string) (string, error) {
	clusterID := meta.GetLabels()[label.Cluster]
	if clusterID == "" {
		return "", microerror.Maskf(clusterLabelNotFoundError, "Object has no %#q label, can't detect cluster ID.", label.Cluster)
	}

	// Retrieve the `AzureCluster` CR related to this object.
	cluster := &v1alpha3.AzureCluster{}
	{
		err := ctrlClient.Get(ctx, client.ObjectKey{Name: clusterID, Namespace: meta.GetNamespace()}, cluster)
		if apierrors.IsNotFound(err) {
			return "", microerror.Maskf(errors.NotFoundError, "Looking for AzureCluster named %#q but it was not found. Can't continue.", clusterID)
		} else if err != nil {
			return "", microerror.Mask(err)
		}
	}

	// Extract desired label from Cluster.
	release := cluster.GetLabels()[labelName]

	return release, nil
}

// Ensure the needed escapes are in place. See https://tools.ietf.org/html/rfc6901#section-3 .
func escapeJSONPatchString(input string) string {
	input = strings.ReplaceAll(input, "~", "~0")
	input = strings.ReplaceAll(input, "/", "~1")

	return input
}
