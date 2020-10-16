# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/azure-admission-controller/compare/v1.7.0...HEAD
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
