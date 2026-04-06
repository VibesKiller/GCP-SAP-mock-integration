# Local Development Environment

This Compose stack is the opinionated local runtime for the project. It uses real Apache Kafka in KRaft mode and keeps the bootstrap intentionally simple.

## Included Components

- Apache Kafka broker
- Kafka UI for topic inspection
- PostgreSQL with bootstrap schema migration
- Prometheus
- Grafana
- `sap-mock-api`
- `ingestion-api`
- `event-processor`
- `query-api`

## Kafka Local Conventions

Bootstrap servers:

- host machine: `localhost:9092`
- internal Docker network: `kafka:29092`

Topics:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

Consumer groups:

- `sap-integration.event-processor.v1`

## Recommended Workflow

```bash
cp .env.example .env
make up
make seed
make smoke
```

## Useful Commands

```bash
make kafka-topics
make kafka-test
make logs
make down
```
