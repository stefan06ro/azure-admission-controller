package mutator

import (
	"context"
	"encoding/json"
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

// HttpHandlerFactory creates HTTP handlers for mutating create and update requests.
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

// NewCreateHandler returns a HTTP handler for mutating create requests.
func (h *HttpHandlerFactory) NewCreateHandler(mutator WebhookCreateHandler) http.HandlerFunc {
	mutateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error) {
		// Decode the new CR from the request.
		object, err := mutator.Decode(review.Request.Object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, h.ctrlClient, object)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		// Check if the CR should be mutated by the azure-admission-controller.
		ok, err := filter.IsObjectReconciledByLegacyRelease(ctx, h.logger, h.ctrlReader, object, ownerClusterGetter)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var patch []PatchOperation

		if ok {
			// Mutate the CR and get patch for those mutations.
			patch, err = mutator.OnCreateMutate(ctx, object)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		return patch, nil
	}

	return h.newHttpHandler(mutator, mutateFunc)
}

// NewUpdateHandler returns a HTTP handler for mutating update requests.
func (h *HttpHandlerFactory) NewUpdateHandler(mutator WebhookUpdateHandler) http.HandlerFunc {
	mutateFunc := func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error) {
		// Decode the new updated CR from the request.
		object, err := mutator.Decode(review.Request.Object)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ownerClusterGetter := func(objectMeta metav1.ObjectMetaAccessor) (capi.Cluster, bool, error) {
			ownerCluster, ok, err := generic.TryGetOwnerCluster(ctx, h.ctrlClient, object)
			if err != nil {
				return capi.Cluster{}, false, microerror.Mask(err)
			}

			return ownerCluster, ok, nil
		}

		// Check if the CR should be mutated by the azure-admission-controller.
		ok, err := filter.IsObjectReconciledByLegacyRelease(ctx, h.logger, h.ctrlReader, object, ownerClusterGetter)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var patch []PatchOperation

		if ok {
			// Decode the old CR from the request (before the update).
			oldObject, err := mutator.Decode(review.Request.OldObject)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			// Mutate the CR and get patch for those mutations.
			patch, err = mutator.OnUpdateMutate(ctx, oldObject, object)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		return patch, nil
	}

	return h.newHttpHandler(mutator, mutateFunc)
}

// newHttpHandler returns a HTTP handler for mutating a request with the specified mutation
// function.
// This function is basically the same as the existing Handler func, with the only difference that
// it is now wrapped into the New...Handler funcs above in order to first decode the CR and check
// if it should be mutated by azure-admission-controller.
func (h *HttpHandlerFactory) newHttpHandler(webhookHandler WebhookHandlerBase, mutateFunc func(ctx context.Context, review v1beta1.AdmissionReview) ([]PatchOperation, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			webhookHandler.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %q", request.Header.Get("Content-Type")))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			webhookHandler.Log("level", "error", "message", "unable to read request", "stack", microerror.JSON(err))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			webhookHandler.Log("level", "error", "message", "unable to parse admission review request", "stack", microerror.JSON(err))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var patch []PatchOperation
		if review.Request.DryRun != nil && *review.Request.DryRun {
			webhookHandler.Log("level", "debug", "message", "Dry run is not supported. Request processing stopped.", "stack", microerror.JSON(err))
		} else {
			patch, err = mutateFunc(request.Context(), review)
			if err != nil {
				writeResponse(webhookHandler, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
				return
			}
		}

		resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, extractName(review.Request))
		patchData, err := json.Marshal(patch)
		if err != nil {
			webhookHandler.Log("level", "error", "message", fmt.Sprintf("unable to serialize patch for %s", resourceName), "stack", microerror.JSON(err))
			writeResponse(webhookHandler, writer, errorResponse(review.Request.UID, InternalError))
			return
		}

		webhookHandler.Log("level", "debug", "message", fmt.Sprintf("admitted %s (with %d patches)", resourceName, len(patch)))

		pt := v1beta1.PatchTypeJSONPatch
		writeResponse(webhookHandler, writer, &v1beta1.AdmissionResponse{
			Allowed:   true,
			UID:       review.Request.UID,
			Patch:     patchData,
			PatchType: &pt,
		})
	}
}
