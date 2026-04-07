# ADR-0003: Use Terraform for GCP Infrastructure

## Status

Accepted

## Context

The project needs reproducible cloud infrastructure for GKE, networking, Cloud SQL, Artifact Registry, IAM, Secret Manager and Managed Kafka.

Manual provisioning would make the platform difficult to review, repeat or discuss credibly in an enterprise setting.

## Decision

Use Terraform for Google Cloud infrastructure provisioning.

The repository separates reusable modules under `terraform/modules/` from environment composition under `terraform/envs/dev/` and `terraform/envs/prod/`.

## Consequences

- Infrastructure design is visible, reviewable and reusable.
- Dev and prod can diverge intentionally through variables instead of copy-pasted resources.
- Terraform state must be stored remotely and protected.
- Some operational cleanup behaviors, such as Cloud SQL and Service Networking teardown, require environment-specific deletion policies.
