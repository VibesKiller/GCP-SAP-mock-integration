# GCP SAP Mock Integration Platform

A small implementation of a realistic enterprise technology stack.

This repository is intended to serve as a portfolio for DevOps, Cloud, and Platform Engineer interviews, and aims to demonstrate skills in designing, developing, integrating, and maintaining business processes, as well as the ability to manage a cloud-based system built with Infrastructure as Code (IaC).

## Purpose

This project simulates a SAP-like system publishing business events such as:

- `sales_order.created`
- `sales_order.updated`
- `customer.updated`
- `invoice.issued`

The platform receives SAP payloads, validates and normalizes them, publishes canonical events to Apache Kafka, processes them asynchronously, persists projections in PostgreSQL and exposes read-only APIs for querying customers, orders and invoices.

## Technology Stack

- Language: Go
- Event streaming: Apache Kafka
- Local runtime: Docker Compose
- Database: PostgreSQL locally, Cloud SQL for PostgreSQL on GCP
- Kubernetes: GKE for cloud, MicroK8s for local cluster testing
- Infrastructure as Code: Terraform
- Kubernetes packaging: Helm
- Container registry: Google Artifact Registry
- Secret management: Google Secret Manager and Kubernetes Secret references
- Observability: Prometheus and Grafana
- CI/CD: GitHub Actions
- API documentation: OpenAPI
- Logging: JSON structured logs with correlation IDs

## System Dependencies

Required for local development:

- `git`
- `make`
- `curl`
- `jq`
- Docker Engine with Docker Compose v2
- Go `1.24.x`

Required for Kubernetes and cloud workflows:

- Terraform `>= 1.7`
- Helm `3.x`
- `kubectl`
- Google Cloud CLI, including `gcloud`
- Docker authentication configured for Google Artifact Registry

Optional but useful:

- MicroK8s, for a persistent local Kubernetes cluster
- `yamllint`, for validating GitHub Actions workflow YAML
- `psql`, for manual PostgreSQL checks outside containers

Typical Ubuntu/Debian baseline packages:

```bash
sudo apt-get update
sudo apt-get install -y git make curl jq ca-certificates
```

Install Docker, Go, Terraform, Helm and the Google Cloud CLI from their official distribution channels, because those tools have versioned release streams and platform-specific installation steps.

## Implemented Workflow

The implemented event flow is:

```text
sap-mock-api
  -> ingestion-api
  -> Apache Kafka
  -> event-processor
  -> PostgreSQL / Cloud SQL
  -> query-api
```

Service responsibilities:

- `sap-mock-api` exposes endpoints that generate realistic SAP-style sample events.
- `ingestion-api` validates incoming payloads, creates the canonical event envelope and publishes to Kafka.
- `event-processor` consumes Kafka events, applies idempotency through `processed_events`, persists data and routes unrecoverable failures to the DLQ.
- `query-api` exposes read-only APIs for customers, orders and invoices.

Kafka contracts:

- Sales orders topic: `sap.sales-orders.v1`
- Customers topic: `sap.customers.v1`
- Invoices topic: `sap.invoices.v1`
- Dead-letter topic: `sap.integration.dlq.v1`
- Main consumer group: `sap-integration.event-processor.v1`

The local stack uses real Apache Kafka in Docker Compose. The GCP target is designed for Google Managed Service for Apache Kafka with TLS and Workload Identity based authentication.

## Run The Local Stack

Start the full local environment:

```bash
cp .env.example .env
make up
make smoke
```

This starts PostgreSQL, Apache Kafka, Kafka UI, Prometheus, Grafana and the Go services.

Useful local commands:

```bash
make logs
make seed
make kafka-test
make down
```

Local URLs:

- SAP mock API: `http://localhost:8080`
- Ingestion API: `http://localhost:8081`
- Query API: `http://localhost:8083`
- Kafka UI: `http://localhost:8085`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Run On Kubernetes

For a persistent local Kubernetes cluster with MicroK8s:

```bash
make microk8s-deploy
make microk8s-smoke
```

For GKE dev deployment:

```bash
cp terraform/envs/dev/backend.hcl.example terraform/envs/dev/backend.hcl
cp terraform/envs/dev/terraform.tfvars.example terraform/envs/dev/terraform.tfvars

terraform -chdir=terraform/envs/dev init -backend-config=backend.hcl
terraform -chdir=terraform/envs/dev apply -var-file=terraform.tfvars

make gke-deploy
make gke-status
make gke-smoke
```

Open observability on GKE:

```bash
make gke-grafana
make gke-prometheus
make gke-observability-status
make gke-observability-stop
```

## CI And Validation

Run the main local quality gate:

```bash
make ci
```

The repository also includes GitHub Actions workflows for linting, testing, building, Docker image publishing and controlled dev/prod deployment.

## Notes

- Production deployment is designed to be manual or approval-gated.
- The project keeps local development simple while preserving cloud-ready design choices such as Workload Identity, Secret Manager, Helm values per environment and Terraform modules.
- Detailed documentation lives in `docs/architecture.md`, `docs/decisions/`, `docs/runbooks/`, `deploy/helm/platform/README.md` and `terraform/README.md`.