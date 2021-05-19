package generic

import (
	"context"

	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/normalize"
)

func ValidateOrganizationLabelUnchanged(old, new metav1.Object) error {
	if _, exists := old.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(organizationLabelNotFoundError, "meta CR doesn't contain Organization label %#q", label.Organization)
	}

	if _, exists := new.GetLabels()[label.Organization]; !exists {
		return microerror.Maskf(organizationLabelNotFoundError, "patch CR doesn't contain Organization label %#q", label.Organization)
	}

	if old.GetLabels()[label.Organization] != new.GetLabels()[label.Organization] {
		return microerror.Maskf(organizationLabelWasChangedError, "Organization label %#q must not be changed", label.Organization)
	}

	return nil
}

func ValidateOrganizationLabelContainsExistingOrganization(ctx context.Context, ctrlClient client.Client, obj metav1.Object) error {
	organizationName, ok := obj.GetLabels()[label.Organization]
	if !ok {
		return microerror.Maskf(organizationLabelNotFoundError, "CR doesn't contain Organization label %#q", label.Organization)
	}

	organization := &securityv1alpha1.Organization{}
	err := ctrlClient.Get(ctx, client.ObjectKey{Name: normalize.AsDNSLabelName(organizationName)}, organization)
	if apierrors.IsNotFound(err) {
		return microerror.Maskf(organizationNotFoundError, "Organization label %#q must contain an existing organization, got %#q but didn't find any CR with name %#q", label.Organization, organizationName, normalize.AsDNSLabelName(organizationName))
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func ValidateOrganizationLabelMatchesCluster(ctx context.Context, ctrlClient client.Client, obj metav1.Object) error {
	organizationName, ok := obj.GetLabels()[label.Organization]
	if !ok {
		return microerror.Maskf(organizationLabelNotFoundError, "CR doesn't contain Organization label %#q", label.Organization)
	}

	clusterName, ok := obj.GetLabels()[label.Cluster]
	if !ok {
		return microerror.Maskf(clusterLabelNotFoundError, "CR doesn't contain Cluster label %#q", label.Cluster)
	}

	cluster := capi.Cluster{}
	{
		clusters := &capi.ClusterList{}
		err := ctrlClient.List(ctx, clusters, client.MatchingLabels{label.Cluster: clusterName})
		if err != nil {
			return microerror.Mask(err)
		}

		// We want exactly one result.
		if len(clusters.Items) != 1 {
			return microerror.Maskf(clusterNotFoundError, "Expected one Cluster CR with label %#q=%#q. %d found.", label.Cluster, clusterName, len(clusters.Items))
		}

		cluster = clusters.Items[0]
	}

	clusterOrg, ok := cluster.GetLabels()[label.Organization]
	if !ok {
		return microerror.Maskf(organizationLabelNotFoundError, "Cluster CR doesn't contain Organization label %#q", label.Organization)
	}

	if clusterOrg != organizationName {
		return microerror.Maskf(nodepoolOrgDoesNotMatchClusterOrgError, "Organization label %#q (%#q) does not match Cluster's (%#q)", label.Organization, organizationName, clusterOrg)
	}

	return nil
}
