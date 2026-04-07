# Runbook: Kafka Broker Down

## Trigger

- `ingestion-api` readiness fails because Kafka is unreachable.
- `event-processor` logs repeated `fetch kafka message` or dial errors.
- Prometheus alert `KafkaPublishFailures` fires.
- Local `make kafka-test` cannot list topics or describe the consumer group.

## Immediate Checks

Local Docker Compose:

```bash
make ps
docker compose -f deploy/local/docker-compose.yml logs --tail=200 kafka
make kafka-topics
make kafka-test
```

GKE:

```bash
kubectl -n sap-integration-dev get pods
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-ingestion-api --tail=100
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=100
kubectl -n sap-integration-dev describe configmap sap-integration-platform-ingestion-api
```

GCP Managed Kafka:

```bash
gcloud managed-kafka clusters list --location=<region> --project=<project-id>
gcloud managed-kafka clusters describe <cluster-name> --location=<region> --project=<project-id>
```

## What To Validate

- `KAFKA_BROKERS` points to the correct bootstrap servers.
- `KAFKA_AUTH_MODE` is `none` locally and `google_access_token` on GKE/GCP.
- TLS is enabled for GCP Managed Kafka.
- Workload Identity annotations exist for `ingestion-api` and `event-processor`.
- Kafka ACLs allow the producer to write and the consumer group to read.

## Mitigation

- Local: restart only Kafka-dependent services after broker recovery.

```bash
docker compose -f deploy/local/docker-compose.yml restart ingestion-api event-processor
```

- GKE: restart the affected deployments after confirming Kafka has recovered.

```bash
kubectl -n sap-integration-dev rollout restart deploy/sap-integration-platform-ingestion-api
kubectl -n sap-integration-dev rollout restart deploy/sap-integration-platform-event-processor
```

## Follow-Up

- Check for consumer lag growth after broker recovery.
- Check DLQ traffic for messages that failed during the outage window.
- Record whether the failure was broker availability, auth, DNS, ACL or client configuration.
