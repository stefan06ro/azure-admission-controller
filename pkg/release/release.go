package release

import (
	"context"
	"fmt"
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	azureOperatorComponentName = "azure-operator"
)

// TryFindReleaseForObject tries to find a Release CR for the specified object.
//
// In order to get the release it needs a release name which is set in the CR label
// release.giantswarm.io/version, so it first tries to get that label from the object. If the label
// is not set, it uses the specified clusterGetter to get Cluster to which the CR belongs to, and
// then it tries to get release.giantswarm.io/version label from the Cluster CR.
//
// Finally, it fetches the release CR by the release name.
func TryFindReleaseForObject(ctx context.Context, ctrlReader client.Reader, objectMeta metav1.Object, ownerClusterGetter func(metav1.Object) capi.Cluster) (releasev1alpha1.Release, bool, error) {
	// Try to get release label from the CR
	var releaseVersionLabel string
	if objectMeta.GetLabels() != nil && objectMeta.GetLabels()[label.ReleaseVersion] != "" {
		// Get release label from the object itself
		releaseVersionLabel = objectMeta.GetLabels()[label.ReleaseVersion]
	} else {
		// Get release label from the Cluster CR
		cluster := ownerClusterGetter(objectMeta)
		if cluster.Labels != nil && cluster.Labels[label.ReleaseVersion] != "" {
			releaseVersionLabel = cluster.Labels[label.ReleaseVersion]
		}
	}

	// check if we have found the release label
	if releaseVersionLabel == "" {
		return releasev1alpha1.Release{}, false, nil
	}

	release, err := FindRelease(ctx, ctrlReader, releaseVersionLabel)
	if err != nil {
		return releasev1alpha1.Release{}, false, microerror.Mask(err)
	}

	return release, true, nil
}

// FindRelease gets Release CR with the specified name.
func FindRelease(ctx context.Context, ctrlReader client.Reader, releaseVersion string) (releasev1alpha1.Release, error) {
	// Release CR always starts with a "v".
	if !strings.HasPrefix(releaseVersion, "v") {
		releaseVersion = fmt.Sprintf("v%s", releaseVersion)
	}

	// Retrieve the `Release` CR.
	release := releasev1alpha1.Release{}
	{
		err := ctrlReader.Get(ctx, client.ObjectKey{Name: releaseVersion}, &release)
		if apierrors.IsNotFound(err) {
			return releasev1alpha1.Release{}, microerror.Maskf(releaseNotFoundError, "Looking for Release %s but it was not found. Can't continue.", releaseVersion)
		} else if err != nil {
			return releasev1alpha1.Release{}, microerror.Mask(err)
		}
	}

	return release, nil
}

// GetComponentVersionsFromRelease gets all release components from the release with the specified name.
func GetComponentVersionsFromRelease(ctx context.Context, ctrlReader client.Reader, releaseVersion string) (map[string]string, error) {
	release, err := FindRelease(ctx, ctrlReader, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ret := GetComponentVersionsFromReleaseCR(release)
	return ret, nil
}

// GetComponentVersionsFromReleaseCR gets all release components from the specified Release CR.
func GetComponentVersionsFromReleaseCR(release releasev1alpha1.Release) map[string]string {
	componentVersions := map[string]string{}
	// Search the desired component.
	for _, component := range release.Spec.Components {
		componentVersions[component.Name] = component.Version
	}

	return componentVersions
}

// ContainsAzureOperator checks if the specified release contains azure-operator.
func ContainsAzureOperator(release releasev1alpha1.Release) bool {
	componentVersions := GetComponentVersionsFromReleaseCR(release)
	return componentVersions[azureOperatorComponentName] != ""
}

// IsLegacy checks if the specified release is a legacy release, i.e. a release without
// Cluster API controllers.
// If the release has azure-operator, it is considered to be a legacy release.
func IsLegacy(release releasev1alpha1.Release) bool {
	return ContainsAzureOperator(release)
}
