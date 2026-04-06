# ADR-0002: Use PostgreSQL for Read Models

## Status

Accepted

## Context

The platform needs queryable projections for downstream APIs while remaining event-driven.

## Decision

PostgreSQL is used for read models and operational queries. Kafka remains the streaming backbone, while PostgreSQL provides low-latency query access for the read-only API layer.

## Consequences

- Write-side processing remains asynchronous.
- Query shapes can evolve independently from the raw event model.
- Backup, migration and schema governance become part of the platform story.
