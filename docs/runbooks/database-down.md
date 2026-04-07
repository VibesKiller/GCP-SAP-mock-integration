# Runbook: Database Down

## Trigger

- `query-api` readiness fails.
- `event-processor` logs PostgreSQL connection errors.
- Query API returns 5xx responses.
- Cloud SQL maintenance, failover or connectivity issues are visible in GCP.

## Immediate Checks

GKE:

```bash
kubectl -n sap-integration-dev get pods
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=100
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-query-api --tail=100
kubectl -n sap-integration-dev get secret sap-integration-postgres-dev
```

Cloud SQL:

```bash
gcloud sql instances list
gcloud sql instances describe <instance-name> --project=<project-id>
gcloud sql operations list --instance=<instance-name> --project=<project-id> --limit=10
```

## What To Validate

- `POSTGRES_URL` exists in the Kubernetes secret and points to the private Cloud SQL address.
- GKE workloads can route to the Cloud SQL private IP through the VPC.
- Cloud SQL instance is running and not in maintenance or failed state.
- Database schema has been applied from `platform/database/migrations/001_initial_schema.sql`.
- `event-processor` retry metrics are not growing indefinitely.

## Mitigation

- Do not replay events while the database is unavailable.
- Let `event-processor` retry transient failures; it is idempotent through `processed_events`.
- If the database recovers but readiness remains stuck, restart the affected workloads.

```bash
kubectl -n sap-integration-dev rollout restart deploy/sap-integration-platform-event-processor
kubectl -n sap-integration-dev rollout restart deploy/sap-integration-platform-query-api
```

## Follow-Up

- Review Kafka consumer lag because processing stops while the database is unavailable.
- Verify `processed_events` and read model counts after recovery.
- Capture Cloud SQL operation details in the incident notes.
