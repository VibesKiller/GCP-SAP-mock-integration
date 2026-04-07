# GCP SAP Mock Integration Platform

Kafka-native SAP integration platform on Google Cloud, implemented in Go and packaged with Terraform, Helm and GitHub Actions.

The repository is designed as a senior DevOps / Cloud / Platform Engineering portfolio project: realistic enough to discuss in enterprise interviews, but intentionally small enough to run locally, deploy to GKE and reason about end to end.

## What This Project Demonstrates

- Event-driven integration with Apache Kafka as the primary backbone, not a queue-compatible substitute.
- Go services with clear runtime boundaries, health endpoints, JSON logs, correlation IDs and Prometheus metrics.
- Idempotent asynchronous processing with retry and a dead-letter topic.
- PostgreSQL read models backed by a `processed_events` audit table.
- Local development with Docker Compose, Kafka, PostgreSQL, Prometheus and Grafana.
- Kubernetes deployment on GKE using Helm with probes, resource limits, HPA and optional NetworkPolicy.
- Google Cloud provisioning with Terraform modules for VPC, GKE, Cloud SQL, Artifact Registry, IAM, Secret Manager and Managed Service for Apache Kafka.
- CI/CD workflows for lint, test, build, image publishing and environment deployment.

## Architecture Summary

```text
SAP mock events
  -> sap-mock-api
  -> ingestion-api
  -> Apache Kafka topics
  -> event-processor
  -> PostgreSQL projections
  -> query-api
```

Runtime services:

- `sap-mock-api`: simulates SAP-originated events and can dispatch realistic sample payloads.
- `ingestion-api`: validates SAP payloads, normalizes them and publishes canonical events to Kafka.
- `event-processor`: consumes Kafka, handles retry/DLQ, persists projections and records idempotency state.
- `query-api`: exposes read-only APIs for customers, sales orders and invoices.
- `notification-worker`: reserved extension point for future fan-out.

Cloud/platform layer:

- GKE for application workloads.
- Cloud SQL PostgreSQL for read models.
- Artifact Registry for service images.
- Secret Manager as secret source of truth.
- Managed Service for Apache Kafka on GCP for cloud event streaming.
- Helm for workload packaging and runtime configuration.
- GitHub Actions for CI/CD orchestration.

## Business Flow

Supported SAP-style business events:

- `sales_order.created`
- `sales_order.updated`
- `customer.updated`
- `invoice.issued`

Each event follows this path:

1. `sap-mock-api` emits or dispatches a realistic SAP payload.
2. `ingestion-api` validates required fields and normalizes the payload.
3. `ingestion-api` wraps the payload in a canonical event envelope.
4. The event is published to Kafka using an aggregate-specific message key.
5. `event-processor` consumes the topic using a stable consumer group.
6. `event-processor` inserts into `processed_events` to enforce idempotency.
7. Domain projections are persisted into PostgreSQL tables.
8. `query-api` exposes read-only query endpoints over the projections.

## Canonical Event Envelope

```json
{
  "event_id": "uuid",
  "event_type": "sales_order.created",
  "version": "v1",
  "source": "sap-s4hana",
  "occurred_at": "2026-04-02T09:45:00Z",
  "correlation_id": "uuid",
  "payload": {}
}
```

The envelope is the stable internal contract between ingestion and downstream processing. SAP-specific payload differences are normalized before Kafka.

## Kafka Role In The Architecture

Kafka is the integration backbone and event durability boundary.

Topics:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

Consumer groups:

- `sap-integration.event-processor.v1`
- `sap-integration.notification-worker.v1` reserved for future fan-out

Partitioning strategy:

- sales order events use `sales_order_id`
- customer events use `customer_id`
- invoice events use `invoice_id`
- DLQ events reuse the original partition key when available

Kafka headers:

- `event_id`
- `event_type`
- `version`
- `source`
- `correlation_id`
- `partition_key`
- DLQ failure metadata such as `failure_reason`, `original_topic`, `original_partition`, `original_offset`

Local Kafka bootstrap servers:

- host: `localhost:9092`
- Docker network: `kafka:29092`
- MicroK8s bridge listener: configured by `make microk8s-deploy`

GCP Kafka mode:

- Managed Service for Apache Kafka
- TLS enabled
- Kafka clients configured with `KAFKA_AUTH_MODE=google_access_token`
- authentication uses Google ADC / Workload Identity and a workload-specific Google service account principal

Reference files:

- `platform/kafka/topic-catalog.yaml`
- `platform/kafka/consumer-groups.yaml`
- `platform/policies/kafka-topic-naming.md`

## Persistence Model

PostgreSQL tables:

