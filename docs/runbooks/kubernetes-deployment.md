# Kubernetes Deployment Runbook

This runbook explains how to deploy the SAP integration platform to Kubernetes or GKE using the Helm chart under `deploy/helm/platform`.

## Scope

The chart deploys application workloads only. It does not provision:

- GKE clusters
- Kafka brokers
- PostgreSQL
- Artifact Registry repositories
- Kubernetes Secret objects

Those dependencies are intentionally managed outside the chart boundary.

## Namespace Assumption

The recommended operating model is one dedicated namespace per environment:

- `sap-integration-dev`
- `sap-integration-prod`

This keeps ownership clear, makes NetworkPolicy simpler and avoids mixing integration workloads with unrelated applications.

## Prerequisites

Before installing the chart, ensure:

1. the target cluster already exists
2. application container images have been pushed to Artifact Registry
3. Kafka is reachable from the application namespace
4. PostgreSQL is reachable from the application namespace
5. Kubernetes Secrets containing `POSTGRES_URL` already exist
6. if Workload Identity is used, the target Google service accounts and IAM bindings are already provisioned

## Kafka Dependency Model

Kafka is treated as a shared platform dependency and can live outside the application namespace.

The relevant Helm values are:

- `global.kafka.bootstrapServers`
- `global.kafka.authMode`
- `global.kafka.tls.enabled`
- `global.kafka.gcp.accessTokenScope`
- `global.kafka.topics.salesOrders`
- `global.kafka.topics.customers`
- `global.kafka.topics.invoices`
- `global.kafka.topics.dlq`
- `global.kafka.consumerGroups.eventProcessor`

Application responsibilities:

- `ingestion-api` publishes business events to Kafka
- `event-processor` consumes those topics and produces to the DLQ when needed
- `query-api` does not connect to Kafka
- `sap-mock-api` does not connect to Kafka directly

For Google Managed Kafka, the current Go implementation uses:

- TLS enabled
- SASL/PLAIN over Kafka using a short-lived Google access token refreshed from ADC
- one principal email per Kafka-enabled workload

Required Helm values for that path:

- `global.kafka.authMode=google_access_token`
- `global.kafka.tls.enabled=true`
- `services.ingestionApi.kafkaPrincipalEmail=<ingestion-api-gsa>`
- `services.eventProcessor.kafkaPrincipalEmail=<event-processor-gsa>`

For non-GCP Kafka clusters that require SASL/PLAIN, inject `KAFKA_SASL_PASSWORD` through the `kafkaSASLPasswordSecret` values rather than putting the password in a ConfigMap.

## Required Secrets

The chart references existing Kubernetes Secrets rather than embedding sensitive values.

Minimum expected secret:

- `POSTGRES_URL` for `event-processor`
- `POSTGRES_URL` for `query-api`

Example:

```bash
kubectl -n sap-integration-dev create secret generic sap-integration-postgres-dev \
  --from-literal=url='postgres://query_user:replace-with-url-encoded-password@postgres.example.internal:5432/integration?sslmode=require'
```

In GKE, a professional pattern is to keep the secret in Google Secret Manager and sync it into Kubernetes using an approved mechanism. The Helm chart stays intentionally neutral and only references the resulting Kubernetes Secret.

## Ingress Model

The default GKE overlays use host-based routing:

- `sap-mock-api-dev.example.internal`
- `ingestion-api-dev.example.internal`
- `query-api-dev.example.internal`

Production follows the same pattern with production hostnames under the example domain. This avoids relying on path rewrites because the Go services expose native paths such as `/api/v1/...`, `/health`, `/ready` and `/metrics`.

## Deploy To Dev

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-dev \
  --create-namespace \
  -f deploy/helm/platform/values.yaml \
  -f deploy/helm/platform/values-dev.yaml
```

## Deploy To Prod

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-prod \
  --create-namespace \
  -f deploy/helm/platform/values.yaml \
  -f deploy/helm/platform/values-prod.yaml
```

## Render Before Deploy

Use these commands locally before pushing a chart change:

```bash
make helm-template
make helm-template-dev
make helm-template-prod
make helm-lint
```

If `helm` is not installed, these targets fail fast with a clear message.

## Local Kubernetes Note

For a local cluster such as `kind` or `minikube`, disable ingress unless you actually want to manage an ingress controller:

```bash
helm upgrade --install sap-integration-platform deploy/helm/platform \
  --namespace sap-integration-dev \
  --create-namespace \
  -f deploy/helm/platform/values.yaml \
  -f deploy/helm/platform/values-dev.yaml \
  --set global.ingress.enabled=false \
  --set global.kafka.bootstrapServers='kafka.kafka.svc.cluster.local:9092'
```

