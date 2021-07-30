package unittest

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type FakeDecoder struct{}

func NewFakeDecoder() *FakeDecoder {
	return &FakeDecoder{}
}

func (d *FakeDecoder) Decode(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	return nil, nil, nil
}
