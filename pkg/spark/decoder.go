package spark

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type Decoder struct{}

func (d *Decoder) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	sparkCR := &v1alpha1.Spark{}
	if _, _, err := mutator.Deserializer.Decode(rawObject.Raw, nil, sparkCR); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse Spark CR: %v", err)
	}

	return sparkCR, nil
}
