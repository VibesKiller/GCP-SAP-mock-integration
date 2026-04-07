# Terraform on Google Cloud

This Terraform layer provisions the shared Google Cloud foundation for the SAP-style integration platform.

The design intentionally separates concerns:

- Terraform provisions platform infrastructure and cloud identities.
- Helm deploys application workloads into GKE.
- Kafka topic contracts remain versioned in the repository and are reused during provisioning.

## Module Layout

- `modules/network`: VPC, subnet, Cloud NAT and private service access for Cloud SQL.
- `modules/gke-cluster`: Standard regional GKE cluster and node pool with Workload Identity.
- `modules/postgresql`: Cloud SQL for PostgreSQL with private IP and baseline operational settings.
- `modules/artifact-registry`: Docker repository for application images.
- `modules/iam-service-accounts`: Google service accounts, project roles and Workload Identity bindings.
- `modules/secret-manager`: Secret containers, initial versions and least-privilege secret access.
- `modules/managed-kafka`: Google Managed Service for Apache Kafka cluster, topics and ACLs.

## Environment Layout

- `envs/dev`: lower-cost defaults, deletion protection disabled, smaller node pool.
- `envs/prod`: higher-capacity defaults, stronger protection settings, production-oriented sizing.

## Architecture Assumptions

- GKE uses private nodes and Workload Identity.
- Cloud SQL is exposed over private IP inside the VPC.
- GKE nodes reach the internet through Cloud NAT.
- Artifact Registry is shared per environment and stores the four application images.
- Kafka is provisioned with Google Managed Service for Apache Kafka.
- Kafka topics are derived from `platform/kafka/topic-catalog.yaml` to keep infrastructure aligned with the platform contract.
- Kafka bootstrap servers are retrieved after provisioning with `gcloud managed-kafka clusters describe`, because the current Terraform resource does not expose the bootstrap address as a computed attribute.

## Provisioning Order

1. Enable required Google APIs.
2. Create network, subnetwork, Cloud NAT and private service access.
3. Create Artifact Registry and Google service accounts.
4. Generate application secrets and store them in Secret Manager.
5. Provision Cloud SQL and Managed Kafka inside the VPC.
6. Provision GKE and bind Kubernetes service accounts to Google service accounts.
7. Deploy workloads with Helm using the Terraform outputs.

## Remote State

Use a GCS backend for each environment. Example backend config files are provided in each environment folder.

```hcl
bucket = "your-terraform-state-bucket"
prefix = "envs/dev"
```

Restrict access to the state bucket because it contains sensitive values, including generated database passwords in state.

## Commands

Dev:

```bash
terraform -chdir=terraform/envs/dev init -backend-config=backend.hcl
terraform -chdir=terraform/envs/dev plan -var-file=terraform.tfvars
terraform -chdir=terraform/envs/dev apply -var-file=terraform.tfvars
```

Prod:

```bash
terraform -chdir=terraform/envs/prod init -backend-config=backend.hcl
terraform -chdir=terraform/envs/prod plan -var-file=terraform.tfvars
terraform -chdir=terraform/envs/prod apply -var-file=terraform.tfvars
```

## Helm Integration

After `apply`, use these outputs as the source of truth for Helm values:

- Artifact Registry repository URL for image repositories.
- GKE cluster name and `gcloud container clusters get-credentials` command.
- Google service account emails for Workload Identity annotations.
- Secret Manager secret IDs for secret-sync patterns.
- Kafka cluster lookup command for bootstrap address retrieval.

## Kafka Notes

This repo stays intentionally Kafka-native:

- topic names follow `platform/policies/kafka-topic-naming.md`
- topics are provisioned explicitly, not auto-created
- ACLs are derived from topic owners and consumer-group ownership
- application service accounts receive `roles/managedkafka.client`
- workloads are expected to connect to the managed bootstrap address over the VPC
