package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apiextensions/v2/pkg/label"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func EnsureReleaseVersionLabel(ctx context.Context, ctrlClient client.Client, meta metav1.Object) (*mutator.PatchOperation, error) {
	if meta.GetLabels()[label.ReleaseVersion] == "" {
		// Retrieve the Cluster ID.
		clusterID := meta.GetLabels()[label.Cluster]
		if clusterID == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Object has no %s label, can't detect release version.", label.Cluster)
		}

		// Get release from Cluster.
		release, err := getLabelValueFromCluster(ctx, ctrlClient, meta, label.ReleaseVersion)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if release == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Cluster %s did not have a release label set. Can't continue.", clusterID)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(label.ReleaseVersion)), release), nil
	}

	return nil, nil
}

func getLabelValueFromCluster(ctx context.Context, ctrlClient client.Client, meta metav1.Object, labelName string) (string, error) {
	clusterID := meta.GetLabels()[label.Cluster]
	if clusterID == "" {
		return "", microerror.Maskf(errors.InvalidOperationError, "Object has no %s label, can't detect cluster ID.", label.Cluster)
	}

	// Retrieve the `Cluster` CR related to this object.
	cluster := &capiv1alpha3.Cluster{}
	{
		err := ctrlClient.Get(ctx, client.ObjectKey{Name: clusterID, Namespace: meta.GetNamespace()}, cluster)
		if apierrors.IsNotFound(err) {
			return "", microerror.Maskf(errors.InvalidOperationError, "Looking for Cluster named %s but it was not found. Can't continue.", clusterID)
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
