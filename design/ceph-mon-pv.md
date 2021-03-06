# Ceph monitor PV storage

**Target version**: Rook 1.1

## Overview

Currently all of the storage for Ceph monitors (data, logs, etc..) is provided
using HostPath volume mounts. Supporting PV storage for Ceph monitors in
environments with dynamically provisioned volumes (AWS, GCE, etc...) will allow
monitors to migrate without requiring the monitor state to be rebuilt, and
avoids the operational complexity of dealing with HostPath storage.

The general approach taken in this design document is to augment the CRD with a
persistent volume claim template describing the storage requirements of
monitors. This template is used by Rook to dynamically create a volume claim for
each monitor.

## Monitor storage specification

The monitor specification in the CRD is updated to include a persistent volume
claim template that is used to generate PVCs for monitor database storage.

```go
type MonSpec struct {
	Count                int  `json:"count"`
	PreferredCount       int  `json:"preferredCount"`
	AllowMultiplePerNode bool `json:"allowMultiplePerNode"`
	VolumeClaimTemplate  *v1.PersistentVolumeClaim
}
```

The `VolumeClaimTemplate` is used by Rook to create PVCs for monitor storage.
The current set of template fields used by Rook when creating PVCs are
`StorageClassName` and `Resources`. Rook follows the standard convention of
using the default storage class when one is not specified in a volume claim
template. If the storage resource requirements are not specified in the claim
template, then Rook will use a default value. This is possible because unlike
the storage requirements of OSDs (xf: StorageClassDeviceSets), reasonable
defaults (e.g. 5-10 GB) exist for monitor daemon storage needs.

*Logs and crash data*. The current implementation continues the use of a
HostPath volume based on `dataDirHostPath` for storing daemon log and crash
data. This is a temporary exception that will be resolved as we converge on an
approach that works for all Ceph daemon types.

Finally, the entire volume claim template may be left unspecified in the CRD
in which case the existing HostPath mechanism is used for all monitor storage.

## Upgrades and CRD changes

When a new monitor is created it uses the _current_ storage specification found
in the CRD. Once a monitor has been created, it's backing storage is not
changed. This makes upgrades particularly simple because existing monitors
continue to use the same storage.

Once a volume claim template is defined in the CRD new monitors will be created
with PVC storage. In order to remove old monitors based on HostPath storage
first define a volume claim template in the CRD and then fail over each monitor.

## Clean up

Like `StatefulSets` removal of an monitor deployment does not automatically
remove the underlying PVC. This is a safety mechanism so that the data is not
automatically destroyed. The PVCs can be removed manually once the cluster is
healthy.

## Requirements and configuration

Rook currently makes explicit scheduling decisions for monitors by using node
selectors to force monitor node affinity. This means that the volume binding
should not occur until the pod is scheduled onto a specific node in the cluster.
This should be done by using the `WaitForFirstConsumer` binding policy on the
storage class used to provide PVs to monitors:

```
kind: StorageClass
volumeBindingMode: WaitForFirstConsumer
```

When using existing HostPath storage or non-local PVs that can migrate (e.g.
network volumes like RBD or EBS) existing monitor scheduling will continue to
work as expected. However, because monitors are scheduled without considering
the set of available PVs, when using statically provisioned local volumes Rook
expects volumes to be available. Therefore, when using locally provisioned
volumes take care to ensure that each node has storage provisioned.

Note that these limitations are currently imposed because of the explicit
scheduling implementation in Rook. These restrictions will be removed or
significantly relaxed once monitor scheduling is moved under control of Kubernetes
itself (ongoing work).

## Migration

Monitor migration is currently not implemented. When nodes are decommissioned,
for example, any monitors on the nodes are migrated using a fail over mechanism
rather than also migrating the PVC. This support is currently being worked on in
conjunction with the enhancements to scheduling.
