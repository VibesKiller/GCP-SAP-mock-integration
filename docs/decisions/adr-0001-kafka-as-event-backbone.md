# ADR-0001: Use Apache Kafka as the Event Backbone

## Status

Accepted

## Context

The platform must model a realistic enterprise integration workload where SAP-like business events are ingested, replayed and consumed by multiple downstream capabilities.

## Decision

Apache Kafka is the central event backbone for the platform. Topic design, consumer group ownership and DLQ handling are defined as first-class platform concerns.

## Consequences

- The repository is structured around explicit topic governance.
- Local development uses Kafka in Docker Compose rather than an in-memory or queue-compatible substitute.
- Consumers are designed with replay, lag monitoring and idempotency in mind.
