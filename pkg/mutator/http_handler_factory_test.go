package mutator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/cluster"
	"github.com/giantswarm/azure-admission-controller/pkg/release"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

const (
	// legacyRelease is a Giant Swarm release which contains azure-operator and
	// it is expected that azure-operator will reconcile CRs that belong to
	// clusters with this release. You can find Release CR manifest in testdata
	// directory.
	legacyRelease = "14.1.4"

	// capiRelease is a new Giant Swarm CAPI-style release which does not
	// contain azure-operator and it is expected that azure-operator will not
	// reconcile CRs that belong to clusters with this release. You can find
	// Release CR manifest in testdata directory.
	capiRelease = "20.0.0-v1alpha3"
)

type object interface {
	runtime.Object
	metav1.ObjectMetaAccessor
}

func TestHttpHandler(t *testing.T) {
	type testCase struct {
		name          string
		object        object
		oldObject     object
		operation     admission.Operation
		expectedError *microerror.Error
	}

	testCases := []testCase{
		{
			name: "Mutate Cluster creation for legacy releases",
			object: builder.BuildCluster(
				builder.Name("ab123"),
				builder.Labels(map[string]string{
					label.ReleaseVersion: legacyRelease,
				})),
			operation: admission.Create,
		},
		{
			name: "Mutate Cluster creation for CAPI releases",
			object: builder.BuildCluster(
				builder.Name("ab123"),
				builder.Labels(map[string]string{
					label.ReleaseVersion: capiRelease,
				})),
			operation: admission.Create,
		},
		{
			name: "Mutate Cluster creation for non-existing releases",
			object: builder.BuildCluster(
				builder.Name("ab123"),
				builder.Labels(map[string]string{
					label.ReleaseVersion: "0.0.0-NonExistentRelease",
				})),
			expectedError: release.ReleaseNotFoundError,
			operation:     admission.Create,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()
			loadReleases(t, ctx, ctrlClient)

			//
			// We want to test webhook HTTP handler for mutating create/update requests. For creating
			// the handler we will use HttpHandlerFactory to create the HTTP handler, so here we
			// are basically testing the handlers created by the factory.
			//
			// Mutation logic itself here does not matter, so we are using a generic
			// WebhookHandlerMock as a WebhookCreateHandler/WebhookUpdateHandler interface implementation.
			//
			var httpHandlerFactory *HttpHandlerFactory
			{
				c := HttpHandlerFactoryConfig{
					CtrlCache:  ctrlClient, // Passing client here, for the sake of simpler test code
					CtrlClient: ctrlClient,
				}
				httpHandlerFactory, err = NewHttpHandlerFactory(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			webhookHandlerMock := WebhookHandlerMock{
				DecodeFunc: func(runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
					return tc.object, nil
				},
			}

			var httpHandler http.HandlerFunc
			switch tc.operation {
			case admission.Create:
				httpHandler = httpHandlerFactory.NewCreateHandler(&webhookHandlerMock)
			case admission.Update:
				httpHandler = httpHandlerFactory.NewUpdateHandler(&webhookHandlerMock)
			default:
				t.Fatal("Unsupported operation")
			}

			//
			// Now that we have an HTTP handler to test, we want to send a request to it.
			//
			// That request will contain admission review for creating/updating an object. This is
			// basically the request body that API server would send to the webhook.
			//
			admissionReviewJson := getAdmissionReview(t, tc.operation, tc.object, tc.oldObject)
			request := getHttpRequest(t, admissionReviewJson)

			//
			// Finally let's call the webhook HTTP handler and get the response.
			//
			// Since we this a test, we will not call the handler with the ResponseWriter, but with
			// a ResponseRecorder from httptest package, so we can easily check the response.
			//
			httpRecorder := httptest.NewRecorder()
			httpHandler.ServeHTTP(httpRecorder, request)

			var admissionReview admission.AdmissionReview
			err = json.Unmarshal(httpRecorder.Body.Bytes(), &admissionReview)
			if err != nil {
				t.Fatal(err)
			}

			//
			// Now let's check the handler response.
			//
			if !admissionReview.Response.Allowed {
				// webhook handler returned an error and it is rejecting the request

				if tc.expectedError == nil {
					// we did not expect any errors
					msg := "Request is not allowed."

					if admissionReview.Response.Result != nil {
						msg += fmt.Sprintf(" Got status code %d. Error message: %s.",
							admissionReview.Response.Result.Code,
							admissionReview.Response.Result.Message)
					}

					t.Error(msg)
				} else {
					// we expected an error, let's check if we got what we wanted
					expectedErrorMessage := microerror.Mask(tc.expectedError).Error()
					returnedErrorMessage := ""

					if admissionReview.Response.Result != nil {
						// we set error messages in the Result, let's get it from there
						returnedErrorMessage = admissionReview.Response.Result.Message
					}

					// Full error message can have more info, but it contains expected error message.
					if !strings.Contains(returnedErrorMessage, expectedErrorMessage) {
						msg := "Request is not allowed."
						msg += fmt.Sprintf(" Got status code %d. Expected error message '%s', got '%s'.",
							admissionReview.Response.Result.Code,
							expectedErrorMessage,
							returnedErrorMessage)

						t.Fatalf(msg)
					}
				}
			}
		})
	}
}

func getHttpRequest(t *testing.T, admissionReview []byte) *http.Request {
	requestBody := bytes.NewBuffer(admissionReview)
	request, err := http.NewRequest("POST", "", requestBody)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	return request
}

func getAdmissionReview(t *testing.T, operation admission.Operation, object runtime.Object, oldObject runtime.Object) []byte {
	objectJson, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}

	admissionRequest := &admission.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  object.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		},
		Operation: operation,
		Object: runtime.RawExtension{
			Raw:    objectJson,
			Object: nil,
		},
	}

	if oldObject != nil {
		oldObjectJson, err := json.Marshal(oldObject)
		if err != nil {
			t.Fatal(err)
		}

		admissionRequest.OldObject = runtime.RawExtension{
			Raw:    oldObjectJson,
			Object: nil,
		}
	}

	admissionReview := admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{},
		Request:  admissionRequest,
	}

	admissionReviewJson, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatal(err)
	}

	return admissionReviewJson
}

func loadReleases(t *testing.T, ctx context.Context, client client.Client) {
	testReleasesToLoad := []string{legacyRelease, capiRelease}

	for _, releaseVersion := range testReleasesToLoad {
		var err error
		fileName := fmt.Sprintf("release-v%s.yaml", releaseVersion)

		var bs []byte
		{
			bs, err = ioutil.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
			}
		}

		// First parse kind.
		typeMeta := &metav1.TypeMeta{}
		err = yaml.Unmarshal(bs, typeMeta)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		if typeMeta.Kind != "Release" {
			t.Fatalf("failed to create object from input file %s, the file does not contain Release CR", fileName)
		}

		var r releasev1alpha1.Release
		err = yaml.Unmarshal(bs, &r)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}

		err = client.Create(ctx, &r)
		if err != nil {
			t.Fatalf("failed to create object from input file %s: %#v", fileName, err)
		}
	}
}
