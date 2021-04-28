package mutator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type WebhookHandler interface {
	generic.Decoder
	generic.Logger
	Resource() string
}

type WebhookCreateHandler interface {
	WebhookHandler
	OnCreateMutate(ctx context.Context, object interface{}) ([]PatchOperation, error)
}

type WebhookUpdateHandler interface {
	WebhookHandler
	OnUpdateMutate(ctx context.Context, oldObject interface{}, object interface{}) ([]PatchOperation, error)
}

var (
	scheme        = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(scheme)
	Deserializer  = codecs.UniversalDeserializer()
	InternalError = errors.New("internal admission controller error")
)

type HandlerFactoryConfig struct {
	CtrlClient client.Client
}

type HandlerFactory struct {
	ctrlClient client.Client
}

func NewHandlerFactory(config HandlerFactoryConfig) (*HandlerFactory, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}

	h := &HandlerFactory{
		ctrlClient: config.CtrlClient,
	}

	return h, nil
}

func (h *HandlerFactory) NewCreateHandler(mutator WebhookCreateHandler) http.HandlerFunc {
	mutateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error) {
		object, err := mutator.Decode(review.Request.Object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ok, err := filter.IsCRProcessed(ctx, h.ctrlClient, object.GetObjectMeta())
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var patch []PatchOperation

		if ok {
			patch, err = mutator.OnCreateMutate(ctx, object)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		return patch, nil
	}

	return h.newHandler(mutator, mutateFunc)
}

func (h *HandlerFactory) NewUpdateHandler(mutator WebhookUpdateHandler) http.HandlerFunc {
	mutateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error) {
		oldObject, err := mutator.Decode(review.Request.OldObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		object, err := mutator.Decode(review.Request.Object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ok, err := filter.IsCRProcessed(ctx, h.ctrlClient, object.GetObjectMeta())
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var patch []PatchOperation

		if ok {
			patch, err = mutator.OnUpdateMutate(ctx, oldObject, object)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		return patch, nil
	}

	return h.newHandler(mutator, mutateFunc)
}

func (h *HandlerFactory) newHandler(mutator WebhookHandler, mutateFunc func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			mutator.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %q", request.Header.Get("Content-Type")))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			mutator.Log("level", "error", "message", "unable to read request", "stack", microerror.JSON(err))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			mutator.Log("level", "error", "message", "unable to parse admission review request", "stack", microerror.JSON(err))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var patch []PatchOperation
		if review.Request.DryRun != nil && *review.Request.DryRun {
			mutator.Log("level", "debug", "message", "Dry run is not supported. Request processing stopped.", "stack", microerror.JSON(err))
		} else {
			patch, err = mutateFunc(request.Context(), review)
			if err != nil {
				writeResponse(mutator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
				return
			}
		}

		resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, extractName(review.Request))
		patchData, err := json.Marshal(patch)
		if err != nil {
			mutator.Log("level", "error", "message", fmt.Sprintf("unable to serialize patch for %s", resourceName), "stack", microerror.JSON(err))
			writeResponse(mutator, writer, errorResponse(review.Request.UID, InternalError))
			return
		}

		mutator.Log("level", "debug", "message", fmt.Sprintf("admitted %s (with %d patches)", resourceName, len(patch)))

		pt := v1beta1.PatchTypeJSONPatch
		writeResponse(mutator, writer, &v1beta1.AdmissionResponse{
			Allowed:   true,
			UID:       review.Request.UID,
			Patch:     patchData,
			PatchType: &pt,
		})
	}
}

func extractName(request *v1beta1.AdmissionRequest) string {
	if request.Name != "" {
		return request.Name
	}

	obj := metav1beta1.PartialObjectMetadata{}
	if _, _, err := Deserializer.Decode(request.Object.Raw, nil, &obj); err != nil {
		return "<unknown>"
	}

	if obj.Name != "" {
		return obj.Name
	}
	if obj.GenerateName != "" {
		return obj.GenerateName + "<generated>"
	}
	return "<unknown>"
}

func writeResponse(mutator WebhookHandler, writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	})
	if err != nil {
		mutator.Log("level", "error", "message", "unable to serialize response", "stack", microerror.JSON(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}
	if _, err := writer.Write(resp); err != nil {
		mutator.Log("level", "error", "message", "unable to write response", "stack", microerror.JSON(err))
	}
}

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
