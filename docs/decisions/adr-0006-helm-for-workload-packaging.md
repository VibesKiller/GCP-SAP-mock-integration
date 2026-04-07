# ADR-0006: Use Helm for Kubernetes Workload Packaging

## Status

Accepted

## Context

The platform deploys multiple services with shared conventions: labels, probes, resource limits, service accounts, Kafka configuration, database secret references and environment overlays.

Raw Kubernetes manifests would be repetitive and harder to keep consistent.

## Decision

Use Helm to package the application workloads.

The chart lives under `deploy/helm/platform/` and supports base, dev, prod and MicroK8s values.

## Consequences

- Runtime configuration is centralized and overrideable per environment.
- Terraform remains responsible for cloud resources while Helm remains responsible for Kubernetes workloads.
- Values must avoid hardcoded secrets and instead reference Kubernetes Secrets or cloud-derived runtime values.
- The chart can be linted and rendered in CI before deployment.
