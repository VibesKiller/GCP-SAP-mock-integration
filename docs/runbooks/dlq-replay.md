# Runbook: Dead-Letter Queue Replay

## Purpose

Define the ownership and process for handling non-recoverable Kafka messages routed to DLQ topics.

## Ownership

- `sap.integration.dlq.v1`: event-processor owner

## Replay Principles

- Replay only after identifying the root cause.
- Preserve the original correlation ID when re-submitting a message.
- Use a new `event_id` when replaying through the normal ingestion path, otherwise `processed_events` will treat it as already handled.
- Record replay decisions in the relevant incident notes.
- Never replay from production directly into a development cluster.

## Tooling Direction

Replay tooling will be implemented in a later step as a controlled script or admin job, not as ad-hoc manual commands.
