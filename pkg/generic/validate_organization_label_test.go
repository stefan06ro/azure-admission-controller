package generic

import (
	"context"
	"strconv"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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

func Test_WhenCreatingClusterWithExistingOrganizationThenValidationSucceeds(t *testing.T) {
	var err error
	ctx := context.Background()

	scheme := runtime.NewScheme()
	err = securityv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	organization := &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "giantswarm",
		},
		Spec: securityv1alpha1.OrganizationSpec{},
	}

	ctrlClient := fake.NewFakeClientWithScheme(scheme, organization)

	obj := newObjectWithOrganization(to.StringPtr("giantswarm"))
	err = ValidateOrganizationLabelContainsExistingOrganization(ctx, ctrlClient, obj)
	if err != nil {
		t.Fatalf("it shouldn't fail when using an existing Organization")
	}
}

func Test_WhenCreatingClusterWithNonExistingOrganizationThenValidationFails(t *testing.T) {
	var err error
	ctx := context.Background()

	scheme := runtime.NewScheme()
	err = securityv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	organization := &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "giantswarm",
		},
		Spec: securityv1alpha1.OrganizationSpec{},
	}

	ctrlClient := fake.NewFakeClientWithScheme(scheme, organization)

	obj := newObjectWithOrganization(to.StringPtr("non-existing"))
	err = ValidateOrganizationLabelContainsExistingOrganization(ctx, ctrlClient, obj)
	if err == nil {
		t.Fatalf("it should fail when using a non existing Organization")
	}
}

func Test_WhenCreatingClusterWithExistingOrganizationWithNonNormalizedNameThenValidationSucceeds(t *testing.T) {
	var err error
	ctx := context.Background()

	scheme := runtime.NewScheme()
	err = securityv1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	organization := &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-awesome-organization",
		},
		Spec: securityv1alpha1.OrganizationSpec{},
	}

	ctrlClient := fake.NewFakeClientWithScheme(scheme, organization)

	obj := newObjectWithOrganization(to.StringPtr("My Awesome Organization"))
	err = ValidateOrganizationLabelContainsExistingOrganization(ctx, ctrlClient, obj)
	if err != nil {
		t.Fatalf("it didn't find the Organization with the normalized name")
	}
}
