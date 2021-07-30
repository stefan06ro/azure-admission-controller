package machinepool

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	builder "github.com/giantswarm/azure-admission-controller/internal/test/machinepool"
	"github.com/giantswarm/azure-admission-controller/internal/vmcapabilities"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestMachinePoolUpdateValidate(t *testing.T) {
	type testCase struct {
		name         string
		oldNodePool  *capiexp.MachinePool
		newNodePool  *capiexp.MachinePool
		errorMatcher func(err error) bool
	}

	testCases := []testCase{
		{
			name:         "case 0: FailureDomains unchanged",
			oldNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"1", "2"})),
			newNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"1", "2"})),
			errorMatcher: nil,
		},
		{
			name:         "case 1: FailureDomains changed",
			oldNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"1"})),
			newNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"2"})),
			errorMatcher: IsFailureDomainWasChangedError,
		},
		{
			name:         "case 2: FailureDomains changed but object is being deleted",
			oldNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"1"})),
			newNodePool:  builder.BuildMachinePool(builder.FailureDomains([]string{"2"}), builder.WithDeletionTimestamp()),
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

			stubAPI := unittest.NewEmptyResourceSkuStubAPI()
			vmcaps, err := vmcapabilities.New(vmcapabilities.Config{
				Azure:  stubAPI,
				Logger: newLogger,
			})
			if err != nil {
				t.Fatal(microerror.JSON(err))
			}

			handler, err := NewWebhookHandler(WebhookHandlerConfig{
				CtrlClient: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
				VMcaps:     vmcaps,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on MachinePool update.
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
