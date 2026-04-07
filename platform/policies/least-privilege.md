# Least Privilege Baseline

The platform is designed with least privilege as a default expectation.

## Principles

- Each workload gets a dedicated Kubernetes service account.
- Workload Identity is preferred over static cloud credentials.
- Secret access is scoped per service and per environment.
- PostgreSQL users are separated by runtime responsibility when query and write paths diverge.
- Kafka ACLs are provisioned per producer, topic and consumer group in the Terraform-managed Kafka layer.

## GCP Mapping

- `ingestion-api`: write access to `sap.sales-orders.v1`, `sap.customers.v1` and `sap.invoices.v1`, plus read access to its own secrets.
- `event-processor`: read business topics, write `sap.integration.dlq.v1` and write PostgreSQL projections.
- `query-api`: read-only access to PostgreSQL, no Kafka write permissions.
- `notification-worker`: read selected business topics and outbound secrets only.
