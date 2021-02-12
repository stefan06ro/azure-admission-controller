package errors

import (
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestIgnoreCAPIErrorForField(t *testing.T) {
	testCases := []struct {
		name                string
		field               string
		inputError          error
		expectedToBeError   bool
		expectedMessage     string
		expectedCausesCount int
	}{
		{
			name:  "case 0 asserts that a single error will remain present, if not ignored",
			field: "metadata.Something",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is wrong",
					),
				}),
			expectedToBeError:   true,
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [metadata.Name: Invalid value: "testing": Resource name is wrong]`,
			expectedCausesCount: 1,
		},
		{
			name:  "case 1 asserts that a single error can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is wrong",
					),
				}),
			expectedToBeError: false,
		},
		{
			name:  "case 2 asserts that multiple errors can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("SSHSomething"),
						"testing",
						"Baaaam",
					),
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is very wrong",
					),
					field.Invalid(
						field.NewPath("template").Child("SomeOtherThing"),
						"testing",
						"Boom",
					),
				}),
			expectedToBeError:   true,
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [spec.SSHSomething: Invalid value: "testing": Baaaam, template.SomeOtherThing: Invalid value: "testing": Boom]`,
			expectedCausesCount: 2,
		},
		{
			name:  "case 3 asserts that multiple errors can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("SSHSomething"),
						"testing",
						"Baaaam",
					),
					field.Invalid(
						field.NewPath("template").Child("SomeOtherThing"),
						"testing",
						"Boom",
					),
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is very wrong",
					),
				}),
			expectedToBeError:   true,
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [spec.SSHSomething: Invalid value: "testing": Baaaam, template.SomeOtherThing: Invalid value: "testing": Boom]`,
			expectedCausesCount: 2,
		},
		{
			name:  "case 4 asserts that multiple errors can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is very wrong",
					),
					field.Invalid(
						field.NewPath("spec").Child("SSHSomething"),
						"testing",
						"Baaaam",
					),
					field.Invalid(
						field.NewPath("template").Child("SomeOtherThing"),
						"testing",
						"Boom",
					),
				}),
			expectedToBeError:   true,
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [spec.SSHSomething: Invalid value: "testing": Baaaam, template.SomeOtherThing: Invalid value: "testing": Boom]`,
			expectedCausesCount: 2,
		},
		{
			name:  "case 5 asserts that multiple errors can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is very wrong",
					),
					field.Invalid(
						field.NewPath("template").Child("SomeOtherThing"),
						"testing",
						"Boom",
					),
				}),
			expectedToBeError:   true,
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [template.SomeOtherThing: Invalid value: "testing": Boom]`,
			expectedCausesCount: 1,
		},
		{
			name:  "case 6 asserts that multiple errors of the same type can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is wrong",
					),
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name is very wrong",
					),
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name has never been any more wrong",
					),
				}),
			expectedToBeError: false,
		},
		{
			name:  "case 7 asserts that multiple errors with special characters in the message can be ignored",
			field: "metadata.Name",
			inputError: apierrors.NewInvalid(
				schema.GroupKind{Group: "testing.x-k8s.io", Kind: "Test"},
				"test", field.ErrorList{
					field.Invalid(
						field.NewPath("metadata").Child("Name"),
						"testing",
						"Resource name doesn't match regex ^[a-z][a-z0-9-]{0,44}[a-z0-9]$, can contain only lowercase alphanumeric characters and '-', must start/end with an alphanumeric character",
					),
					field.Invalid(
						field.NewPath("template").Child("SomeOtherThing"),
						"testing",
						"Some other thing is completely wrong",
					),
				}),
			expectedMessage:     `Test.testing.x-k8s.io "test" is invalid: [template.SomeOtherThing: Invalid value: "testing": Some other thing is completely wrong]`,
			expectedCausesCount: 1,
			expectedToBeError:   true,
		},
		{
			name:              "case 8 asserts that nothing blows up if a regular error is used",
			field:             "metadata.Name",
			inputError:        errors.New("gotcha"),
			expectedToBeError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.inputError
			err = IgnoreCAPIErrorForField(tc.field, err)

			if status := apierrors.APIStatus(nil); errors.As(err, &status) {
				if tc.expectedMessage != status.Status().Message {
					t.Fatalf("expected message to be %s, got %s", tc.expectedMessage, status.Status().Message)
				}

				if tc.expectedCausesCount != len(status.Status().Details.Causes) {
					t.Fatalf("expected %d causes, got %d", tc.expectedCausesCount, len(status.Status().Details.Causes))
				}

				return
			}

			if (err != nil) == tc.expectedToBeError {
				return
			}

			t.Fatalf("expected error to be of type %T", apierrors.APIStatus(nil))
		})
	}
}
