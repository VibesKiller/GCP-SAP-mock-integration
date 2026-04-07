# Runbook: Dead-Letter Topic Growth

## Trigger

- Prometheus alert `DlqTrafficDetected` fires.
- Kafka UI shows messages accumulating in `sap.integration.dlq.v1`.
- `event-processor` logs `message published to dlq`.

## Immediate Checks

Local:

```bash
make kafka-test
docker compose -f deploy/local/docker-compose.yml logs --tail=200 event-processor
```

GKE:

```bash
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=200
kubectl -n sap-integration-dev get pods
```

Kafka topic catalog:

```bash
sed -n '1,220p' platform/kafka/topic-catalog.yaml
```

## What To Validate

- DLQ message headers include `failure_reason`, `original_topic`, `original_partition`, `original_offset`, `original_key`, `event_id`, `event_type`, `version`, `source`, `correlation_id` and `partition_key` where available.
- The failure is not caused by a schema mismatch or unsupported `event_type`.
- The database is healthy and not causing retries to exhaust.
- `processed_events` contains DLQ status for events with valid event IDs.

## Mitigation

- Stop replay attempts until the root cause is understood.
- Fix the producer, schema normalization or processor logic first.
- If the issue is caused by bad upstream payloads, keep the DLQ as the source of forensic truth.

## Replay Policy

Replay is not an ad-hoc shell operation in production.

- Replay only after root cause resolution.
- Preserve the original `correlation_id` for traceability.
- Use a new `event_id` when replaying through ingestion, otherwise `processed_events` will treat the message as already handled.
- Record the replay decision, time window and message count in incident notes.

See also `docs/runbooks/dlq-replay.md`.
