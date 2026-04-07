# ADR-0004: Use Go for Runtime Services

## Status

Accepted

## Context

The application services need to be small, fast, easy to containerize and clear to read in an interview context.

The project benefits from a language that supports simple HTTP services, Kafka clients, PostgreSQL access, structured logging and Prometheus instrumentation without a heavy framework.

## Decision

Implement the runtime services in Go.

The services use the standard library HTTP server, `log/slog` for JSON structured logs, `pgx` for PostgreSQL, `kafka-go` for Kafka and Prometheus client libraries for metrics.

## Consequences

- Services remain compact and easy to reason about.
- Container images can be built with straightforward multi-stage Dockerfiles.
- Cross-service conventions such as health endpoints, correlation IDs and config via environment variables can be shared through internal packages.
- Kafka authentication for GCP requires explicit client configuration, which is handled through environment-driven Kafka client helpers.
