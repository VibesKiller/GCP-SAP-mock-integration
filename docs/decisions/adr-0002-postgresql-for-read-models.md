# ADR-0002: Use Cloud SQL PostgreSQL for Read Models

## Status

Accepted

## Context

The platform needs queryable projections for customers, sales orders and invoices while keeping event ingestion asynchronous.

The read path should be simple, relational and easy to explain. The project does not require a distributed analytical store or a search engine for the current scope.

## Decision

Use PostgreSQL for read models and minimal event audit state.

Local development uses PostgreSQL in Docker Compose. GCP uses Cloud SQL for PostgreSQL with private networking. The `processed_events` table provides idempotency and an audit trail for processed Kafka messages.

## Consequences

- Query APIs remain low-latency and straightforward.
- Schema migrations become part of the platform lifecycle.
- Event processing must apply projections transactionally with `processed_events`.
- Cloud SQL operational concerns such as private IP, backups and deletion behavior are represented in Terraform.
