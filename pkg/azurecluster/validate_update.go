package azurecluster

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capiutil "sigs.k8s.io/cluster-api/util"
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
	azureClusterNewCR := &capz.AzureCluster{}
	azureClusterOldCR := &capz.AzureCluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureClusterNewCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureClusterOldCR); err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse AzureCluster CR: %v", err)
	}

	if !azureClusterNewCR.GetDeletionTimestamp().IsZero() {
		a.logger.LogCtx(ctx, "level", "debug", "message", "The object is being deleted so we don't validate it")
		return nil
	}

	capi, err := generic.IsCAPIRelease(azureClusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}
	if capi {
		return nil
	}

	err = azureClusterNewCR.ValidateUpdate(azureClusterOldCR)
	err = errors.IgnoreCAPIErrorForField("metadata.Name", err)
	err = errors.IgnoreCAPIErrorForField("spec.networkSpec.subnets", err)
	// TODO(axbarsan): Remove this once all the older clusters have it.
	err = errors.IgnoreCAPIErrorForField("spec.networkSpec.apiServerLB", err)
	err = errors.IgnoreCAPIErrorForField("spec.SubscriptionID", err)
	if err != nil {
		return microerror.Mask(err)
	}

	err = generic.ValidateOrganizationLabelUnchanged(azureClusterOldCR, azureClusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	err = validateControlPlaneEndpointUnchanged(*azureClusterOldCR, *azureClusterNewCR)
	if err != nil {
		return microerror.Mask(err)
	}

	return a.validateRelease(ctx, azureClusterOldCR, azureClusterNewCR)
}

func (a *UpdateValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func (a *UpdateValidator) validateRelease(ctx context.Context, azureClusterOldCR *capz.AzureCluster, azureClusterNewCR *capz.AzureCluster) error {
	oldClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterOldCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from the AzureCluster being updated")
	}
	newClusterVersion, err := semverhelper.GetSemverFromLabels(azureClusterNewCR.Labels)
	if err != nil {
		return microerror.Maskf(errors.ParsingFailedError, "unable to parse version from applied AzureCluster")
	}

	if !newClusterVersion.Equals(oldClusterVersion) {
		cluster, err := capiutil.GetOwnerCluster(ctx, a.ctrlClient, azureClusterNewCR.ObjectMeta)
		if err != nil {
			return microerror.Mask(err)
		}

		clusterCRReleaseVersion, err := semverhelper.GetSemverFromLabels(cluster.Labels)
		if err != nil {
			return microerror.Mask(err)
		}

		if !newClusterVersion.Equals(clusterCRReleaseVersion) {
			return microerror.Maskf(errors.InvalidOperationError, "AzureCluster release version must be set to the same release version as Cluster CR release version label")
		}
	}

	return releaseversion.Validate(ctx, a.ctrlClient, oldClusterVersion, newClusterVersion)
}
