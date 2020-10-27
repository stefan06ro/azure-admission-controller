package generic

import (
	"strconv"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
)

func Test_ValidateOrganizationLabelUnchanged(t *testing.T) {
	testCases := []struct {
		name         string
		old          metav1.Object
		new          metav1.Object
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: no changes",
			old:          newObjectWithOrganization(to.StringPtr("giantswarm")),
			new:          newObjectWithOrganization(to.StringPtr("giantswarm")),
			errorMatcher: nil,
		},
		{
			name:         "case 1: old CR missing organization label",
			old:          newObjectWithOrganization(nil),
			new:          newObjectWithOrganization(to.StringPtr("giantswarm")),
			errorMatcher: errors.IsNotFoundError,
		},
		{
			name:         "case 2: new CR missing organization label",
			old:          newObjectWithOrganization(to.StringPtr("giantswarm")),
			new:          newObjectWithOrganization(nil),
			errorMatcher: errors.IsNotFoundError,
		},
		{
			name:         "case 3: old and new CR have different organization label",
			old:          newObjectWithOrganization(to.StringPtr("giantswarm")),
			new:          newObjectWithOrganization(to.StringPtr("dockzero")),
			errorMatcher: errors.IsInvalidOperationError,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			err := ValidateOrganizationLabelUnchanged(tc.old, tc.new)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}
		})
	}
}
