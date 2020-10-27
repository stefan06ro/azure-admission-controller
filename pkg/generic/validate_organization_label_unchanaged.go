package generic

import (
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func ValidateOrganizationLabelUnchanged(old, new metav1.Object) error {
	if _, exists := old.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(errors.NotFoundError, "old CR doesn't contain Organization label")
	}

	if _, exists := new.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(errors.NotFoundError, "new CR doesn't contain Organization label")
	}

	if old.GetLabels()[label.Organization] != new.GetLabels()[label.Organization] {
		return microerror.Maskf(errors.InvalidOperationError, "Organization label must not be changed")
	}

	return nil
}
