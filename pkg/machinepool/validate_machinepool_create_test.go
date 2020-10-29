package machinepool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	expcapzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

const (
	machinePoolNamespace = "default"
	machinePoolName      = "ab123"
)

func TestMachinePoolCreateValidate(t *testing.T) {
	type testCase struct {
		name         string
		machinePool  []byte
		vmType       string
		allowed      bool
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: instance type supporting [1,2,3], requested [1]",
			machinePool:  machinePoolRawObject([]string{"1"}),
			vmType:       "Standard_A2_v2",
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name:         "case 1: instance type supporting [1,2], requested [3]",
			machinePool:  machinePoolRawObject([]string{"3"}),
			vmType:       "Standard_A4_v2",
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 2: instance type supporting [1,2], requested [2,3]",
			machinePool:  machinePoolRawObject([]string{"2", "3"}),
			vmType:       "Standard_A4_v2",
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 3: instance type supporting [], requested [1]",
			machinePool:  machinePoolRawObject([]string{"1"}),
			vmType:       "Standard_A8_v2",
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name:         "case 4: instance type supporting [], requested []",
			machinePool:  machinePoolRawObject([]string{}),
			vmType:       "Standard_A8_v2",
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name:         "case 5: AzureMachinePool does not exist",
			machinePool:  machinePoolRawObject([]string{}),
			vmType:       "",
			allowed:      false,
			errorMatcher: IsAzureMachinePoolNotFound,
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
				"Standard_A2_v2": {
					Name: to.StringPtr("Standard_A2_v2"),
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
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{"1", "2", "3"},
						},
					},
				},
				"Standard_A4_v2": {
					Name: to.StringPtr("Standard_A4_v2"),
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
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{"1", "2"},
						},
					},
				},
				"Standard_A8_v2": {
					Name: to.StringPtr("Standard_A8_v2"),
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
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("westeurope"),
							Zones:    &[]string{},
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

			ctx := context.Background()
			fakeK8sClient := unittest.FakeK8sClient()
			ctrlClient := fakeK8sClient.CtrlClient()

			// Create AzureMachinePool.
			if tc.vmType != "" {
				amp := &expcapzv1alpha3.AzureMachinePool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      machinePoolName,
						Namespace: machinePoolNamespace,
					},
					Spec: expcapzv1alpha3.AzureMachinePoolSpec{
						Location: "westeurope",
						Template: expcapzv1alpha3.AzureMachineTemplate{
							VMSize: tc.vmType,
						},
					},
				}
				err = ctrlClient.Create(ctx, amp)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Create default GiantSwarm organization.
			organization := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
				Spec: securityv1alpha1.OrganizationSpec{},
			}
			err = ctrlClient.Create(ctx, organization)
			if err != nil {
				t.Fatal(err)
			}

			admit := &CreateValidator{
				ctrlClient: ctrlClient,
				logger:     newLogger,
				vmcaps:     vmcaps,
			}

			// Run admission request to validate AzureConfig updates.
			allowed, err := admit.Validate(ctx, getCreateAdmissionRequest(tc.machinePool))

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

			// Check if the validation result is the expected one.
			if tc.allowed != allowed {
				t.Fatalf("expected %v to be equal to %v", tc.allowed, allowed)
			}
		})
	}
}

type StubAPI struct {
	stubbedSKUs map[string]compute.ResourceSku
}

func NewStubAPI(stubbedSKUs map[string]compute.ResourceSku) vmcapabilities.API {
	return &StubAPI{stubbedSKUs: stubbedSKUs}
}

func (s *StubAPI) List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error) {
	return s.stubbedSKUs, nil
}

func getCreateAdmissionRequest(newMP []byte) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Resource: metav1.GroupVersionResource{
			Version:  "exp.infrastructure.cluster.x-k8s.io/v1alpha3",
			Resource: "azuremachinepool",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    newMP,
			Object: nil,
		},
	}

	return req
}

func machinePoolRawObject(failureDomains []string) []byte {
	mp := capiv1alpha3.MachinePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachinePool",
			APIVersion: "exp.cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      machinePoolName,
			Namespace: machinePoolNamespace,
			Labels: map[string]string{
				"azure-operator.giantswarm.io/version": "5.0.0",
				"giantswarm.io/cluster":                machinePoolName,
				"giantswarm.io/machine-pool":           machinePoolName,
				"giantswarm.io/organization":           "giantswarm",
				"release.giantswarm.io/version":        "13.0.0",
			},
		},
		Spec: capiv1alpha3.MachinePoolSpec{
			FailureDomains: failureDomains,
			Template: v1alpha3.MachineTemplateSpec{
				Spec: v1alpha3.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Namespace: machinePoolNamespace,
						Name:      machinePoolName,
					},
				},
			},
		},
	}
	byt, _ := json.Marshal(mp)
	return byt
}
