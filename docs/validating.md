# Validating webhooks

| Resource           | Field                                               | Create                                                    | Update                                                | Delete |
|--------------------|-----------------------------------------------------|-----------------------------------------------------------|-------------------------------------------------------|--------|
| AzureCluster       | metadata.labels[giantswarm.io/organization]         | Check it is a valid organization name                     | Check it is unchanged                                 | n/a    |
|                    | spec.controlPlaneEndpoint.host                      | Check it is "api.<cluster ID>.<installation base domain>" | Check it is unchanged                                 | n/a    |
|                    | spec.controlPlaneEndpoint.host                      | Check it is 443                                           | Check it is unchanged                                 | n/a    |
|                    | spec.location                                       | Check it matches the installation's location              | Check it is unchanged                                 | n/a    |
| AzureMachine       | metadata.labels[giantswarm.io/organization]         | Check it is a valid organization name                     | Check it is unchanged                                 | n/a    |
|                    | metadata.labels[release.giantswarm.io/version]      | n/a                                                       | Check upgrade is allowed                              | n/a    |
|                    | spec.failureDomain                                  | Check it is supported by the VM type in the region        | Check it is unchanged                                 | n/a    |
|                    | spec.location                                       | Check it matches the installation's location              | Check it is unchanged                                 | n/a    |
|                    | spec.sshPublicKey                                   | Check that the field is empty                             | Check that the field is empty                         | n/a    |
| AzureMachinePool   | metadata.labels[giantswarm.io/organization]         | Check it is a valid organization name                     | Check it is unchanged                                 | n/a    |
|                    | spec.location                                       | Check it matches the installation's location              | Check it is unchanged                                 | n/a    |
|                    | spec.template.acceleratedNetworking                 | If enabled, checks it is supported by the VM type.        | Check it is unchanged                                 | n/a    |
|                    | spec.template.dataDisks                             | Check they are "docker" (100Gb) and "kubelet" (100Gb)     | Check they are "docker" (100Gb) and "kubelet" (100Gb) | n/a    |
|                    | spec.template.osDisk.managedDisk.storageAccountType | Check it is supported by the VM type.                     | Check it is unchanged                                 | n/a    |
|                    | spec.template.sshPublicKey                          | Check that the field is empty                             | Check that the field is empty                         | n/a    |
|                    | spec.template.vmSize                                | Check it is a valid VM type and it is big enough          | Check it is a valid VM type and it is big enough      | n/a    |
| AzureConfig        | metadata.labels[release.giantswarm.io/version]      | n/a                                                       | Check upgrade is allowed                              | n/a    |
| AzureClusterConfig | metadata.labels[release.giantswarm.io/version]      | n/a                                                       | Check upgrade is allowed                              | n/a    |
| Cluster            | metadata.labels[giantswarm.io/organization]         | Check it is a valid organization name                     | Check it is unchanged                                 | n/a    |
|                    | spec.clusterNetwork                                 | Check it is not nil                                       | Check it is unchanged                                 | n/a    |
|                    | spec.clusterNetwork.APIServerPort                   | Check it is 443                                           | Check it is unchanged                                 | n/a    |
|                    | spec.clusterNetwork.serviceDomain                   | Check it is "<cluster ID>.<installation base domain>"     | Check it is unchanged                                 | n/a    |
|                    | spec.clusterNetwork.services                        | Check it is not nil                                       | Check it is unchanged                                 | n/a    |
|                    | spec.clusterNetwork.services.cidrBlocks             | Check it is set to ["172.31.0.0/16"]                      | Check it is unchanged                                 | n/a    |
|                    | spec.controlPlaneEndpoint.host                      | Check it is "api.<cluster ID>.<installation base domain>" | Check it is unchanged                                 | n/a    |
|                    | spec.controlPlaneEndpoint.host                      | Check it is 443                                           | Check it is unchanged                                 | n/a    |
| MachinePool        | metadata.labels[giantswarm.io/organization]         | Check it is a valid organization name                     | Check it is unchanged                                 | n/a    |
|                    | spec.failureDomains                                 | Check they are valid and supported by the VM type.        | Check they are unchanged                              | n/a    |
| Spark              | n/a                                                 | n/a                                                       | n/a                                                   | n/a    |
