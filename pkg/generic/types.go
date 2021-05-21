package generic

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Decoder interface {
	Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error)
}

type Logger interface {
	Log(keyVals ...interface{})
}
