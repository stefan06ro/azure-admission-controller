# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.0.0] - 2021-08-03

### Added

- New `pkg/filter` package with function which checks if the CR belongs to a cluster from a legacy non-CAPI release.
  Release is considered to be "legacy" if it contains azure-operator.
- Unit tests for functions from `release` package.
- `HttpHandlerFactory` for creating HTTP handlers that are using new webhook handlers.
- Cluster webhook handler that replaces mutators and validators.
- AzureCluster webhook handler that replaces mutators and validators.
- MachinePool webhook handler that replaces mutators and validators.
- AzureMachinePool webhook handler that replaces mutators and validators.
- AzureMachine webhook handler that replaces mutators and validators.
- Spark webhook handler that replaces mutator.
- AzureClusterConfig webhook handler that replaces validator.
- AzureConfig webhook handler that replaces validator.
- CR filtering and webhook handler tests that use real CRs from real management clusters as test cases.

### Changed

- Use caching client for `Releases`.
- Upgrade `apiextensions/v2` -> `apiextensions/v3`.
- Upgrade `k8sclient/v4` -> `k8sclient/v5`.
- When importing `sigs.k8s.io/cluster-api/api/v1alpha3` use `capi` as package alias.
- When importing `sigs.k8s.io/cluster-api/exp/api/v1alpha3` use `capiexp` as package alias.
- When importing `sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3` use `capz` as package alias.
- When importing `sigs.k8s.io/cluster-api-provider-azure/exp/api/v1alpha3` use `capzexp` as package alias.
- Remove package names from some file names.
- Move labels mutator functions to `pkg/mutator`.
- Add new `WebhookHandler` interfaces for validation and mutation.
- Prepare helm values to configuration management.
- Update architect-orb to v4.0.0.
- All webhook handlers (previously mutators and validators) are now created in `pkg/app` package.
- The webhook handlers are not added to HTTP handler manually. Now for all handlers we check which wehbook handler
  interfaces are implementing, and according to that we add appropriate HTTP handlers.
- Resource SKU stub API has been moved to `pkg/unittests` since it was being used in multiple places.

## [2.7.0] - 2021-05-19

### Fixed

- Add missing config annotation to Helm Chart.

## [2.6.0] - 2021-05-19

### Changed

- Revert changes for new config system.

## [2.5.0] - 2021-05-19

### Changed

- Use new config system.

## [2.4.1] - 2021-05-14

### Fixed

- Include `AzureConfig`, `AzureClusterConfig` and `AzureMachine` in objects not validated if they are being deleted.

## [2.4.0] - 2021-05-10

### Changed

- Don't validate objects if they are being deleted.

### Added

- Skipping all validation and defaulting for resources with CAPI release label

## [2.3.2] - 2021-02-23

### Changed

- Skipping validation for the `azureCluster.spec.networkSpec.apiServerLB` on update.

## [2.3.1] - 2021-02-23

### Changed

- Changed `azureCluster.spec.networkSpec.apiServerLB` defaulting to include the case where the field does not exist.

## [2.3.0] - 2021-02-23

### Added

- Execute CAPI/CAPZ validation for all resources.
- Execute CAPI/CAPZ defaulting on all resources.

### Changed

- Allow `0` as the minimum node count for the cluster autoscaler.

### Remove

- Don't execute CAPI/CAPZ validation for the `subnet` and `spec.subscriptionID` fields of the `AzureCluster` resource.
- Remove defaulting for the `AzureCluster` `control-plane` subnet.

## [2.2.0] - 2021-02-05

### Fixed

- Add `Cluster` and `AzureCluster` mutate webhook definition in the Helm chart.
- Ensure `cluster-operator.giantswarm.io/version` label has the right value depending on the `release.giantswarm.io/version`
  label when updating `Cluster` and `AzureCluster`.

## [2.1.0] - 2021-02-03

### Added

- Prevent changes to AzureMachinePool Spot VM configuration after creation.

## [2.0.1] - 2021-01-27

### Fixed

- Avoid nil pointer panic while checking for failureDomain.

## [2.0.0] - 2021-01-20

### Changed

- Ensure `azure-operator.giantswarm.io/version` label has the right value depending on the `release.giantswarm.io/version`
  label when updating `Cluster` and `AzureCluster`.

