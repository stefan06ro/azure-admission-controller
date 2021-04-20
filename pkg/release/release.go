package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	gocache "github.com/patrickmn/go-cache"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// Cache of release components
	// key: release version
	// value: map "component name": "component version"
	releaseComponentsCache = gocache.New(24*time.Hour, 24*time.Hour)
)

func GetComponentVersionsFromRelease(ctx context.Context, ctrlClient client.Client, releaseVersion string) (map[string]string, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}

	var components map[string]string
	{
		cachedComponents, ok := releaseComponentsCache.Get(releaseVersion)

		if ok {
			// Release components are found in cache
			components = cachedComponents.(map[string]string)
		} else {
			// Retrieve the `Release` CR.
			release := &releasev1alpha1.Release{}
			{
				err := ctrlClient.Get(ctx, client.ObjectKey{Name: releaseVersion, Namespace: "default"}, release)
				if apierrors.IsNotFound(err) {
					return nil, microerror.Maskf(releaseNotFoundError, "Looking for Release %s but it was not found. Can't continue.", releaseVersion)
				} else if err != nil {
					return nil, microerror.Mask(err)
				}
			}

			components = map[string]string{}
			// Search the desired component.
			for _, component := range release.Spec.Components {
				components[component.Name] = component.Version
			}

			// Save in cache
			releaseComponentsCache.Set(releaseVersion, components, gocache.DefaultExpiration)
		}
	}

	return components, nil
}

func ContainsAzureOperator(ctx context.Context, ctrlClient client.Client, releaseVersion string) (bool, error) {
	componentVersions, err := GetComponentVersionsFromRelease(ctx, ctrlClient, releaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if componentVersions["azure-operator"] == "" {
		return false, nil
	}

	return true, nil
}

