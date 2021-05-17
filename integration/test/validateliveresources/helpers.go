// +build liveinstallation,validate

package validateliveresources

import (
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/azure-admission-controller/integration/env"
)

func NewCtrlClient(schemeBuilder runtime.SchemeBuilder) (client.Client, error) {
	var err error

	var restConfig *rest.Config
	{
		restConfig, err = clientcmd.BuildConfigFromFlags("", env.KubeConfig())
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	runtimeScheme := runtime.NewScheme()
	{
		err = schemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	mapper, err := apiutil.NewDynamicRESTMapper(rest.CopyConfig(restConfig))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ctrlClient, err := client.New(rest.CopyConfig(restConfig), client.Options{Scheme: runtimeScheme, Mapper: mapper})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return ctrlClient, nil
}