## [1.18.0] - 2021-01-15

### Added

- Ensure autoscaler min and max annotations are present when creating or updating a `Machinepool`.

## [1.17.0] - 2021-01-14

### Changed

- Update cert apiVersion to v1.
- Ignore Deprecated releases during upgrade validations.

## [1.16.0] - 2020-12-15

### Added

- Block possibility to change `Spec/Azure/AvailabilityZones` field of `AzureConfig`.

## [0.15.0] - 2020-12-10

### Added

- Validate the Master node CIDR does not change in `AzureConfig` CR.

## [1.14.0] - 2020-12-04

### Added

- Ignore release in validation logic by setting `release.giantswarm.io/ignore` annotation on a `Release` CR.
- Validate Organization label for `AzureMachinePool` and `MachinePool` match `Cluster`'s.
- Make Pod terminate itself when the TLS certificate is expired.

### Changed

- Update `apptest` library and CAPZ fork dependencies.
- On upgrade ignore alpha releases when validating new cluster release version, because upgrading to or from an alpha release is not supported.

## [1.13.3] - 2020-11-17

### Changed

- Default `Cluster.Spec.ClusterNetwork.ServiceDomain` to `cluster.local` and don't allow any other value to be set.

## [1.13.2] - 2020-11-13

- No changes.

## [1.13.1] - 2020-11-13

- No changes.

## [1.13.0] - 2020-11-12

### Changed

- Simplify `Validator` interface to only return `error`, dropping the `bool`.
- Use specific errors for specific business rules.

### Added

- Validate that the `Organization` label contains an existing `Organization`.
- Set default value for `MachinePool.Spec.Replicas` to 1.
- Set `AzureMachine`'s, `AzureCluster`'s, and `AzureMachinePool`'s `location` field on create if empty.
- Validate `AzureMachine`'s, `AzureCluster`'s, and `AzureMachinePool`'s `location` matches the installation's `location`.
- Validate `AzureMachine`'s, `AzureCluster`'s, and `AzureMachinePool`'s `location` never changes.
- Validate `FailureDomain` for `AzureMachine` is a valid and supported one.
- Validate `FailureDomain` for `AzureMachine` never changes.
- Set `release.giantswarm.io/version` label on `MachinePool`, `AzureMachinePool`, and `Spark` CRs on create if empty.
- Set `AzureMachinePool`'s and `MachinePool`'s operators versions on create if missing.
- Add builders to make it easier to write tests.
- New value for `AzureCluster` `release.giantswarm.io/version` label must match the same label on `Cluster` CR
- `Cluster` `Creating` condition - setting `Status=Unknown` is not allowed
- `Cluster` `Creating` condition - new `Status` value must be either `True` or `False`
- `Cluster` `Creating` condition - removing existing condition is not allowed
- `Cluster` `Creating` condition - changing `Status` from `False` to `True` is not allowed
- `Cluster` `Upgrading` condition - setting `Status=Unknown` is not allowed
- `Cluster` `Upgrading` condition - new `Status` value must be either `True` or `False`
- `Cluster` `Upgrading` condition - removing existing condition is not allowed

## [1.12.0] - 2020-10-27

### Added

- Prevent Organization label value change on CR updates.

### Removed

- Removed Organization label value normalization on CR creation.

## [1.11.0] - 2020-10-23

### Added

- Ensure default value and immutability for `AzureCluster.ControlPlaneEndpoint`, `Cluster.ControlPlaneEndpoint` and `Cluster.ClusterNetwork fields`.

## [1.10.0] - 2020-10-23

### Added

- AzureCluster, AzureMachine, AzureMachinePool, Cluster and MachinePools CR's
  are ensured to have normalized form of giantswarm.io/organization label
  value via mutating webhook on CREATE.

### Changed

- Set `400` status code if a request is invalid.

## [1.9.1] - 2020-10-21

### Fixed

- Compare `FailureDomains` field manually when updating instead of relying on `reflect.DeepEqual` which may have issues when the slice is nil or empty.

## [1.9.0] - 2020-10-21

### Added

- Ensure failureDomains in MachinePool are supported by the AzureMachinePool VM type in the current location.

## [1.8.0] - 2020-10-20

### Added

- Block configuring the `DataDisks` field in AzureMachinePool CRs, and set a default value for it.