- `customers`
- `orders`
- `order_items`
- `invoices`
- `processed_events`

`processed_events` provides:

- idempotency guard keyed by `event_id`
- minimal audit trail with Kafka topic, partition, offset and headers
- processing status for normal and DLQ outcomes

Schema reference: `platform/database/migrations/001_initial_schema.sql`.

## APIs

Common runtime endpoints:

- `GET /health`
- `GET /ready`
- `GET /live`
- `GET /metrics`

SAP mock API:

- `GET /api/v1/sample-data`
- `POST /api/v1/simulations/sales-orders/create`
- `POST /api/v1/simulations/sales-orders/update`
- `POST /api/v1/simulations/customers/update`
- `POST /api/v1/simulations/invoices/issue`

Ingestion API:

- `POST /api/v1/sap/sales-orders`
- `PATCH /api/v1/sap/sales-orders/{orderID}`
- `PATCH /api/v1/sap/customers/{customerID}`
- `POST /api/v1/sap/invoices`

Query API:

- `GET /api/v1/customers`
- `GET /api/v1/customers/{customerID}`
- `GET /api/v1/orders`
- `GET /api/v1/orders/{orderID}`
- `GET /api/v1/invoices`
- `GET /api/v1/invoices/{invoiceID}`

OpenAPI contracts:

- `api/openapi/ingestion-api.yaml`
- `api/openapi/query-api.yaml`
- `api/openapi/platform-api.yaml`

## Local Quick Start

Prerequisites:

- Docker
- Docker Compose
- Go 1.24+

Run:

```bash
cp .env.example .env
make up
make smoke
```

Useful commands:

```bash
make logs
make logs SERVICE=event-processor
make seed
make kafka-topics
make kafka-test
make down
```

Local URLs:

- `sap-mock-api`: `http://localhost:8080`
- `ingestion-api`: `http://localhost:8081`
- `event-processor`: `http://localhost:8082`
- `query-api`: `http://localhost:8083`
- Kafka UI: `http://localhost:8085`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Kubernetes Paths

Persistent local Kubernetes with MicroK8s:

```bash
make microk8s-deploy
make microk8s-smoke
```

GKE dev deployment after Terraform provisioning:

```bash
terraform -chdir=terraform/envs/dev apply -var-file=terraform.tfvars
make gke-deploy
make gke-status
make gke-smoke
```

The dev Helm overlay also deploys a compact namespace-local Prometheus and Grafana stack. Access the GKE dashboard with:

```bash
make gke-grafana
```

Then open:

```text
http://localhost:3000/d/sap-integration-platform-overview/sap-integration-platform-overview
```

`make gke-grafana` and `make gke-prometheus` start detached port-forwards. Use `make gke-observability-status` and `make gke-observability-stop` to inspect or stop them.

Production deployment is intentionally manual and should be protected through a GitHub Environment approval gate.

## CI/CD

GitHub Actions workflows:

- `ci`: Go lint/test/build, Docker build smoke, Helm lint/render, Terraform fmt/validate and repository secret hygiene.
- `docker images`: builds and pushes service images to Artifact Registry using Workload Identity Federation.
- `deploy dev`: manual dev deployment, optional Terraform apply, Helm deploy and smoke test.
- `deploy prod`: manual production deployment gated by a confirmation input and GitHub Environment approval.
- `terraform`: Terraform-only workflow for infrastructure changes.

Expected GitHub variables:

- `GCP_PROJECT_ID`
- `GCP_REGION`
- `ARTIFACT_REPOSITORY`
- `TF_STATE_BUCKET`

