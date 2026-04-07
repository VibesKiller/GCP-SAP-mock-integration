# GKE Cluster Module

Creates a regional standard GKE cluster and a single autoscaled node pool.

## Design Choices

- Workload Identity enabled by default
- private nodes with Cloud NAT expected upstream
- default node pool removed
- transient default node pool configuration mirrors the managed node pool to avoid hidden quota spikes during cluster creation
- autoscaled primary node pool
- basic hardening through shielded nodes and secure metadata mode

## Inputs Worth Highlighting

- `node_service_account`: Google service account used by GKE nodes.
- `master_authorized_networks`: optional control-plane allowlist.
- `node_locations`: optional zone list for the node pool. Useful for dev environments with tight regional quotas.
- `disk_type` and `disk_size_gb`: node boot disk profile. Dev can use smaller or standard disks while prod can keep balanced SSD-backed disks.
- `labels`: propagated to the cluster and node pool.
