#!/usr/bin/env bash
set -euo pipefail

required_paths=(
  .env.example
  .github/workflows/ci.yml
  .github/workflows/docker-images.yml
  .github/workflows/deploy-dev.yml
  .github/workflows/deploy-prod.yml
  README.md
  Makefile
  go.mod
  api/openapi/platform-api.yaml
  api/openapi/ingestion-api.yaml
  api/openapi/query-api.yaml
  deploy/local/docker-compose.yml
  deploy/helm/platform/Chart.yaml
  deploy/helm/platform/README.md
  deploy/helm/platform/values.yaml
  deploy/helm/platform/values-dev.yaml
  deploy/helm/platform/values-microk8s.yaml
  deploy/helm/platform/values-prod.yaml
  deploy/helm/platform/files/alerts/platform-alerts.yaml
  deploy/helm/platform/files/dashboards/platform-overview.json
  deploy/helm/platform/templates/observability-configmaps.yaml
  deploy/helm/platform/templates/observability-deployments.yaml
  deploy/helm/platform/templates/observability-services.yaml
  platform/kafka/topic-catalog.yaml
  platform/kafka/consumer-groups.yaml
  platform/database/migrations/001_initial_schema.sql
  platform/monitoring/prometheus.yml
  platform/alerts/platform-alerts.yaml
  platform/dashboards/platform-overview.json
  docs/architecture.md
  docs/decisions/adr-0001-kafka-as-event-backbone.md
  docs/decisions/adr-0002-postgresql-for-read-models.md
  docs/decisions/adr-0003-terraform-for-gcp-infrastructure.md
  docs/decisions/adr-0004-go-for-runtime-services.md
  docs/decisions/adr-0005-gke-for-kubernetes-runtime.md
  docs/decisions/adr-0006-helm-for-workload-packaging.md
  docs/decisions/adr-0007-github-actions-for-ci-cd.md
  docs/diagrams/architecture.mmd
  docs/diagrams/event-flow.mmd
  docs/runbooks/local-development.md
  docs/runbooks/kubernetes-deployment.md
  docs/runbooks/kafka-broker-down.md
  docs/runbooks/database-down.md
  docs/runbooks/dead-letter-topic-growth.md
  docs/runbooks/consumer-lag.md
  docs/runbooks/dlq-replay.md
  docs/runbooks/ci-cd.md
  services/sap-mock-api/main.go
  services/ingestion-api/main.go
  services/event-processor/main.go
  services/query-api/main.go
  scripts/up-local.sh
  scripts/create-topics.sh
  scripts/deploy-gke.sh
  scripts/deploy-microk8s.sh
  scripts/gke-port-forward.sh
  scripts/seed-local.sh
  scripts/smoke-test.sh
  scripts/smoke-test-gke.sh
  scripts/smoke-test-microk8s.sh
  terraform/envs/dev/main.tf
  terraform/envs/prod/main.tf
)

for path in "${required_paths[@]}"; do
  if [[ ! -e "$path" ]]; then
    echo "missing required path: $path" >&2
    exit 1
  fi
done

echo "repository structure looks complete"
