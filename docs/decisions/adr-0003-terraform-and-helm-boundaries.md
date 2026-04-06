# ADR-0003: Separate Terraform and Helm Responsibilities

## Status

Accepted

## Context

The platform targets GCP and GKE and needs a clean separation between infrastructure provisioning and workload deployment.

## Decision

Terraform provisions cloud infrastructure such as GKE, Artifact Registry, networking, PostgreSQL and Secret Manager. Helm packages and deploys Kubernetes workloads and runtime configuration.

## Consequences

- Infrastructure lifecycle and application lifecycle remain independently manageable.
- Environment composition becomes easier to explain in interviews and maintain in Git.
- CI/CD can apply infrastructure and workloads with different approval and rollout policies.
