package mutator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GenericObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}
