package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/filter"
)

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	Deserializer = codecs.UniversalDeserializer()
)

type HttpHandlerFactoryConfig struct {
	CtrlClient client.Client
}

type HttpHandlerFactory struct {
	ctrlClient client.Client
}

func NewHttpHandlerFactory(config HttpHandlerFactoryConfig) (*HttpHandlerFactory, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}

	h := &HttpHandlerFactory{
		ctrlClient: config.CtrlClient,
	}

	return h, nil
}

func (h *HttpHandlerFactory) NewCreateHttpHandler(webhookCreateHandler WebhookCreateHandler) http.HandlerFunc {
	validateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) error {
		object, err := webhookCreateHandler.Decode(review.Request.Object)
		if err != nil {
			return microerror.Mask(err)
		}

		ok, err := filter.IsCRProcessed(ctx, h.ctrlClient, object.GetObjectMeta())
		if err != nil {
			return microerror.Mask(err)
		}

		if ok {
			err = webhookCreateHandler.OnCreateValidate(ctx, object)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}

	return h.newHttpHandler(webhookCreateHandler, validateFunc)
}

func (h *HttpHandlerFactory) NewUpdateHttpHandler(webhookUpdateHandler WebhookUpdateHandler) http.HandlerFunc {
	validateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) error {
		object, err := webhookUpdateHandler.Decode(review.Request.Object)
		if err != nil {
			return microerror.Mask(err)
		}

		ok, err := filter.IsCRProcessed(ctx, h.ctrlClient, object.GetObjectMeta())
		if err != nil {
			return microerror.Mask(err)
		}

		if ok {
			oldObject, err := webhookUpdateHandler.Decode(review.Request.OldObject)
			if err != nil {
				return microerror.Mask(err)
			}
			err = webhookUpdateHandler.OnUpdateValidate(ctx, oldObject, object)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}

	return h.newHttpHandler(webhookUpdateHandler, validateFunc)
}

func (h *HttpHandlerFactory) newHttpHandler(webhookHandler WebhookHandler, validateFunc func(ctx context.Context, review v1beta1.AdmissionReview) error) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			webhookHandler.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %s", request.Header.Get("Content-Type")))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			webhookHandler.Log("level", "error", "message", "unable to read request")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			webhookHandler.Log("level", "error", "message", "unable to parse admission review request")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = validateFunc(request.Context(), review)
		if err != nil {
			writeResponse(webhookHandler, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			return
		}

		writeResponse(webhookHandler, writer, &v1beta1.AdmissionResponse{
			Allowed: true,
			UID:     review.Request.UID,
		})
	}
}

func writeResponse(webhookHandler WebhookHandler, writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	})
	if err != nil {
		webhookHandler.Log("level", "error", "message", "unable to serialize response", "stack", microerror.JSON(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := writer.Write(resp); err != nil {
		webhookHandler.Log("level", "error", "message", "unable to write response", "stack", microerror.JSON(err))
	}

	webhookHandler.Log("level", "info", "message", fmt.Sprintf("Validated request responded with result: %t", response.Allowed))
}

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Reason:  metav1.StatusReasonBadRequest,
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		},
	}
}
