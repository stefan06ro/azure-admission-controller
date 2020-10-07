package azureupdate

import (
	"context"
	"fmt"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureReleases(ctrlClient client.Client, releases []string) error {
	// Create Releases.
	for _, release := range releases {
		req := &releasev1alpha1.Release{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("v%s", release),
			},
		}

		err := ctrlClient.Create(context.Background(), req)
		if err != nil {
			return err
		}
	}

	return nil
}
