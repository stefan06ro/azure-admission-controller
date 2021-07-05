package mutator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type WebhookHandlerMock struct {
	DecodeFunc func(runtime.RawExtension) (metav1.ObjectMetaAccessor, error)
}

func (h *WebhookHandlerMock) Log(_ ...interface{}) {}

func (h *WebhookHandlerMock) Resource() string {
	return "mock_type"
}

func (h *WebhookHandlerMock) Decode(object runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	return h.DecodeFunc(object)
}

func (h *WebhookHandlerMock) OnCreateMutate(_ context.Context, _ interface{}) ([]PatchOperation, error) {
	return nil, nil
}

func (h *WebhookHandlerMock) OnUpdateMutate(_ context.Context, _ interface{}, _ interface{}) ([]PatchOperation, error) {
	return nil, nil
}
