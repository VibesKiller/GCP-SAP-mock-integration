# GCP SAP Mock Integration Platform

A Kafka-native SAP integration platform on Google Cloud, implemented in Go and shaped to be credible in senior DevOps, Cloud and Platform Engineering interviews.

## Why This Repository Exists

This project is intentionally production-like rather than tutorial-like.

It demonstrates:

- Apache Kafka as the real event backbone
- clear runtime boundaries across Go services
- PostgreSQL read models with idempotent event processing
- Docker Compose for a realistic local developer workflow
- structured JSON logging with correlation IDs
- Prometheus metrics and operational endpoints
- Terraform and Helm boundaries aligned with enterprise delivery

## Business Flow

A mock SAP system emits business events such as:

- `sales_order.created`
- `sales_order.updated`
- `customer.updated`
- `invoice.issued`

The platform:

1. receives SAP-style payloads over REST
2. validates and normalizes them into a canonical envelope
3. publishes them to Apache Kafka
4. consumes them asynchronously with retry and DLQ handling
5. persists projections in PostgreSQL
6. exposes read-only APIs for business queries

## Services

### `sap-mock-api`

Simulates realistic upstream SAP payloads and can dispatch them to the ingestion edge.

Core endpoints:

- `GET /api/v1/sample-data`
- `POST /api/v1/simulations/sales-orders/create`
- `POST /api/v1/simulations/sales-orders/update`
- `POST /api/v1/simulations/customers/update`
- `POST /api/v1/simulations/invoices/issue`

### `ingestion-api`

Validates incoming SAP payloads, normalizes them and publishes canonical events to Kafka.

Core endpoints:

- `POST /api/v1/sap/sales-orders`
- `PATCH /api/v1/sap/sales-orders/{orderID}`
- `PATCH /api/v1/sap/customers/{customerID}`
- `POST /api/v1/sap/invoices`
- `GET /health`, `GET /ready`, `GET /live`, `GET /metrics`

### `event-processor`

Consumes Kafka topics, enforces idempotency through `processed_events`, persists PostgreSQL projections and sends non-recoverable failures to the DLQ.

Runtime endpoints:

- `GET /health`, `GET /ready`, `GET /live`, `GET /metrics`

### `query-api`

Exposes read-only APIs for customers, orders and invoices with deterministic pagination and base filters.

Core endpoints:

- `GET /api/v1/customers`
- `GET /api/v1/customers/{customerID}`
- `GET /api/v1/orders`
- `GET /api/v1/orders/{orderID}`
- `GET /api/v1/invoices`
- `GET /api/v1/invoices/{invoiceID}`

## Canonical Event Envelope

All incoming SAP payloads are transformed into this internal contract before entering Kafka:

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

## Kafka Topology

### Topics

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

### Consumer Groups

- `sap-integration.event-processor.v1`
- `sap-integration.notification-worker.v1` reserved for future fan-out

### Message Keys And Partitioning

- `sap.sales-orders.v1`: key = `sales_order_id`
- `sap.customers.v1`: key = `customer_id`
- `sap.invoices.v1`: key = `invoice_id`
- `sap.integration.dlq.v1`: key = original partition key when available

### Bootstrap Servers

- local host access: `localhost:9092`
- internal Docker network access: `kafka:29092`

Reference files:

- [topic-catalog.yaml](/home/git/GIT/GCP-SAP-mock-integration/platform/kafka/topic-catalog.yaml)
- [consumer-groups.yaml](/home/git/GIT/GCP-SAP-mock-integration/platform/kafka/consumer-groups.yaml)
- [kafka-topic-naming.md](/home/git/GIT/GCP-SAP-mock-integration/platform/policies/kafka-topic-naming.md)

## Persistence Model

PostgreSQL schema includes:

- `customers`
- `orders`
- `order_items`
- `invoices`
- `processed_events`

`processed_events` is both the idempotency guard and the minimum audit trail for Kafka processing.

Schema reference:

- [001_initial_schema.sql](/home/git/GIT/GCP-SAP-mock-integration/platform/database/migrations/001_initial_schema.sql)

## Local Quick Start

### Prerequisites

- Docker
- Docker Compose
- Go 1.24+

### 1. Prepare Local Configuration

```bash
cp .env.example .env
```

You can keep the defaults for a first run. They are already tuned for the local stack.

### 2. Start The Full Local Platform

```bash
make up
```

`make up` does four things:

1. starts Docker Compose
2. creates the required Kafka topics
3. restarts `event-processor` once so local consumers see freshly bootstrapped topics
4. waits for the runtime endpoints to become reachable

### 3. Seed Realistic Demo Data

```bash
make seed
```

This sends realistic business events through `sap-mock-api` with `dispatch=true`, so the full path is exercised.

### 4. Run The End-To-End Smoke Test

```bash
make smoke
```

The smoke test verifies:

- health endpoints
- event creation from `sap-mock-api`
- topic offset growth in Kafka
- consumption and persistence into PostgreSQL
- read-only query paths in `query-api`

## Local Developer Commands

- `make up`
- `make down`
- `make logs`
- `make logs SERVICE=query-api`
- `make build`
- `make test`
- `make lint`
- `make fmt`
- `make seed`
- `make smoke`
- `make kafka-topics`
- `make kafka-test`

## Local Service URLs

- sap-mock-api: `http://localhost:8080`
- ingestion-api: `http://localhost:8081`
- event-processor: `http://localhost:8082`
- query-api: `http://localhost:8083`
- Kafka UI: `http://localhost:8085`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Testing Strategy

This repository intentionally keeps tests small but meaningful.

Included in this step:

- domain validation tests for the canonical event envelope
- normalization tests from SAP payloads to canonical payloads
- Kafka topic and message key derivation tests
- query API pagination tests

Run them with:

```bash
make test
```

## OpenAPI Contracts

- [platform-api.yaml](/home/git/GIT/GCP-SAP-mock-integration/api/openapi/platform-api.yaml)
- [ingestion-api.yaml](/home/git/GIT/GCP-SAP-mock-integration/api/openapi/ingestion-api.yaml)
- [query-api.yaml](/home/git/GIT/GCP-SAP-mock-integration/api/openapi/query-api.yaml)

## Local Documentation

- [deploy/local/README.md](/home/git/GIT/GCP-SAP-mock-integration/deploy/local/README.md)
- [local-development.md](/home/git/GIT/GCP-SAP-mock-integration/docs/runbooks/local-development.md)
- [dlq-replay.md](/home/git/GIT/GCP-SAP-mock-integration/docs/runbooks/dlq-replay.md)

## Current Status

Implemented so far:

- four Go services with clear responsibilities
- Kafka producer and consumer code for real local Kafka
- PostgreSQL schema and query projections
- idempotency through `processed_events`
- retry and DLQ handling in the processor
- OpenAPI contracts for ingestion and query APIs
- Docker Compose local environment
- root `.env.example` and service-level `.env.example` files
- Makefile workflow for bootstrap, testing and smoke verification
- minimal but credible Go unit tests

## Next Steps

1. add Helm manifests for Deployments, Services, probes and Workload Identity-aware configuration
2. extend CI with richer checks and image build validation
3. add repository and handler tests around PostgreSQL and HTTP edge cases
4. wire Terraform outputs into deployment values for GKE environments
