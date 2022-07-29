---
sidebar_position: 5
sidebar_label: "CRD Object Class"
---

# HwameiStor's CRD Object Class

Hwameistor defined more object classes to associate PV/PVC with local disks.

|Kind|Abbr.|Function|
|--|--|--|
|LocalDiskNode|ldn|Register a node|
|LocalDisk|ld|Register data disks on nodes and automatically find which disks are available|
|LocalDiskClaim|ldc|Filter and register local data disks|
|LocalStorageNode|lsn|Automatically create a storage pool, i.e., a set of LVMs|
|LocalVolume|lv|Create LVMs and allocate them to PVs|
|LocalDiskExpand|lvexpand|Expand storage pools|


