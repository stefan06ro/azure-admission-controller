package release

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetComponentVersionsFromRelease(ctx context.Context, ctrlReader client.Reader, releaseVersion string) (map[string]string, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}

	// Retrieve the `Release` CR.
	release := &releasev1alpha1.Release{}
	{
		err := ctrlReader.Get(ctx, client.ObjectKey{Name: releaseVersion}, release)
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(releaseNotFoundError, "Looking for Release %s but it was not found. Can't continue.", releaseVersion)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	ret := map[string]string{}
	// Search the desired component.
	for _, component := range release.Spec.Components {
		ret[component.Name] = component.Version
	}

	return ret, nil
}
