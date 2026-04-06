# Runbook: Local Development Bootstrap

## Purpose

Start the full SAP integration platform locally with real Apache Kafka, PostgreSQL, Prometheus, Grafana and the Go services.

## Prerequisites

- Docker
- Docker Compose
- Go 1.24+

## Local Configuration

1. Copy the root environment template.
2. Keep the defaults unless you need different host ports.

```bash
cp .env.example .env
```

Primary local values:

- Kafka bootstrap server from the host: `localhost:9092`
- Kafka bootstrap server inside Docker: `kafka:29092`
- PostgreSQL host port: `5432`
- `sap-mock-api`: `8080`
- `ingestion-api`: `8081`
- `event-processor`: `8082`
- `query-api`: `8083`

## Bootstrap The Platform

```bash
make up
```

This command:

1. starts the Docker Compose stack
2. creates the required Kafka topics
3. restarts `event-processor` once so the local consumer subscribes after topic bootstrap
4. waits for the application health endpoints

Kafka topics created locally:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

Consumer group used by the processor:

- `sap-integration.event-processor.v1`

## Seed Demo Data

```bash
make seed
```

The seed flow dispatches realistic sample business events through `sap-mock-api`, so the full path is exercised:

1. customer update
2. sales order creation
3. sales order update
4. invoice issuance

## Run The Smoke Test

```bash
make smoke
```

The smoke test verifies:

1. health endpoints for all services
2. Kafka topic readiness
3. event production through the mock SAP flow
4. topic offset growth in Kafka
5. event persistence into `processed_events`
6. customer, order and invoice retrieval through `query-api`

## Useful Commands

```bash
make logs
make logs SERVICE=ingestion-api
make kafka-topics
make kafka-test
make test
make lint
make down
```

## Local URLs

- `sap-mock-api`: `http://localhost:8080`
- `ingestion-api`: `http://localhost:8081`
- `event-processor`: `http://localhost:8082`
- `query-api`: `http://localhost:8083`
- Kafka UI: `http://localhost:8085`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Manual Demo Flow

```bash
curl -X POST 'http://localhost:8080/api/v1/simulations/customers/update?dispatch=true'
curl -X POST 'http://localhost:8080/api/v1/simulations/sales-orders/create?dispatch=true'
curl -X POST 'http://localhost:8080/api/v1/simulations/invoices/issue?dispatch=true'
curl 'http://localhost:8083/api/v1/customers'
curl 'http://localhost:8083/api/v1/orders'
curl 'http://localhost:8083/api/v1/invoices'
```
