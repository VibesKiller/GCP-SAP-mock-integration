# Terraform Dev Environment

The dev stack keeps the same architectural boundaries as production but uses lighter sizing and lower protection defaults.

## What It Provisions

- custom VPC with one application subnet, flow logs and Cloud NAT
- private service access for Cloud SQL
- regional standard GKE cluster with Workload Identity
- Cloud SQL for PostgreSQL with private IP
- Artifact Registry Docker repository
- Google service accounts for nodes and workloads
- Secret Manager secrets for the generated database password and SAP token placeholder
- Managed Service for Apache Kafka cluster, platform topics and ACLs

## Suggested Workflow

```bash
cp terraform/envs/dev/backend.hcl.example terraform/envs/dev/backend.hcl
cp terraform/envs/dev/terraform.tfvars.example terraform/envs/dev/terraform.tfvars
terraform -chdir=terraform/envs/dev init -backend-config=backend.hcl
terraform -chdir=terraform/envs/dev plan -var-file=terraform.tfvars
terraform -chdir=terraform/envs/dev apply -var-file=terraform.tfvars
```

## Notes

- Dev still uses private networking to stay structurally close to production.
- Deletion protection is disabled to keep iteration practical.
- Kafka topics are provisioned from the repository catalog rather than being auto-created by clients.
- The GKE control-plane allowlist auto-detects the current workstation public IPv4 address during `terraform plan/apply`. Add stable VPN or office CIDRs to `master_authorized_networks` only when needed.
- The dev IP plan keeps `10.30.0.0/16` reserved for Private Service Access and uses `10.50.0.0/20` for GKE Services to avoid overlap during partial apply/retry workflows.