Expected GitHub secrets:

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_CICD_SERVICE_ACCOUNT`

No service account JSON key is required.

## Observability

Prometheus scrapes all Go services through `/metrics`.

Core application metrics:

- `sap_integration_sap_mock_api_simulations_total`
- `sap_integration_sap_mock_api_dispatch_total`
- `sap_integration_ingestion_api_requests_total`
- `sap_integration_ingestion_api_published_total`
- `sap_integration_ingestion_api_publish_duration_seconds`
- `sap_integration_event_processor_consumed_total`
- `sap_integration_event_processor_retries_total`
- `sap_integration_event_processor_dlq_total`
- `sap_integration_event_processor_processing_duration_seconds`
- `sap_integration_query_api_requests_total`

Operational assets:

- Prometheus config: `platform/monitoring/prometheus.yml`
- Alert rules: `platform/alerts/platform-alerts.yaml`
- Grafana dashboard: `platform/dashboards/platform-overview.json`
- Consumer lag runbook: `docs/runbooks/consumer-lag.md`

GKE dev observability:

- Helm deploys Prometheus and Grafana when `observability.enabled=true`
- dashboard assets are packaged under `deploy/helm/platform/files/`
- `make gke-smoke` validates Prometheus/Grafana readiness when enabled

Consumer lag requires a Kafka exporter or managed Kafka lag metric integration. The dashboard and alert rule are prepared for `kafka_consumergroup_lag` when that metric is available.

## Logging

All services use JSON structured logs via `log/slog`.

Common log fields:

- `service`
- `environment`
- `method`
- `path`
- `status`
- `duration_ms`
- `correlation_id`

Event-specific fields:

- `event_id`
- `event_type`
- `topic`
- `message_key`
- `partition`
- `offset`
- `failure_reason`

Correlation IDs are accepted from `X-Correlation-ID`, generated when missing and propagated from SAP mock dispatch into ingestion and event processing.

## Security Posture

- No secrets are committed to the repository.
- `.tfvars`, Terraform state, backend config, local `.env` files, kubeconfigs and service account keys are ignored.
- Workload Identity is used for GKE workloads.
- Secret Manager stores generated runtime secrets.
- Helm references Kubernetes Secrets instead of embedding secret values.
- Containers are configured as non-root where practical.
- Kafka and database permissions are documented using least-privilege ownership boundaries.

## Documentation Map

- Architecture: `docs/architecture.md`
- ADRs: `docs/decisions/`
- Runbooks: `docs/runbooks/`
- Diagrams: `docs/diagrams/`
- Helm chart: `deploy/helm/platform/README.md`
- Terraform: `terraform/README.md`
- Local runtime: `deploy/local/README.md`

## How To Present This Project In Interviews

Use this framing:

"I built a Kafka-native SAP integration platform on GCP to show the full platform engineering path: ingestion, canonical event modeling, Kafka topic governance, idempotent consumers, PostgreSQL projections, GKE deployment, Terraform infrastructure, Helm packaging, observability and CI/CD. The local stack runs with real Kafka in Docker Compose, while the cloud path uses GKE, Cloud SQL, Artifact Registry, Secret Manager and Managed Service for Apache Kafka. The goal was not to build a huge system, but to demonstrate production tradeoffs in a compact architecture."

Strong talking points:

- why Kafka is the event boundary rather than an implementation detail
- how message keys preserve aggregate ordering
- how `processed_events` prevents duplicate processing
- how retry differs from DLQ handling
- why Terraform provisions infrastructure while Helm deploys workloads
- how Workload Identity avoids static cloud credentials
- how the smoke tests prove the full API to DB path

## CV Bullet Points

- Designed and implemented a Kafka-native SAP-style integration platform on GCP using Go, GKE, Cloud SQL PostgreSQL, Terraform, Helm and GitHub Actions.
- Built idempotent event processing with Apache Kafka topics, consumer groups, retry handling, DLQ routing and PostgreSQL audit tracking through `processed_events`.
- Provisioned reusable Terraform modules for VPC networking, GKE, Cloud SQL, Artifact Registry, IAM service accounts, Secret Manager and Managed Service for Apache Kafka.
- Packaged production-like Kubernetes workloads with Helm, including probes, resource limits, HPA, service accounts, secret references and environment overlays.
- Added Prometheus/Grafana observability, JSON structured logging with correlation IDs and end-to-end smoke tests covering mock SAP API to database persistence.

## LinkedIn Bullet Points

- Built a compact enterprise-style SAP integration platform on Google Cloud with Go, Apache Kafka, GKE, Terraform, Helm and Cloud SQL.
- Demonstrated production concerns end to end: idempotency, DLQ, retry, observability, Workload Identity, CI/CD and cloud deployment smoke tests.
- Kept the design intentionally realistic: real Kafka locally with Docker Compose, Managed Kafka on GCP, clear platform boundaries and interview-ready documentation.

## Final Execution Checklist

Local:

```bash
cp .env.example .env
make up
make smoke
make kafka-test
```

MicroK8s:

```bash
make microk8s-deploy
make microk8s-smoke
```

GCP dev:

```bash
cp terraform/envs/dev/backend.hcl.example terraform/envs/dev/backend.hcl
cp terraform/envs/dev/terraform.tfvars.example terraform/envs/dev/terraform.tfvars
terraform -chdir=terraform/envs/dev init -backend-config=backend.hcl
terraform -chdir=terraform/envs/dev apply -var-file=terraform.tfvars
make gke-deploy
make gke-smoke
```

CI/CD:

```bash
make ci
```

Operational verification:

```bash
make gke-status
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=100
kubectl -n sap-integration-dev get hpa,pods,svc
```
