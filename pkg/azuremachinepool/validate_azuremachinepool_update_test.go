package azuremachinepool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
)

func TestAzureMachinePoolUpdateValidate(t *testing.T) {
	tr := true
	fa := false
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
		oldNodePool  []byte
		newNodePool  []byte
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: AcceleratedNetworking is enabled in CR and we don't change it or the instance type",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: nil,
		},
		{
			name:         "case 1: AcceleratedNetworking is disabled in CR and we don't change it or the instance type",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: nil,
		},
		{
			name:         "case 2: Enabled and try disabling it, keeping same instance type",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 3: Enabled, try updating to new instance type that supports it",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[1], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: nil,
		},
		{
			name:         "case 4: Enabled, try updating to new instance type that does NOT supports it",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(unsupportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 5: Disabled and try enabling it",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 6: changed from nil to true",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 7: changed from true to nil",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &tr, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 8: changed from nil to false",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 9: changed from false to nil",
			oldNodePool:  azureMPRawObject(supportedInstanceType[0], &fa, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(supportedInstanceType[0], nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 10: changed from premium to standard storage",
			oldNodePool:  azureMPRawObject(premiumStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 11: changed from standard to premium storage",
			oldNodePool:  azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(premiumStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			errorMatcher: nil,
		},
		{
			name:         "case 12: change storage account type",
			oldNodePool:  azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(premiumStorageInstanceType, nil, string(compute.StorageAccountTypesPremiumLRS), desiredDataDisks, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:        "case 13: change datadisks",
			oldNodePool: azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool: azureMPRawObject(premiumStorageInstanceType, nil, string(compute.StorageAccountTypesPremiumLRS), []capzv1alpha3.DataDisk{
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
			}, "westeurope"),
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 14: changed location",
			oldNodePool:  azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "westeurope"),
			newNodePool:  azureMPRawObject(standardStorageInstanceType, nil, string(compute.StorageAccountTypesStandardLRS), desiredDataDisks, "northeastitaly"),
			errorMatcher: IsInvalidOperationError,
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
			stubAPI := NewStubAPI(stubbedSKUs)
			vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
				Azure:  stubAPI,
				Logger: newLogger,
			})
			if err != nil {
				panic(microerror.JSON(err))
			}

			admit := &UpdateValidator{
				logger: newLogger,
				vmcaps: vmcaps,
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(context.Background(), getUpdateAdmissionRequest(tc.oldNodePool, tc.newNodePool))

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

func getUpdateAdmissionRequest(oldMP []byte, newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    oldMP,
			Object: nil,
		},
	}

	return req
}

func azureMPRawObject(vmSize string, acceleratedNetworkingEnabled *bool, storageAccountType string, dataDisks []capzv1alpha3.DataDisk, location string) []byte {
	mp := v1alpha3.AzureMachinePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureMachinePool",
			APIVersion: "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"giantswarm.io/cluster":                "ab123",
				"giantswarm.io/machine-pool":           "ab123",
				"giantswarm.io/organization":           "giantswarm",
				"release.giantswarm.io/version":        "13.0.0",
			},
		},
		Spec: v1alpha3.AzureMachinePoolSpec{
			Location: location,
			Template: v1alpha3.AzureMachineTemplate{
				VMSize: vmSize,
				OSDisk: capzv1alpha3.OSDisk{
					ManagedDisk: capzv1alpha3.ManagedDisk{
						StorageAccountType: storageAccountType,
					},
				},
				AcceleratedNetworking: acceleratedNetworkingEnabled,
				DataDisks:             dataDisks,
			},
		},
	}
	byt, _ := json.Marshal(mp)
	return byt
}
