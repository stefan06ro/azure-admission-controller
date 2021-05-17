// +build liveinstallation,validate

package validateliveresources

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	machinepoolpkg "github.com/giantswarm/azure-admission-controller/pkg/machinepool"
)

func TestMachinePools(t *testing.T) {
	var err error

	ctx := context.Background()

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatal(err)
	}

	schemeBuilder := runtime.SchemeBuilder{
		capi.AddToScheme,
		capiexp.AddToScheme,
		capz.AddToScheme,
		capzexp.AddToScheme,
		releasev1alpha1.AddToScheme,
		securityv1alpha1.AddToScheme,
	}

	ctrlClient, err := NewCtrlClient(schemeBuilder)
	if err != nil {
		t.Fatal(err)
	}

	var resourceSkusClient compute.ResourceSkusClient
	{
		settings, err := auth.GetSettingsFromEnvironment()
		if err != nil {
			t.Fatal(err)
		}
		authorizer, err := settings.GetAuthorizer()
		if err != nil {
			t.Fatal(err)
		}
		resourceSkusClient = compute.NewResourceSkusClient(settings.GetSubscriptionID())
		resourceSkusClient.Client.Authorizer = authorizer
	}

	var vmCapabilities *vmcapabilities.VMSKU
	{
		vmCapabilities, err = vmcapabilities.New(vmcapabilities.Config{
			Logger: logger,
			Azure:  vmcapabilities.NewAzureAPI(vmcapabilities.AzureConfig{ResourceSkuClient: &resourceSkusClient}),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	var machinePoolList capiexp.MachinePoolList
	err = ctrlClient.List(ctx, &machinePoolList)
	if err != nil {
		t.Fatal(err)
	}

	var machinePoolWebhookHandler *machinepoolpkg.WebhookHandler
	{
		c := machinepoolpkg.WebhookHandlerConfig{
			CtrlClient: ctrlClient,
			Logger:     logger,
			VMcaps:     vmCapabilities,
		}
		machinePoolWebhookHandler, err = machinepoolpkg.NewWebhookHandler(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, machinePool := range machinePoolList.Items {
		err = machinePoolWebhookHandler.OnCreateValidate(ctx, &machinePool)
		if err != nil {
			t.Fatal(err)
		}

		updatedMachinePool := machinePool.DeepCopy()

		updatedMachinePool.Labels["test.giantswarm.io/dummy"] = "this is not really saved"
		err = machinePoolWebhookHandler.OnUpdateValidate(ctx, &machinePool, updatedMachinePool)
		if err != nil {
			t.Fatal(err)
		}
	}
}
