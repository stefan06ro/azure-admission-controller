package azureupdate

import (
	"context"
	"fmt"
	"strconv"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReleaseWithState struct {
	Version string
	Ignored bool
	State   releasev1alpha1.ReleaseState
}

func ensureReleases(ctrlClient client.Client, releases []ReleaseWithState) error {
	// Create Releases.
	for _, release := range releases {
		req := &releasev1alpha1.Release{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("v%s", release.Version),
				Annotations: map[string]string{
					"release.giantswarm.io/ignore": strconv.FormatBool(release.Ignored),
				},
			},
			Spec: releasev1alpha1.ReleaseSpec{
				State: release.State,
			},
		}

		err := ctrlClient.Create(context.Background(), req)
		if err != nil {
			return err
		}
	}

	return nil
}
