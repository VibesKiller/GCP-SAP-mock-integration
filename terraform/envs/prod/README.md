# Terraform Prod Environment

The prod stack mirrors the dev module composition but hardens the defaults and increases capacity.

## What Changes Compared With Dev

- deletion protection enabled
- larger GKE node pool limits
- regional Cloud SQL availability
- larger Kafka cluster capacity
- production labels for governance and operations

## Suggested Workflow

```bash
cp terraform/envs/prod/backend.hcl.example terraform/envs/prod/backend.hcl
cp terraform/envs/prod/terraform.tfvars.example terraform/envs/prod/terraform.tfvars
terraform -chdir=terraform/envs/prod init -backend-config=backend.hcl
terraform -chdir=terraform/envs/prod plan -var-file=terraform.tfvars
terraform -chdir=terraform/envs/prod apply -var-file=terraform.tfvars
```

## Operational Notes

- Keep remote state in a protected GCS bucket with versioning enabled.
- Limit access to production state and production service-account impersonation.
- Treat Helm values for Workload Identity annotations as derived from Terraform outputs, not hand-maintained drift.
