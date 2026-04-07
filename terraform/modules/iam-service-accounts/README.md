# IAM Service Accounts Module

Creates Google service accounts, project-level role bindings and optional GKE Workload Identity bindings.

## Intended Usage

- one node-pool service account for GKE nodes
- one workload service account per application service
- stable mapping between Kubernetes service account names and Google service accounts

The module keeps IAM policy composition close to the environment stack, where the platform team can review least-privilege decisions alongside the workload architecture.

When the GKE cluster is created in the same Terraform stack, set `create_workload_identity_bindings = false` and create the Workload Identity IAM members after the cluster exists. This avoids applying bindings before the `${project_id}.svc.id.goog` workload pool is available.
