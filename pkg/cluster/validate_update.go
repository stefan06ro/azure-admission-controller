package cluster

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
	"github.com/giantswarm/azure-admission-controller/internal/semverhelper"
	"github.com/giantswarm/azure-admission-controller/pkg/generic"
	"github.com/giantswarm/azure-admission-controller/pkg/validator"
)

type UpdateValidator struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

type UpdateValidatorConfig struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

func NewUpdateValidator(config UpdateValidatorConfig) (*UpdateValidator, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	v := &UpdateValidator{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return v, nil
}

func (a *UpdateValidator) Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error {
	clusterNewCR := &capiv1alpha3.Cluster{}
	clusterOldCR := &capiv1alpha3.Cluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, clusterNewCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, clusterOldCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse Cluster CR: %v", err)
	}

	err := generic.ValidateOrganizationLabelUnchanged(clusterOldCR, clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateClusterNetworkUnchanged(*clusterOldCR, *clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpointUnchanged(*clusterOldCR, *clusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	oldClusterVersion, err := semverhelper.GetSemverFromLabels(clusterOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(clusterNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func validateClusterNetworkUnchanged(old capiv1alpha3.Cluster, new capiv1alpha3.Cluster) error {
	// Was nil and stayed nil. Not good but not changed so ok from this validator point of view.
	if old.Spec.ClusterNetwork == nil && new.Spec.ClusterNetwork == nil {
		return nil
	}

	// Was nil or became nil.
	if old.Spec.ClusterNetwork == nil && new.Spec.ClusterNetwork != nil ||
		old.Spec.ClusterNetwork != nil && new.Spec.ClusterNetwork == nil {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Check APIServerPort and ServiceDomain is unchanged.
	if *old.Spec.ClusterNetwork.APIServerPort != *new.Spec.ClusterNetwork.APIServerPort ||
		old.Spec.ClusterNetwork.ServiceDomain != new.Spec.ClusterNetwork.ServiceDomain {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Was nil and stayed nil. Not good but not changed so ok from this validator point of view.
	if old.Spec.ClusterNetwork.Services == nil && new.Spec.ClusterNetwork.Services == nil {
		return nil
	}

	// Check Services have not blanked out.
	if old.Spec.ClusterNetwork.Services == nil && new.Spec.ClusterNetwork.Services != nil ||
		old.Spec.ClusterNetwork.Services != nil && new.Spec.ClusterNetwork.Services == nil {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	// Check Services didn't change.
	if !reflect.DeepEqual(*old.Spec.ClusterNetwork.Services, *new.Spec.ClusterNetwork.Services) {
		return microerror.Maskf(clusterNetworkWasChangedError, "ClusterNetwork can't be changed.")
	}

	return nil
}
