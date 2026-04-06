#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

require_command docker
require_command curl

log "Starting local stack from $COMPOSE_FILE"
compose up --build -d

log "Ensuring Kafka topics exist"
"$SCRIPT_DIR/create-topics.sh"

log "Restarting event-processor after local topic bootstrap"
compose restart event-processor >/dev/null

log "Waiting for service health endpoints"
wait_for_http "$SAP_MOCK_API_URL/health" "sap-mock-api"
wait_for_http "$INGESTION_API_URL/health" "ingestion-api"
wait_for_http "$EVENT_PROCESSOR_URL/health" "event-processor"
wait_for_http "$QUERY_API_URL/health" "query-api"

cat <<SUMMARY

Local stack is ready.

Service URLs:
- sap-mock-api:    $SAP_MOCK_API_URL
- ingestion-api:  $INGESTION_API_URL
- event-processor:$EVENT_PROCESSOR_URL
- query-api:      $QUERY_API_URL
- Kafka UI:       $KAFKA_UI_URL
- Prometheus:     $PROMETHEUS_URL
- Grafana:        $GRAFANA_URL

Next useful commands:
- make seed
- make smoke
- make kafka-test
- make logs SERVICE=query-api
SUMMARY
