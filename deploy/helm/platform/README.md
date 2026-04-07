# Helm Chart: SAP Integration Platform

This chart packages the Go services for Kubernetes and GKE without attempting to manage external platform dependencies such as Kafka or PostgreSQL.

## What The Chart Deploys

- Namespace creation when requested
- ServiceAccount per workload
- ConfigMap per workload for non-sensitive runtime configuration
- Deployment per enabled workload
- ClusterIP Service per enabled HTTP workload
- Ingress for `sap-mock-api`, `ingestion-api` and `query-api`
- HPA for services with autoscaling enabled
- Optional egress-only NetworkPolicy baseline
- Optional namespace-local Prometheus and Grafana for GKE/dev demos

## External Dependencies

This chart assumes the following already exist outside the chart boundary:

- Kafka cluster reachable from the application namespace
- PostgreSQL instance reachable from the application namespace
- Kubernetes Secrets containing sensitive values such as `POSTGRES_URL`
- a secret delivery mechanism such as External Secrets, Secret Sync or CSI if Google Secret Manager is the source of truth
- Artifact Registry images already built and pushed

## Kafka Assumptions

Kafka is intentionally treated as an external dependency.

Configurable values:

- `global.kafka.bootstrapServers`
- `global.kafka.authMode`
- `global.kafka.tls.enabled`
- `global.kafka.tls.serverName`
- `global.kafka.gcp.accessTokenScope`
- `global.kafka.topics.salesOrders`
- `global.kafka.topics.customers`
- `global.kafka.topics.invoices`
- `global.kafka.topics.dlq`
- `global.kafka.consumerGroups.eventProcessor`

Application dependency flow:

- `ingestion-api` writes business topics
- `event-processor` reads business topics and writes the DLQ when required
- `query-api` has no Kafka dependency
- `sap-mock-api` only depends on `ingestion-api`

For Google Managed Kafka with the current Go implementation:

- set `global.kafka.authMode=google_access_token`
- set `global.kafka.tls.enabled=true`
- set `services.ingestionApi.kafkaPrincipalEmail` to the Google service account email used by `ingestion-api`
- set `services.eventProcessor.kafkaPrincipalEmail` to the Google service account email used by `event-processor`

This keeps local Docker Compose development on plaintext Kafka while enabling GKE workloads to authenticate with refreshed Google access tokens obtained through Application Default Credentials.

## Namespace Assumption

The default logical namespace is `sap-integration`, but this is overridden by `values-dev.yaml` and `values-prod.yaml`.

The chart supports either:

- `global.namespace.create=true` to let Helm create it
- `global.namespace.create=false` when the namespace is managed elsewhere

The chart is written with the assumption that this platform runs in a dedicated namespace. That keeps the optional NetworkPolicy focused and makes operational ownership easier to explain.

## Required Secrets

The chart does not create application secrets.

At minimum, create a secret containing `POSTGRES_URL` for:

- `event-processor`
- `query-api`

Example:

```bash
kubectl -n sap-integration-dev create secret generic sap-integration-postgres-dev \
  --from-literal=url='postgres://query_user:replace-with-url-encoded-password@postgres.example.internal:5432/integration?sslmode=require'
```

In a GKE setup that uses Google Secret Manager, the expected pattern is:

1. Terraform provisions the Secret Manager secret and IAM bindings
2. a secret sync mechanism exposes the value as a Kubernetes Secret in the application namespace
3. this chart references the Kubernetes Secret by name only

## Install On Dev

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-dev \
  --create-namespace \
  -f deploy/helm/platform/values-dev.yaml
```

## Install On Prod

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-prod \
  --create-namespace \
  -f deploy/helm/platform/values-prod.yaml
```

## Local Cluster Note

For a local Kubernetes cluster, the quickest path is to disable ingress and use port-forwarding:

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-dev \
  --create-namespace \
  -f deploy/helm/platform/values-dev.yaml \
  --set global.ingress.enabled=false
```

Then access query-api with:

```bash
kubectl -n sap-integration-dev port-forward svc/sap-integration-platform-query-api 8083:80
```

If the Kafka bootstrap address differs from the GKE overlays, override it explicitly:

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-dev \
  --create-namespace \
  -f deploy/helm/platform/values-dev.yaml \
  --set global.ingress.enabled=false \
  --set global.kafka.bootstrapServers='kafka.kafka.svc.cluster.local:9092'
```

## NetworkPolicy Note

`global.networkPolicy` is an egress baseline, not a full zero-trust implementation.

Use it when the Kafka and PostgreSQL network ranges are known. In GKE, populate the correct private CIDRs for:

- Kafka listeners
- PostgreSQL endpoint
- optionally additional control-plane integrations

## Optional Observability On GKE

The chart can deploy a compact Prometheus + Grafana stack for the application namespace.

This is enabled in `values-dev.yaml` and disabled in `values-prod.yaml` by default:

```yaml
observability:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
```

The Prometheus instance scrapes the four Go services through their ClusterIP Services and loads the alert rules packaged under `deploy/helm/platform/files/alerts/`.

Grafana provisions:

- datasource: `Prometheus`
- folder: `SAP Integration Platform`
- dashboard: `SAP Integration Platform Overview`

Access Grafana after a GKE dev deployment with:

```bash
make gke-grafana
```

Then open:

```text
http://localhost:3000/d/sap-integration-platform-overview/sap-integration-platform-overview
```

For production, keep `observability.enabled=false` unless this namespace-local stack is intentionally approved. A shared organization-wide observability stack is usually a better production default.
