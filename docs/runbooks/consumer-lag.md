# Runbook: Consumer Lag Investigation

## Trigger

Prometheus raises `HighConsumerLag` for `sap-integration.event-processor.v1`, or Kafka UI shows a growing backlog on business topics.

## Scope

Primary consumer group:

- `sap-integration.event-processor.v1`

Topics:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`

## Immediate Checks

Local:

```bash
make kafka-test
docker compose -f deploy/local/docker-compose.yml exec -T kafka kafka-consumer-groups \
  --bootstrap-server kafka:29092 \
  --describe \
  --group sap-integration.event-processor.v1
```

GKE:

```bash
kubectl -n sap-integration-dev get pods,hpa
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=200
kubectl -n sap-integration-dev top pods
```

Database:

```bash
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-query-api --tail=100
```

## What To Validate

- `event-processor` is running and ready.
- Database latency or connection errors are not blocking processing.
- Retry metrics are not growing rapidly.
- DLQ count is not increasing at the same time.
- Topic partition count supports the desired number of consumer replicas.

## Mitigation

- If the consumer is stuck, restart `event-processor` after confirming idempotency is active.

```bash
kubectl -n sap-integration-dev rollout restart deploy/sap-integration-platform-event-processor
```

- If CPU or memory is saturated, scale the deployment only up to the useful partition parallelism.

```bash
kubectl -n sap-integration-dev scale deploy/sap-integration-platform-event-processor --replicas=2
```

- If the database is the bottleneck, fix database health before adding consumers.

## Follow-Up

- Compare event processing latency before and after mitigation.
- Review recent deploys, schema changes and Cloud SQL operations.
- Add a capacity note if sustained event volume exceeds current partition or database capacity.
