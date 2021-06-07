package mutator

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

func CopyAzureOperatorVersionLabelFromAzureClusterCR(ctx context.Context, ctrlClient client.Client, meta metav1.Object) (*PatchOperation, error) {
	if meta.GetLabels()[label.AzureOperatorVersion] == "" {
		azureOperatorVersion, err := getLabelValueFromAzureCluster(ctx, ctrlClient, meta, label.AzureOperatorVersion)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if azureOperatorVersion == "" {
			return nil, microerror.Maskf(azureOperatorVersionLabelNotFoundError, "Cannot find label %#q in AzureCluster CR. Can't continue.", label.AzureOperatorVersion)
		}

		return PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(label.AzureOperatorVersion)), azureOperatorVersion), nil
	}

	return nil, nil
}

func EnsureComponentVersionLabelFromRelease(ctx context.Context, ctrlReader client.Reader, meta metav1.Object, componentName string, labelName string) (*PatchOperation, error) {
	var releaseVersion = meta.GetLabels()[label.ReleaseVersion]
	if releaseVersion == "" {
		return nil, microerror.Maskf(releaseLabelNotFoundError, "Cannot find label %#q in CR. Can't continue.", label.ReleaseVersion)
	}

	componentVersions, err := release.GetComponentVersionsFromRelease(ctx, ctrlReader, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if componentVersions[componentName] == "" {
		return nil, microerror.Maskf(componentNotFoundInReleaseError, "Component %q was not found in release %q. Can't continue.", componentName, releaseVersion)
	}

	if meta.GetLabels()[labelName] != componentVersions[componentName] {
		return PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(labelName)), componentVersions[componentName]), nil
	}

	return nil, nil
}
