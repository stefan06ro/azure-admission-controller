package filter

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/release"
)

// IsObjectReconciledByLegacyRelease checks if the object is reconciled by an operator which is the
// part of a legacy Giant Swarm release (a release that does not have Cluster API controllers).
func IsObjectReconciledByLegacyRelease(ctx context.Context, logger micrologger.Logger, ctrlReader client.Reader, object metav1.ObjectMetaAccessor, ownerClusterGetter generic.OwnerClusterGetter) (bool, error) {
	objectName := fmt.Sprintf("%s/%s", object.GetObjectMeta().GetNamespace(), object.GetObjectMeta().GetName())
	var releaseVersionLabel string
	if object.GetObjectMeta().GetAnnotations() != nil && object.GetObjectMeta().GetAnnotations()[label.ReleaseVersion] != "" {
		releaseVersionLabel = object.GetObjectMeta().GetAnnotations()[label.ReleaseVersion]
	}

	// Try to get release from the CR.
	releaseCR, ok, err := release.TryFindReleaseForObject(ctx, ctrlReader, object, ownerClusterGetter)
	if release.IsReleaseNotFoundError(err) {
		logger.Debugf(ctx, "Object %s not reconciled by a legacy release (Release CR %s not found).", objectName, releaseVersionLabel)
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	if !ok {
		// Release not found for the object.
		logger.Debugf(ctx, "Object %s not reconciled by a legacy release (cannot determine the release).", objectName)
		return false, nil
	}

	// Now when we have release CR, let's check if this is a legacy release.
	isLegacy := release.IsLegacy(releaseCR)
	if isLegacy {
		logger.Debugf(ctx, "Object %s is reconciled by a legacy release %s.", objectName, releaseVersionLabel)
	} else {
		logger.Debugf(ctx, "Object %s not reconciled by a legacy release %s.", objectName, releaseVersionLabel)
	}

	return isLegacy, nil
}
