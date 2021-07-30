package validator

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/filter"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
)

type HttpHandlerFactoryConfig struct {
	CtrlReader client.Reader
	CtrlClient client.Client
	Logger     micrologger.Logger
}

// HttpHandlerFactory creates HTTP handlers for validating create and update requests.
type HttpHandlerFactory struct {
	ctrlReader client.Reader
	ctrlClient client.Client
	logger     micrologger.Logger
}

func NewHttpHandlerFactory(config HttpHandlerFactoryConfig) (*HttpHandlerFactory, error) {
	if config.CtrlReader == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlReader must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &HttpHandlerFactory{
		ctrlReader: config.CtrlReader,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return h, nil
}

// NewCreateHandler returns a HTTP handler for validating create requests.
func (h *HttpHandlerFactory) NewCreateHandler(webhookCreateHandler WebhookCreateHandler) http.HandlerFunc {
	validateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) error {
		// Decode the new CR from the request.
		object, err := webhookCreateHandler.Decode(review.Request.Object)
		if err != nil {
			return microerror.Mask(err)
		}

		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, h.ctrlClient, object)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		// Check if the CR should be validated by the azure-admission-controller.
		ok, err := filter.IsObjectReconciledByLegacyRelease(ctx, h.logger, h.ctrlReader, object, ownerClusterGetter)
		if err != nil {
			return microerror.Mask(err)
		}

		if ok {
			// Validate the CR.
			err = webhookCreateHandler.OnCreateValidate(ctx, object)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}

	return h.newHttpHandler(webhookCreateHandler, validateFunc)
}

// NewUpdateHandler returns a HTTP handler for validating update requests.
func (h *HttpHandlerFactory) NewUpdateHandler(webhookUpdateHandler WebhookUpdateHandler) http.HandlerFunc {
	validateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) error {
		// Decode the new updated CR from the request.
		object, err := webhookUpdateHandler.Decode(review.Request.Object)
		if err != nil {
			return microerror.Mask(err)
		}

		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, h.ctrlClient, object)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		// Check if the CR should be validated by the azure-admission-controller.
		ok, err := filter.IsObjectReconciledByLegacyRelease(ctx, h.logger, h.ctrlReader, object, ownerClusterGetter)
		if err != nil {
			return microerror.Mask(err)
		}

		if ok {
			// Decode the old CR from the request (before the update).
			oldObject, err := webhookUpdateHandler.Decode(review.Request.OldObject)
			if err != nil {
				return microerror.Mask(err)
			}

			// Validate the CR.
			err = webhookUpdateHandler.OnUpdateValidate(ctx, oldObject, object)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}

	return h.newHttpHandler(webhookUpdateHandler, validateFunc)
}

// newHttpHandler returns a HTTP handler for validating a request with the specified validation
// function.
// This function is basically the same as the existing Handler func, with the only difference that
// it is now wrapped into the New...Handler funcs above in order to first decode the CR and check
// if it should be validated by azure-admission-controller.
func (h *HttpHandlerFactory) newHttpHandler(webhookHandler WebhookHandlerBase, validateFunc func(ctx context.Context, review v1beta1.AdmissionReview) error) http.HandlerFunc {
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
