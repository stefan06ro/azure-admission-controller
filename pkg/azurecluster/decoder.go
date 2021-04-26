package azurecluster

import (
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type Decoder struct{}

func (d *Decoder) Decode(rawObject runtime.RawExtension) (metav1.ObjectMetaAccessor, error) {
	azureClusterCR := &capzv1alpha3.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(rawObject.Raw, nil, azureClusterCR); err != nil {
		return nil, microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	return azureClusterCR, nil
}
