# Mutating webhooks

| Resource           | Field                                                 | Create                                                                              | Update                 | Delete |
|--------------------|-------------------------------------------------------|-------------------------------------------------------------------------------------|------------------------|--------|
| AzureCluster       | spec.controlPlaneEndpoint.host                        | ensure it is set if it was ""                                                       | n/a                    | n/a    |
|                    | spec.controlPlaneEndpoint.port                        | ensure it is set if it was 0                                                        | n/a                    | n/a    |
|                    | spec.location                                         | set it to the control plane region if it was ""                                     | n/a                    | n/a    |
| AzureConfig        | n/a                                                   | n/a                                                                                 | n/a                    | n/a    |
| AzureClusterConfig | n/a                                                   | n/a                                                                                 | n/a                    | n/a    |
| AzureMachine       | spec.location                                         | set it to the control plane region if it was ""                                     | n/a                    | n/a    |
| AzureMachinePool   | spec.location                                         | set it to the control plane region if it was ""                                     | n/a                    | n/a    |
|                    | spec.template.osDisk.managedDisk.storageAccountType   | if empty, set to Premium_LRS or Standard_LRS based on the VM type support           | n/a                    | n/a    |
|                    | spec.template.dataDisks                               | if empty, set it to the default disk setup (two 100Gb disks for kubelet and docker) | n/a                    | n/a    |
|                    | metadata.labels[release.giantswarm.io/version]        | if not set, it copies it from the Cluster CR.                                       | n/a                    | n/a    |
|                    | metadata.labels[azure-operator.giantswarm.io/version] | if not set, it copies it from the Cluster CR.                                       | n/a                    | n/a    |
| Cluster            | spec.clusterNetwork                                   | ensure it is set if it was nil                                                      | n/a                    | n/a    |
|                    | spec.controlPlaneEndpoint.host                        | ensure it is set if it was ""                                                       | n/a                    | n/a    |
|                    | spec.controlPlaneEndpoint.port                        | ensure it is set if it was 0                                                        | n/a                    | n/a    |
| MachinePool        | spec.replicas                                         | set to 1 if set to nil                                                              | set to 1 if set to nil | n/a    |
|                    | metadata.labels[release.giantswarm.io/version]        | if not set, it copies it from the Cluster CR.                                       | n/a                    | n/a    |
|                    | metadata.labels[azure-operator.giantswarm.io/version] | if not set, it copies it from the Cluster CR.                                       | n/a                    | n/a    |
| Spark              | metadata.labels[release.giantswarm.io/version]        | if not set, it copies it from the Cluster CR.                                       | n/a                    | n/a    |
