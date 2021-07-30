package azuremachinepool

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/resource"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachinePoolUpdateValidate(t *testing.T) {
	unsupportedInstanceType := []string{
		"Standard_D16_v3",
	}
	supportedInstanceType := []string{
		"Standard_D4_v3",
		"Standard_D8_v3",
	}
	premiumStorageInstanceType := "Standard_D4s_v3"
	standardStorageInstanceType := "Standard_D4_v3"
	type testCase struct {
		name         string
		oldNodePool  *capzexp.AzureMachinePool
		newNodePool  *capzexp.AzureMachinePool
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: AcceleratedNetworking is enabled in CR and we don't change it or the instance type",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: nil,
		},
		{
			name:         "case 1: AcceleratedNetworking is disabled in CR and we don't change it or the instance type",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			errorMatcher: nil,
		},
		{
			name:         "case 2: Enabled and try disabling it, keeping same instance type",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 3: Enabled, try updating to new instance type that supports it",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[1]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: nil,
		},
		{
			name:         "case 4: Enabled, try updating to new instance type that does NOT supports it",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(unsupportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: IsVmsizeDoesNotSupportAcceleratedNetworkingError,
		},
		{
			name:         "case 5: Disabled and try enabling it",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 6: changed from nil to true",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(nil)),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 7: changed from true to nil",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(true))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(nil)),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 8: changed from nil to false",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(nil)),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 9: changed from false to nil",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(to.BoolPtr(false))),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(supportedInstanceType[0]), builder.AcceleratedNetworking(nil)),
			errorMatcher: IsAcceleratedNetworkingWasChangedError,
		},
		{
			name:         "case 10: changed from premium to standard storage",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(premiumStorageInstanceType)),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(standardStorageInstanceType)),
			errorMatcher: IsSwitchToVmSizeThatDoesNotSupportAcceleratedNetworkingError,
		},
		{
			name:         "case 11: changed from standard to premium storage",
			oldNodePool:  builder.BuildAzureMachinePool(builder.VMSize(standardStorageInstanceType)),
			newNodePool:  builder.BuildAzureMachinePool(builder.VMSize(premiumStorageInstanceType)),
			errorMatcher: nil,
		},
		{
			name:         "case 12: change storage account type",
			oldNodePool:  builder.BuildAzureMachinePool(builder.StorageAccountType(compute.StorageAccountTypesStandardLRS)),
			newNodePool:  builder.BuildAzureMachinePool(builder.StorageAccountType(compute.StorageAccountTypesPremiumLRS)),
			errorMatcher: IsStorageAccountWasChangedError,
		},
		{
			name:        "case 13: change datadisks",
			oldNodePool: builder.BuildAzureMachinePool(),
			newNodePool: builder.BuildAzureMachinePool(builder.DataDisks([]capz.DataDisk{
				{
					NameSuffix: "docker",
					DiskSizeGB: 30,
					Lun:        to.Int32Ptr(21),
				},
				{
					NameSuffix: "kubelet",
					DiskSizeGB: 50,
					Lun:        to.Int32Ptr(22),
				},
			})),
			errorMatcher: IsDatadisksFieldIsSetError,
		},
		{
			name:         "case 14: changed location",
			oldNodePool:  builder.BuildAzureMachinePool(builder.Location("westeurope")),
			newNodePool:  builder.BuildAzureMachinePool(builder.Location("northeastitaly")),
			errorMatcher: IsLocationWasChangedError,
		},
		{
			name:         "case 15: disable spot instance configuration",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("-1")})),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(nil)),
			errorMatcher: IsSpotVMOptionsWasChangedError,
		},
		{
			name:         "case 16: enable spot instance configuration",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(nil)),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("-1")})),
			errorMatcher: IsSpotVMOptionsWasChangedError,
		},
		{
			name:         "case 17: change spot instance price configuration",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("1.24322")})),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("-1")})),
			errorMatcher: IsSpotVMOptionsWasChangedError,
		},
		{
			name:         "case 18: keep spot instances disabled",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(nil)),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(nil)),
			errorMatcher: nil,
		},
		{
			name:         "case 19: keep spot instances price configuration unknown",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: nil})),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: nil})),
			errorMatcher: nil,
		},
		{
			name:         "case 20: keep spot instances price configuration unchanged",
			oldNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("1.24322")})),
			newNodePool:  builder.BuildAzureMachinePool(builder.SpotVMOptions(&capz.SpotVMOptions{MaxPrice: toQuantityPtr("1.24322")})),
			errorMatcher: nil,
		},
		{
			name:         "case 21: changed location but object is being deleted",
			oldNodePool:  builder.BuildAzureMachinePool(builder.Location("westeurope")),
			newNodePool:  builder.BuildAzureMachinePool(builder.Location("northeastitaly"), builder.WithDeletionTimestamp()),
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			stubbedSKUs := map[string]compute.ResourceSku{
				"Standard_D4_v3": {
					Name: to.StringPtr("Standard_D4_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("True"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("False"),
						},
					},
				},
				"Standard_D4s_v3": {
					Name: to.StringPtr("Standard_D4s_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("True"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("True"),
						},
					},
				},
				"Standard_D8_v3": {
					Name: to.StringPtr("Standard_D8_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("True"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("False"),
						},
					},
				},
				"Standard_D16_v3": {
					Name: to.StringPtr("Standard_D16_v3"),
					Capabilities: &[]compute.ResourceSkuCapabilities{
						{
							Name:  to.StringPtr("AcceleratedNetworkingEnabled"),
							Value: to.StringPtr("False"),
						},
						{
							Name:  to.StringPtr("vCPUs"),
							Value: to.StringPtr("4"),
						},
						{
							Name:  to.StringPtr("MemoryGB"),
							Value: to.StringPtr("16"),
						},
						{
							Name:  to.StringPtr("PremiumIO"),
							Value: to.StringPtr("False"),
						},
					},
				},
			}
			stubAPI := unittest.NewResourceSkuStubAPI(stubbedSKUs)
			vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
				Azure:  stubAPI,
				Logger: newLogger,
			})
			if err != nil {
				panic(microerror.JSON(err))
			}

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				CtrlClient: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Location:   "westeurope",
				Logger:     newLogger,
				VMcaps:     vmcaps,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on AzureMachinePool update.
			err = handler.OnUpdateValidate(ctx, tc.oldNodePool, tc.newNodePool)

			// Check if the error is the expected one.
			switch {
			case err == nil && tc.errorMatcher == nil:
				// fall through
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("expected %#v got %#v", nil, err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("expected %#v got %#v", "error", nil)
			case !tc.errorMatcher(err):
				t.Fatalf("unexpected error: %#v", err)
			}
		})
	}
}

func toQuantityPtr(d string) *resource.Quantity {
	r := resource.MustParse(d)

	return &r
}
