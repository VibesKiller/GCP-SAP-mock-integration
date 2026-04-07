# Runbook: Dead-Letter Queue Replay

## Purpose

Define the ownership and safe process for handling non-recoverable Kafka messages routed to the DLQ.

## Ownership

- Topic: `sap.integration.dlq.v1`
- Owner: `event-processor`
- Primary investigator: platform engineer on call for the integration platform

## Preconditions

Replay must not start until all of these are true:

- the root cause is identified and fixed
- the affected time window is known
- the replay count is estimated
- the target environment is confirmed
- stakeholders agree whether replay should preserve or supersede existing projections

## Investigation Commands

Local:

```bash
make kafka-test
docker compose -f deploy/local/docker-compose.yml logs --tail=200 event-processor
```

GKE:

```bash
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=200
```

Database audit:

```sql
select event_id, event_type, status, processed_at
from processed_events
where status <> 'processed'
order by processed_at desc
limit 50;
```

## Replay Principles

- Replay only after identifying the root cause.
- Preserve the original correlation ID when re-submitting a message.
- Use a new `event_id` when replaying through the normal ingestion path; otherwise `processed_events` will treat it as already handled.
- Never replay from production directly into a development cluster.
- Record replay decisions in incident notes.

## Tooling Direction

Replay tooling should be implemented as a controlled admin job or script that:

- reads from `sap.integration.dlq.v1`
- filters by original topic, event type and time window
- rewrites the event ID when replaying through ingestion
- keeps the original correlation ID and source metadata
- emits a replay report

The current project intentionally documents this operational path without adding a broad admin API surface.
