package generic

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

func EnsureComponentVersionLabel(ctx context.Context, ctrlClient client.Client, meta metav1.Object, labelName string) (*mutator.PatchOperation, error) {
	if meta.GetLabels()[labelName] == "" {
		componentVersion, err := getLabelValueFromCluster(ctx, ctrlClient, meta, labelName)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if componentVersion == "" {
			return nil, microerror.Maskf(errors.InvalidOperationError, "Cannot find label %q in Cluster CR. Can't continue.", labelName)
		}

		return mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", escapeJSONPatchString(labelName)), componentVersion), nil
	}

	return nil, nil
}
