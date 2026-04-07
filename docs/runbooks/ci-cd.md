# Runbook: CI/CD Operations

## Workflows

- `ci`: validates Go, Docker build smoke, Helm rendering, Terraform validation and repository secret hygiene.
- `docker images`: builds and pushes service images to Artifact Registry.
- `deploy dev`: manually deploys dev with optional Terraform apply and smoke test.
- `deploy prod`: manually deploys prod after explicit confirmation and GitHub Environment approval.
- `terraform`: lightweight Terraform formatting workflow for infrastructure-only changes.

## Required GitHub Variables

- `GCP_PROJECT_ID`
- `GCP_REGION`
- `ARTIFACT_REPOSITORY`
- `TF_STATE_BUCKET`

## Required GitHub Secrets

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_CICD_SERVICE_ACCOUNT`

No service account JSON key should be committed or configured.

## Dev Deployment

Use the `deploy dev` workflow with `apply_infra=true` when infrastructure should converge before workloads are deployed.

Use `apply_infra=false` for a faster workload-only deployment after Terraform has already been applied.

## Production Deployment

Use the `deploy prod` workflow and type `deploy-prod` in the confirmation input. Protect the `production` GitHub Environment with required reviewers.

## Failure Handling

- If Terraform validation fails, run `make terraform-validate` locally.
- If Helm rendering fails, run `make helm-template-dev` or `make helm-template-prod` locally.
- If smoke test fails, inspect `ingestion-api`, `event-processor` and `query-api` logs in the target namespace.
