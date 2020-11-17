# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Update `apptest` library and CAPZ fork.

## [1.13.3] - 2020-11-17

- Default `Cluster.Spec.ClusterNetwork.ServiceDomain` to `cluster.local` and don't allow any other value to be set. 

## [1.13.2] - 2020-11-13

## [1.13.1] - 2020-11-13

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

[Unreleased]: https://github.com/giantswarm/azure-admission-controller/compare/v1.13.3...HEAD
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
