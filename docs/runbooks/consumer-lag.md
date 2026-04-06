# Runbook: Consumer Lag Investigation

## Trigger

Prometheus raises `HighConsumerLag` for any platform consumer group.

## Checks

1. Verify broker health and partition availability.
2. Inspect the affected consumer group in Kafka UI.
3. Confirm whether the consumer deployment is healthy in Kubernetes or local Docker.
4. Review recent deployments, schema changes and downstream database latency.
5. Check DLQ traffic for correlated failures.

## Initial Mitigations

- restart a stuck consumer only after confirming idempotent processing guarantees
- scale out consumer replicas if partitioning allows it
- pause upstream traffic generation if the backlog threatens downstream SLAs
