package azureupdate

import (
	"context"
	"fmt"
	"testing"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/pkg/unittest"
)

func TestAzureClusterConfigValidate(t *testing.T) {
	releases := []ReleaseWithState{
		{
			Version: "11.3.0",
			State:   releasev1alpha1.StateDeprecated,
		},
		{
			Version: "11.3.1",
			State:   releasev1alpha1.StateActive,
		},
		{
			Version: "11.4.0",
			State:   releasev1alpha1.StateActive,
		},
		{
			Version: "12.0.0",
			State:   releasev1alpha1.StateActive,
		},
		{
			Version: "12.0.1",
			State:   releasev1alpha1.StateDeprecated,
		},
		{
			Version: "12.1.0",
			State:   releasev1alpha1.StateActive,
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

			releases:     []ReleaseWithState{{Version: "invalid", State: releasev1alpha1.StateActive}},
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			errorMatcher: errors.IsInvalidReleaseError,
		},
		{
			name: "case 8",
			ctx:  context.Background(),

			releases:     []ReleaseWithState{{Version: "invalid", State: releasev1alpha1.StateActive}},
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
			newVersion:   "12.0.1",
			errorMatcher: nil,
		},
		{
			name: "case 14",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "12.1.0",
			errorMatcher: releaseversion.IsSkippingReleaseError,
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

			scheme := runtime.NewScheme()
			err = v1.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = corev1alpha1.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = capiexp.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = capzexp.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = capi.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = capz.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = providerv1alpha1.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}
			err = releasev1alpha1.AddToScheme(scheme)
			if err != nil {
				panic(err)
			}

			ctrlClient := fake.NewFakeClientWithScheme(scheme)

			handler, err := NewAzureClusterConfigWebhookHandler(AzureClusterConfigWebhookHandlerConfig{
				CtrlClient: ctrlClient,
				Decoder:    unittest.NewFakeDecoder(),
				Logger:     newLogger,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Create needed releases.
			err = ensureReleases(ctrlClient, tc.releases)
			if err != nil {
				t.Fatal(err)
			}

			// Create AzureConfigs.
			ac := &providerv1alpha1.AzureConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      controlPlaneName,
					Namespace: controlPlaneNameSpace,
				},
				Spec: providerv1alpha1.AzureConfigSpec{},
			}
			err = ctrlClient.Create(tc.ctx, ac)
			if err != nil {
				t.Fatal(err)
			}

			// Run validating webhook handler on AzureClusterConfig update.
			err = handler.OnUpdateValidate(tc.ctx, azureClusterConfigObj(tc.oldVersion), azureClusterConfigObj(tc.newVersion))

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

func azureClusterConfigObj(version string) *corev1alpha1.AzureClusterConfig {
	azureClusterConfig := corev1alpha1.AzureClusterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureClusterConfig",
			APIVersion: "core.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-azure-cluster-config", controlPlaneName),
			Namespace: controlPlaneNameSpace,
		},
		Spec: corev1alpha1.AzureClusterConfigSpec{
			Guest: corev1alpha1.AzureClusterConfigSpecGuest{
				ClusterGuestConfig: corev1alpha1.ClusterGuestConfig{
					ReleaseVersion: version,
					ID:             controlPlaneName,
				},
			},
			VersionBundle: corev1alpha1.AzureClusterConfigSpecVersionBundle{},
		},
	}

	return &azureClusterConfig
}
