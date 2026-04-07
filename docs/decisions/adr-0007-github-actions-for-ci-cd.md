# ADR-0007: Use GitHub Actions for CI/CD

## Status

Accepted

## Context

The project should be presentable as a GitHub portfolio repository and demonstrate a realistic delivery path from code validation to container image publishing and environment deployment.

The workflows must remain understandable and not depend on long-lived cloud keys.

## Decision

Use GitHub Actions for CI/CD.

CI validates Go code, tests, Docker builds, Helm chart rendering, Terraform formatting/validation and repository secret hygiene. Separate workflows build and push images, deploy dev and gate prod through a manual environment approval model.

Authentication to GCP uses Workload Identity Federation rather than service account JSON keys.

## Consequences

- Pull requests receive fast feedback across application and platform layers.
- Artifact Registry images are produced by a dedicated workflow.
- Dev deployment can be run manually for demonstration and smoke testing.
- Production deployment requires explicit confirmation and should be protected by GitHub Environments.
