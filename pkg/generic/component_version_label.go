package generic

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

func CopyComponentVersionLabelFromAzureClusterCR(ctx context.Context, ctrlClient client.Client, meta metav1.Object, labelName string) (*mutator.PatchOperation, error) {
	if meta.GetLabels()[labelName] == "" {
		componentVersion, err := getLabelValueFromAzureCluster(ctx, ctrlClient, meta, labelName)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if componentVersion == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Cannot find label %q in AzureCluster CR. Can't continue.", labelName)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(labelName)), componentVersion), nil
	}

	return nil, nil
}

func EnsureComponentVersionLabelFromRelease(ctx context.Context, ctrlClient client.Client, meta metav1.Object, componentName string, labelName string) (*mutator.PatchOperation, error) {
	if meta.GetLabels()[labelName] == "" {
		var releaseVersion = meta.GetLabels()[label.ReleaseVersion]
		if releaseVersion == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Release label not found. Can't continue.")
		}

		componentVersions, err := release.GetComponentVersionsFromRelease(ctx, ctrlClient, releaseVersion)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if componentVersions[componentName] == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Component %q was not found in release %q. Can't continue.", componentName, releaseVersion)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(labelName)), componentVersions[componentName]), nil
	}

	return nil, nil
}
