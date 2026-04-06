# Architecture Overview

## Purpose

This repository implements a realistic SAP-style integration platform on Google Cloud with Apache Kafka as the system-of-record for asynchronous business events.

## Service Responsibilities

- `sap-mock-api` simulates SAP-originated business payloads and can optionally dispatch them to the ingestion edge.
- `ingestion-api` validates SAP payloads, normalizes them into a canonical event envelope and publishes them to Kafka.
- `event-processor` consumes Kafka topics, enforces idempotency, persists PostgreSQL projections and sends poison messages to the DLQ.
- `query-api` exposes read-only APIs over PostgreSQL read models.

## Canonical Envelope

Every business event is normalized to this envelope:

- `event_id`
- `event_type`
- `version`
- `source`
- `occurred_at`
- `correlation_id`
- `payload`

This gives the platform a stable internal contract regardless of how SAP payloads evolve at the edge.

## Kafka Topology

Primary topics:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

Consumer group:

- `sap-integration.event-processor.v1`

Partitioning strategy:

- sales orders by `sales_order_id`
- customers by `customer_id`
- invoices by `invoice_id`
- DLQ by original partition key when available

Useful Kafka headers:

- `event_id`
- `event_type`
- `version`
- `source`
- `correlation_id`
- `partition_key`
- DLQ-specific replay headers

## Persistence Model

PostgreSQL stores projections and audit data in the following tables:

- `customers`
- `orders`
- `order_items`
- `invoices`
- `processed_events`

`processed_events` is the idempotency guard and minimum audit trail for processed Kafka messages.

## Operational Notes

- the processor uses transactional inserts into `processed_events` plus domain tables to avoid marking failed work as processed
- the processor creates controlled placeholder records for out-of-order reference data where reasonable
- health, readiness and liveness endpoints are available for every runtime service
- structured JSON logs propagate `correlation_id` from the ingestion edge to the processor and query APIs
