# ADR-0005: Use GKE for the Kubernetes Runtime

## Status

Accepted

## Context

The cloud target is Google Cloud Platform, and the project needs a credible Kubernetes runtime for Helm deployments, workload identity, autoscaling and operational probes.

A local-only Kubernetes story would not demonstrate the platform engineering concerns expected for cloud roles.

## Decision

Use GKE as the production-like Kubernetes runtime on GCP.

The chart remains portable enough to run on MicroK8s for persistent local Kubernetes validation, but the cloud deployment target is GKE.

## Consequences

- Workloads can use Workload Identity instead of static service account keys.
- The platform can demonstrate probes, HPA, resource requests, secret references and environment overlays.
- Network access to the GKE control plane must be managed carefully through authorized networks.
- GKE quota and disk sizing are represented in Terraform variables for dev and prod.
