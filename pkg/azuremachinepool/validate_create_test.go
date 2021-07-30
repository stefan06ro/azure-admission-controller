package azuremachinepool

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/azuremachinepool"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureMachinePoolCreateValidate(t *testing.T) {
	unsupportedInstanceType := []string{
		"Standard_A2_v2",
		"Standard_A4_v2",
		"Standard_A8_v2",
		"Standard_D2_v3",
		"Standard_D2s_v3",
	}
	type testCase struct {
		name         string
		nodePool     *capzexp.AzureMachinePool
		errorMatcher func(err error) bool
	}

	var testCases []testCase

	for i, instanceType := range unsupportedInstanceType {
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", i*3, instanceType),
			nodePool:     builder.BuildAzureMachinePool(builder.VMSize(instanceType), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: IsVmsizeDoesNotSupportAcceleratedNetworkingError,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking disabled", i*3+1, instanceType),
			nodePool:     builder.BuildAzureMachinePool(builder.VMSize(instanceType), builder.AcceleratedNetworking(to.BoolPtr(false))),
			errorMatcher: nil,
		})

		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking nil", i*3+2, instanceType),
			nodePool:     builder.BuildAzureMachinePool(builder.VMSize(instanceType), builder.AcceleratedNetworking(nil)),
			errorMatcher: nil,
		})
	}

	// Non existing instance type.
	{
		instanceType := "this_is_a_random_name"
		testCases = append(testCases, testCase{
			name:         fmt.Sprintf("case %d: instance type %s with accelerated networking enabled", len(testCases), instanceType),
			nodePool:     builder.BuildAzureMachinePool(builder.VMSize(instanceType), builder.AcceleratedNetworking(to.BoolPtr(true))),
			errorMatcher: vmcapabilities.IsSkuNotFoundError,
		})
	}

	{
		testCases = append(testCases, testCase{
			name: fmt.Sprintf("case %d: data disks already set", len(testCases)),
			nodePool: builder.BuildAzureMachinePool(builder.VMSize("Standard_D4_v3"), builder.DataDisks([]capz.DataDisk{
				{
					NameSuffix: "docker",
					DiskSizeGB: 50,
					Lun:        to.Int32Ptr(21),
				},
				{
					NameSuffix: "kubelet",
					DiskSizeGB: 50,
					Lun:        to.Int32Ptr(22),
				},
			})),
			errorMatcher: IsDatadisksFieldIsSetError,
		})
	}

	testCases = append(testCases, testCase{
		name:         fmt.Sprintf("case %d: invalid location", len(testCases)-1),
		nodePool:     builder.BuildAzureMachinePool(builder.VMSize("Standard_D4_v3"), builder.Location("eastgalicia")),
		errorMatcher: IsUnexpectedLocationError,
	})

	testCases = append(testCases, testCase{
		name:         fmt.Sprintf("case %d: invalid organization", len(testCases)-1),
		nodePool:     builder.BuildAzureMachinePool(builder.VMSize("Standard_D4_v3"), builder.Organization("wrongorg")),
		errorMatcher: generic.IsNodepoolOrgDoesNotMatchClusterOrg,
	})

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

			// Create cluster CR.
			cluster := &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ab123",
					Labels: map[string]string{
						label.Cluster:      "ab123",
						label.Organization: "giantswarm",
					},
				},
			}
			err = ctrlClient.Create(ctx, cluster)
			if err != nil {
				t.Fatal(err)
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
				},
				"Standard_D2_v3": {
					Name: to.StringPtr("Standard_D2_v3"),
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
				},
				"Standard_D2s_v3": {
					Name: to.StringPtr("Standard_D2s_v3"),
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
				},
				"Standard_D4_v3": {
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

			// Run validating webhook handler on AzureMachinePool creation.
			err = handler.OnCreateValidate(ctx, tc.nodePool)

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

type StubAPI struct {
	stubbedSKUs map[string]compute.ResourceSku
}

func NewStubAPI(stubbedSKUs map[string]compute.ResourceSku) vmcapabilities.API {
	return &StubAPI{stubbedSKUs: stubbedSKUs}
}

func (s *StubAPI) List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error) {
	return s.stubbedSKUs, nil
}
