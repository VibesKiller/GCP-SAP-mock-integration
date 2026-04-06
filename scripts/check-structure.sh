#!/usr/bin/env bash
set -euo pipefail

required_paths=(
  .env.example
  README.md
  Makefile
  go.mod
  api/openapi/platform-api.yaml
  api/openapi/ingestion-api.yaml
  api/openapi/query-api.yaml
  deploy/local/docker-compose.yml
  deploy/helm/platform/Chart.yaml
  platform/kafka/topic-catalog.yaml
  platform/kafka/consumer-groups.yaml
  platform/database/migrations/001_initial_schema.sql
  docs/architecture.md
  docs/runbooks/local-development.md
  services/sap-mock-api/main.go
  services/ingestion-api/main.go
  services/event-processor/main.go
  services/query-api/main.go
  scripts/up-local.sh
  scripts/create-topics.sh
  scripts/seed-local.sh
  scripts/smoke-test.sh
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
