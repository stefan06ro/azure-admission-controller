//go:build liveinstallation

package validateliveresources

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/azure-admission-controller/integration/env"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func NewReadOnlyCtrlClient(t *testing.T) client.Client {
	var err error

	schemeBuilder := runtime.SchemeBuilder{
		v1alpha1.AddToScheme,
		corev1alpha1.AddToScheme,
		corev1.AddToScheme,
		capi.AddToScheme,
		capiexp.AddToScheme,
		capz.AddToScheme,
		capzexp.AddToScheme,
		releasev1alpha1.AddToScheme,
		securityv1alpha1.AddToScheme,
	}

	var restConfig *rest.Config
	{
		restConfig, err = clientcmd.BuildConfigFromFlags("", env.KubeConfig())
		if err != nil {
			t.Fatal(err)
		}
	}

	runtimeScheme := runtime.NewScheme()
	{
		err = schemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			t.Fatal(err)
		}
	}

	mapper, err := apiutil.NewDynamicRESTMapper(rest.CopyConfig(restConfig))
	if err != nil {
		t.Fatal(err)
	}

	ctrlClient, err := client.New(rest.CopyConfig(restConfig), client.Options{Scheme: runtimeScheme, Mapper: mapper})
	if err != nil {
		t.Fatal(err)
	}

	readOnlyClient := &ReadOnlyCtrlClient{
		t:      t,
		client: ctrlClient,
	}

	return readOnlyClient
}

func NewDecoder() runtime.Decoder {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	return codecs.UniversalDeserializer()
}

func NewVMCapabilities(t *testing.T, logger micrologger.Logger) *vmcapabilities.VMSKU {
	var err error

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

	return vmCapabilities
}

func SetAzureEnvironmentVariables(t *testing.T, ctx context.Context, ctrlReader client.Reader) {
	var err error
	objectKey := client.ObjectKey{
		Namespace: "giantswarm",
		Name:      "credential-default",
	}

	var secret corev1.Secret
	err = ctrlReader.Get(ctx, objectKey, &secret)
	if err != nil {
		t.Fatal(err)
	}

	// Set AZURE_ENVIRONMENT
	err = os.Setenv(auth.EnvironmentName, "AzurePublicCloud")
	if err != nil {
		t.Fatal(err)
	}

	// Set AZURE_TENANT_ID
	err = os.Setenv(auth.TenantID, string(secret.Data["azure.azureoperator.tenantid"]))
	if err != nil {
		t.Fatal(err)
	}

	// Set AZURE_SUBSCRIPTION_ID
	err = os.Setenv(auth.SubscriptionID, string(secret.Data["azure.azureoperator.subscriptionid"]))
	if err != nil {
		t.Fatal(err)
	}

	// Set AZURE_CLIENT_ID
	err = os.Setenv(auth.ClientID, string(secret.Data["azure.azureoperator.clientid"]))
	if err != nil {
		t.Fatal(err)
	}

	// Set AZURE_CLIENT_SECRET
	err = os.Setenv(auth.ClientSecret, string(secret.Data["azure.azureoperator.clientsecret"]))
	if err != nil {
		t.Fatal(err)
	}
}
