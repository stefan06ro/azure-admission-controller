package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
	"github.com/giantswarm/azure-admission-controller/pkg/normalize"
)

// EnsureOrganizationLabelNormalized ensures that given object has normalized
// value for organization metadata label.
func EnsureOrganizationLabelNormalized(ctx context.Context, obj metav1.Object) (*mutator.PatchOperation, error) {
	lbls := obj.GetLabels()

	org := lbls[label.Organization]
	normalized := normalize.AsDNSLabelName(org)

	if org == normalized {
		// All good. No need to do anything.
		return nil, nil
	}

	// Replace Organization label with normalized value.
	// NOTE: Label key must be encoded as per RFC6901: https://tools.ietf.org/html/rfc6901#section-3
	path := fmt.Sprintf("/metadata/labels/%s", strings.ReplaceAll(label.Organization, "/", "~1"))
	p := mutator.PatchReplace(path, normalized)
	return &p, nil
}
