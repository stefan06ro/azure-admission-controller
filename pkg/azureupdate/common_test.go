package azureupdate

import (
	"context"
	"fmt"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureReleases(g8sclient versioned.Interface, releases []string) error {
	// Create Releases.
	for _, release := range releases {
		req := &releasev1alpha1.Release{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("v%s", release),
			},
		}

		_, err := g8sclient.ReleaseV1alpha1().Releases().Create(context.Background(), req, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
