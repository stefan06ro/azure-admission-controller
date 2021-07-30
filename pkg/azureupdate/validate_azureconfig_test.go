package azureupdate

import (
	"context"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
)

func TestMasterCIDR(t *testing.T) {
	testCases := []struct {
		name         string
		ctx          context.Context
		oldCIDR      string
		newCIDR      string
		oldAZs       []int
		newAZs       []int
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0: CIDR changed",
			ctx:  context.Background(),

			oldCIDR:      "10.0.1.0/24",
			newCIDR:      "10.0.2.0/24",
			errorMatcher: IsMasterCIDRChange,
		},
		{
			name: "case 1: CIDR unchanged",
			ctx:  context.Background(),

			oldCIDR:      "10.0.1.0/24",
			newCIDR:      "10.0.1.0/24",
			errorMatcher: nil,
		},
		{
			name: "case 2: CIDR was unset, being set",
			ctx:  context.Background(),

			oldCIDR:      "",
			newCIDR:      "10.0.1.0/24",
			errorMatcher: nil,
		},
		{
			name: "case 3: CIDR was set, being unset",
			ctx:  context.Background(),

			oldCIDR:      "10.0.1.0/24",
			newCIDR:      "",
			errorMatcher: IsMasterCIDRChange,
		},
		{
			name: "case 4: AZs changed",
			ctx:  context.Background(),

			oldAZs:       []int{1, 2, 3},
			newAZs:       []int{1},
			errorMatcher: IsAvailabilityZonesChange,
		},
		{
			name: "case 5: AZ unchanged",
			ctx:  context.Background(),

			oldAZs:       []int{1, 2, 3},
			newAZs:       []int{1, 2, 3},
			errorMatcher: nil,
		},
		{
			name: "case 6: AZ set for the first time",
			ctx:  context.Background(),

			oldAZs:       []int{},
			newAZs:       []int{1, 2, 3},
			errorMatcher: nil,
		},
		{
			name: "case 7: AZ set for the first time",
			ctx:  context.Background(),

			oldAZs:       nil,
			newAZs:       []int{1, 2, 3},
			errorMatcher: nil,
		},
		{
			name: "case 8: AZ allow different order",
			ctx:  context.Background(),

			oldAZs:       []int{3, 2, 1},
			newAZs:       []int{1, 2, 3},
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
			fakeCtrlClient, err := getFakeCtrlClient()
			if err != nil {
				panic(microerror.JSON(err))
			}

			handler, err := NewAzureConfigWebhookHandler(AzureConfigWebhookHandlerConfig{
				CtrlClient: fakeCtrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on AzureConfig update.
			err = handler.OnUpdateValidate(tc.ctx, azureConfigObj("13.0.0", tc.oldCIDR, tc.oldAZs), azureConfigObj("13.0.0", tc.newCIDR, tc.newAZs))

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

func TestAzureConfigValidate(t *testing.T) {
	releases := []ReleaseWithState{
		{
			Version: "11.3.0",
			State:   v1alpha1.StateDeprecated,
		},
		{
			Version: "11.3.1",
			State:   v1alpha1.StateActive,
		},
		{
			Version: "11.4.0",
			State:   v1alpha1.StateActive,
		},
		{
			Version: "12.0.0",
			State:   v1alpha1.StateActive,
		},
		{
			Version: "12.1.1",
			State:   v1alpha1.StateDeprecated,
		},
		{
			Version: "12.1.2",
			State:   v1alpha1.StateActive,
		},
		{
			Version: "13.0.1",
			State:   v1alpha1.StateDeprecated,
		},
		{
			Version: "13.0.2",
			Ignored: true,
			State:   v1alpha1.StateActive,
		},
		{
			Version: "13.1.0",
			State:   v1alpha1.StateActive,
		},
	}

	testCases := []struct {
		name         string
		ctx          context.Context
		releases     []ReleaseWithState
		oldVersion   string
		newVersion   string
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			errorMatcher: nil,
		},
		{
			name: "case 1",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
		{
			name: "case 2",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "12.0.0",
			errorMatcher: releaseversion.IsSkippingReleaseError,
		},
		{
			name: "case 3",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.0",
			errorMatcher: nil,
		},
		{
			name: "case 4",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
		{
			name: "case 5",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "",
			errorMatcher: errors.IsParsingFailed,
		},
		{
			name: "case 6",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "",
			newVersion:   "11.3.1",
			errorMatcher: errors.IsParsingFailed,
		},
		{
			name: "case 7",
			ctx:  context.Background(),

			releases:     []ReleaseWithState{{Version: "invalid", State: v1alpha1.StateActive}},
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			errorMatcher: errors.IsInvalidReleaseError,
		},
		{
			name: "case 8",
			ctx:  context.Background(),

			releases:     []ReleaseWithState{{Version: "invalid", State: v1alpha1.StateActive}},
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			errorMatcher: errors.IsInvalidReleaseError,
		},
		{
			name: "case 9",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.3.0",
			errorMatcher: releaseversion.IsDowngradingIsNotAllowedError,
		},
		{
			name: "case 10",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.0.0", // does not exist
			newVersion:   "11.3.0", // exists
			errorMatcher: nil,
		},
		{
			name: "case 11",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0", // exists
			newVersion:   "11.5.0", // does not exist
			errorMatcher: releaseversion.IsReleaseNotFoundError,
		},
		{
			name: "case 12",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.5.0", // does not exist
			newVersion:   "11.5.0", // does not exist
			errorMatcher: nil,
		},
		{
			name: "case 13",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
		{
			name: "case 14",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "12.1.2",
			newVersion:   "13.1.0",
			errorMatcher: nil,
		},
		{
			name: "case 15",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "12.0.0",
			newVersion:   "13.1.0",
			errorMatcher: releaseversion.IsSkippingReleaseError,
		},
		{
			name: "case 16",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "12.1.1",
			newVersion:   "13.1.0",
			errorMatcher: nil,
		},

		{
			name: "case 17",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "12.1.2",
			newVersion:   "13.1.0",
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
			fakeCtrlClient, err := getFakeCtrlClient()
			if err != nil {
				panic(microerror.JSON(err))
			}

			handler, err := NewAzureConfigWebhookHandler(AzureConfigWebhookHandlerConfig{
				CtrlClient: fakeCtrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Create needed releases.
			err = ensureReleases(fakeCtrlClient, tc.releases)
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on AzureConfig update.
			err = handler.OnUpdateValidate(tc.ctx, azureConfigObj(tc.oldVersion, "10.0.0.0/24", nil), azureConfigObj(tc.newVersion, "10.0.0.0/24", nil))

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

func azureConfigObj(version string, cidr string, azs []int) *providerv1alpha1.AzureConfig {
	azureConfig := providerv1alpha1.AzureConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureConfig",
			APIVersion: "provider.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": version,
			},
		},
		Spec: providerv1alpha1.AzureConfigSpec{
			Cluster: providerv1alpha1.Cluster{},
			Azure: providerv1alpha1.AzureConfigSpecAzure{
				AvailabilityZones: azs,
				VirtualNetwork: providerv1alpha1.AzureConfigSpecAzureVirtualNetwork{
					MasterSubnetCIDR: cidr,
				},
			},
			VersionBundle: providerv1alpha1.AzureConfigSpecVersionBundle{},
		},
	}

	return &azureConfig
}