## [1.7.0] - 2020-10-16

### Added

- Check that SSH keys are not set in AzureMachine and AzureMachinePool CRs.
- Add mutating webhook to set storage account type in the AzureMachinePool CR if unset.

### Changed

- Block changing node pool instance type from one that supports premium storage to one that does not.

## [1.6.2] - 2020-10-07

### Fixed

- Add missing schemes to controller client.

## [1.6.1] - 2020-10-07

### Changed

- Validate parameters when building objects.
- Remove unnecesary k8sclients.

## [1.6.0] - 2020-10-07

### Added

- Added validating webhook for AzureNodePools that checks instance type is valid and meets minimum requirements.
- Added validating webhook for AzureNodePools that checks instance type supports accelerated networking if enabled.

### Changed

- Updated to Go 1.15.

## [1.5.0] - 2020-08-19

### Changed

- Allow skipping patches.

## [1.4.0] - 2020-08-05

### Removed

- AWS related controllers.

## [1.3.0] - 2020-07-23

### Changed

- When parsing the release version during Azure upgrades, we are now more tolerant when parsing the versions string so it works as well with leading `v` versions, like `v1.2.3`.

## [1.2.0] - 2020-07-20

### Added

- Validation Webhooks that check for valid upgrade paths for legacy Azure clusters.
- Added application to Azure app collection.

## [1.1.0] - 2020-07-16

### Added

- Handling of creation and updates to [`AWSMachineDeployment`](https://docs.giantswarm.io/reference/cp-k8s-api/awsmachinedeployments.infrastructure.giantswarm.io) (`awsmachinedeployments.infrastructure.giantswarm.io`) resources, with defaulting of the [`.spec.node_spec.aws.instanceDistribution.onDemandPercentageAboveBaseCapacity`](https://docs.giantswarm.io/reference/cp-k8s-api/awsmachinedeployments.infrastructure.giantswarm.io/#v1alpha2-.spec.provider.instanceDistribution.onDemandPercentageAboveBaseCapacity) attribute.

## [1.0.0] - 2020-06-15

- Several changes

## [0.1.0] - 2020-06-10

- First release.

[Unreleased]: https://github.com/giantswarm/azure-admission-controller/compare/v3.0.0...HEAD
[3.0.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.7.0...v3.0.0
[2.7.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.4.1...v2.5.0
[2.4.1]: https://github.com/giantswarm/azure-admission-controller/compare/v2.4.0...v2.4.1
[2.4.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.3.2...v2.4.0
[2.3.2]: https://github.com/giantswarm/azure-admission-controller/compare/v2.3.1...v2.3.2
[2.3.1]: https://github.com/giantswarm/azure-admission-controller/compare/v2.3.0...v2.3.1
[2.3.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/azure-admission-controller/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/giantswarm/azure-admission-controller/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.18.0...v2.0.0
[1.18.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.17.0...v1.18.0
[1.17.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.16.0...v1.17.0
[1.16.0]: https://github.com/giantswarm/azure-admission-controller/compare/v0.15.0...v1.16.0
[0.15.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.14.0...v0.15.0
[1.14.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.13.3...v1.14.0
[1.13.3]: https://github.com/giantswarm/azure-admission-controller/compare/v1.13.2...v1.13.3
[1.13.2]: https://github.com/giantswarm/azure-admission-controller/compare/v1.13.1...v1.13.2
[1.13.1]: https://github.com/giantswarm/azure-admission-controller/compare/v1.13.0...v1.13.1
[1.13.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.12.0...v1.13.0
[1.12.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.11.0...v1.12.0
[1.11.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.10.0...v1.11.0
[1.10.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.9.1...v1.10.0
[1.9.1]: https://github.com/giantswarm/azure-admission-controller/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.8.0...v1.9.0
[1.8.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.7.0...v1.8.0
[1.7.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.6.2...v1.7.0
[1.6.2]: https://github.com/giantswarm/azure-admission-controller/compare/v1.6.1...v1.6.2
[1.6.1]: https://github.com/giantswarm/azure-admission-controller/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/giantswarm/azure-admission-controller/compare/v1.0.0...v0.0.1
[0.0.1]: https://github.com/giantswarm/azure-admission-controller/releases/tag/v0.0.1
