# ADR-0001: Use Apache Kafka as the Event Backbone

## Status

Accepted

## Context

The platform models a realistic SAP-style integration workload where business events must be ingested, durably stored, replayable and consumed asynchronously by downstream services.

A simple HTTP-only integration would be easier to implement, but it would not demonstrate the event-driven platform concerns expected in enterprise SAP, cloud and data integration scenarios.

## Decision

Use Apache Kafka as the central event backbone.

Kafka topics are explicit integration contracts. Topic naming, message keys, headers, consumer group ownership and DLQ behavior are documented and versioned in the repository.

Local development uses Apache Kafka in Docker Compose. GCP uses Managed Service for Apache Kafka.

## Consequences

- Producers and consumers are decoupled by durable event streams.
- Message keys preserve aggregate-level ordering for customers, sales orders and invoices.
- Consumers must be idempotent because Kafka delivery can be at least once.
- Consumer lag, retry behavior and DLQ growth become first-class operational signals.
- The project remains Kafka-native and does not depend on queue-compatible substitutes.