Then access the query API with:

```bash
kubectl -n sap-integration-dev port-forward svc/sap-integration-platform-query-api 8083:80
```

## MicroK8s Local Path

This repository includes a dedicated MicroK8s workflow for a persistent local Kubernetes environment.

Files involved:

- `deploy/helm/platform/values-microk8s.yaml`
- `scripts/deploy-microk8s.sh`
- `scripts/smoke-test-microk8s.sh`

Execution model:

1. local Docker Compose provides Kafka and PostgreSQL
2. Kafka is recreated with a Kubernetes-accessible listener
3. the Go service images are built locally
4. the images are imported into MicroK8s
5. Helm deploys the workloads into `sap-integration-local`
6. an in-cluster smoke test validates the end-to-end path

Use:

```bash
make microk8s-deploy
make microk8s-smoke
```

## Configurable Values

Most operational tuning is exposed through Helm values:

- image repository, tag and pull policy per service
- replica counts
- resource requests and limits
- HPA thresholds for `ingestion-api`
- readiness and liveness probes
- service account annotations for Workload Identity
- Kafka bootstrap servers, topics and consumer group
- ingress class, hosts, annotations and TLS
- namespace creation strategy
- optional egress NetworkPolicy rules

## NetworkPolicy Guidance

The included NetworkPolicy is intentionally minimal.

It allows:

- DNS resolution through `kube-dns`
- egress to explicitly configured Kafka CIDRs and ports
- egress to explicitly configured PostgreSQL CIDRs and ports

This is a pragmatic baseline for a dedicated application namespace, not a full zero-trust policy set.

## Post-Deployment Checks

```bash
kubectl -n sap-integration-dev get deploy,svc,ingress,hpa
kubectl -n sap-integration-dev get pods
kubectl -n sap-integration-dev describe deploy sap-integration-platform-ingestion-api
kubectl -n sap-integration-dev logs deploy/sap-integration-platform-event-processor --tail=100
```

Check that:

1. all pods become ready
2. the ingress address is allocated if enabled
3. the `event-processor` can resolve Kafka and PostgreSQL endpoints
4. `query-api` returns healthy responses through the Service or Ingress

## GKE Smoke Test

After Terraform has provisioned GCP, deploy the workloads and run the GKE smoke test from the repository root:

```bash
make gke-deploy
make gke-status
make gke-smoke
```

`make gke-deploy` uses Terraform outputs to:

1. configure `kubectl` for the GKE cluster
2. build and push service images to Artifact Registry
3. create the application namespace
4. create the `POSTGRES_URL` Kubernetes Secret from Secret Manager and Cloud SQL outputs
5. run the PostgreSQL migration from a temporary in-cluster pod
6. deploy the Helm release with the Managed Kafka bootstrap address and Workload Identity annotations

The smoke test uses the current `kubectl` context and starts a temporary `curlimages/curl` pod inside `sap-integration-dev`. It verifies the end-to-end path through internal Kubernetes Services rather than Ingress:

1. rollout status for the four application deployments
2. health and readiness endpoints inside the namespace
3. Prometheus and Grafana readiness when chart observability is enabled
4. Grafana dashboard provisioning when chart observability is enabled
5. `sap-mock-api` dispatch to `ingestion-api`
6. Kafka publish and consume through Google Managed Kafka
7. PostgreSQL projection reads through `query-api`

Useful overrides:

```bash
K8S_NAMESPACE=sap-integration-prod make gke-smoke
HELM_RELEASE=sap-integration-platform make gke-smoke
KUBECTL=/path/to/kubectl make gke-smoke
```

## GKE Grafana Access

The dev overlay enables a small namespace-local Prometheus and Grafana stack:

- Prometheus Service: `sap-integration-platform-prometheus`
- Grafana Service: `sap-integration-platform-grafana`
- Dashboard UID: `sap-integration-platform-overview`

Port-forward Grafana:

```bash
make gke-grafana
```

The target starts a detached port-forward and stores PID/log files under `.cache/port-forward/`, so the terminal returns immediately.

Then open:

```text
http://localhost:3000/d/sap-integration-platform-overview/sap-integration-platform-overview
```

Port-forward Prometheus when you need to inspect targets or alerts:

```bash
make gke-prometheus
```

Check or stop detached port-forwards:

```bash
make gke-observability-status
make gke-observability-stop
```

Useful URLs:

- `http://localhost:9090/targets`
- `http://localhost:9090/alerts`

The Kafka consumer lag panel remains dashboard-ready but requires a Kafka exporter or managed Kafka lag metric scraped by this Prometheus instance.
