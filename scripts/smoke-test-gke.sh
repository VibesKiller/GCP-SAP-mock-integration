#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

KUBECTL="${KUBECTL:-kubectl}"
K8S_NAMESPACE="${K8S_NAMESPACE:-sap-integration-dev}"
HELM_RELEASE="${HELM_RELEASE:-sap-integration-platform}"
SMOKE_POD_NAME="${SMOKE_POD_NAME:-sap-integration-gke-smoke}"
SMOKE_IMAGE="${SMOKE_IMAGE:-curlimages/curl:8.7.1}"

SAP_SERVICE="${SAP_SERVICE:-${HELM_RELEASE}-sap-mock-api}"
INGESTION_SERVICE="${INGESTION_SERVICE:-${HELM_RELEASE}-ingestion-api}"
EVENT_PROCESSOR_SERVICE="${EVENT_PROCESSOR_SERVICE:-${HELM_RELEASE}-event-processor}"
QUERY_SERVICE="${QUERY_SERVICE:-${HELM_RELEASE}-query-api}"
PROMETHEUS_SERVICE="${PROMETHEUS_SERVICE:-${HELM_RELEASE}-prometheus}"
GRAFANA_SERVICE="${GRAFANA_SERVICE:-${HELM_RELEASE}-grafana}"

CUSTOMER_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/customer-update.json"
ORDER_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/sales-order-create.json"
INVOICE_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/invoice-issued.json"

CUSTOMER_ID="${CUSTOMER_ID:-$(extract_json_string "$CUSTOMER_SAMPLE" customer_id)}"
ORDER_ID="${ORDER_ID:-$(extract_json_string "$ORDER_SAMPLE" sales_document_id)}"
INVOICE_ID="${INVOICE_ID:-$(extract_json_string "$INVOICE_SAMPLE" billing_document_id)}"

require_command "$KUBECTL"

kubectl_cmd() {
  "$KUBECTL" "$@"
}

run_in_smoke_pod() {
  kubectl_cmd -n "$K8S_NAMESPACE" exec "$SMOKE_POD_NAME" -- sh -ec "$1"
}

retry_in_smoke_pod() {
  local label="$1"
  local command="$2"
  local attempts="${3:-30}"
  local sleep_seconds="${4:-3}"

  for ((i = 1; i <= attempts; i++)); do
    if run_in_smoke_pod "$command" >/dev/null 2>&1; then
      log "$label verified"
      return 0
    fi
    sleep "$sleep_seconds"
  done

  echo "$label was not verified after $attempts attempts" >&2
  return 1
}

cleanup() {
  kubectl_cmd -n "$K8S_NAMESPACE" delete pod "$SMOKE_POD_NAME" --ignore-not-found >/dev/null 2>&1 || true
}

log "using Kubernetes context: $(kubectl_cmd config current-context)"
log "checking namespace ${K8S_NAMESPACE}"
kubectl_cmd get namespace "$K8S_NAMESPACE" >/dev/null

log "waiting for application rollouts"
for deployment in \
  "${HELM_RELEASE}-sap-mock-api" \
  "${HELM_RELEASE}-ingestion-api" \
  "${HELM_RELEASE}-event-processor" \
  "${HELM_RELEASE}-query-api"; do
  kubectl_cmd -n "$K8S_NAMESPACE" rollout status "deployment/${deployment}" --timeout=240s
done

OBSERVABILITY_ENABLED=false
if kubectl_cmd -n "$K8S_NAMESPACE" get deployment "$PROMETHEUS_SERVICE" >/dev/null 2>&1 &&
  kubectl_cmd -n "$K8S_NAMESPACE" get deployment "$GRAFANA_SERVICE" >/dev/null 2>&1; then
  OBSERVABILITY_ENABLED=true
  log "waiting for observability rollouts"
  kubectl_cmd -n "$K8S_NAMESPACE" rollout status "deployment/${PROMETHEUS_SERVICE}" --timeout=240s
  kubectl_cmd -n "$K8S_NAMESPACE" rollout status "deployment/${GRAFANA_SERVICE}" --timeout=240s
fi

log "starting disposable smoke-test pod"
cleanup
trap cleanup EXIT
kubectl_cmd -n "$K8S_NAMESPACE" run "$SMOKE_POD_NAME" \
  --image="$SMOKE_IMAGE" \
  --restart=Never \
  --command \
  -- sh -c 'sleep 300' >/dev/null
kubectl_cmd -n "$K8S_NAMESPACE" wait --for=condition=Ready "pod/${SMOKE_POD_NAME}" --timeout=240s >/dev/null

log "verifying health and readiness endpoints through ClusterIP services"
run_in_smoke_pod "curl -fsS http://${SAP_SERVICE}/health >/dev/null"
run_in_smoke_pod "curl -fsS http://${INGESTION_SERVICE}/ready >/dev/null"
run_in_smoke_pod "curl -fsS http://${EVENT_PROCESSOR_SERVICE}/ready >/dev/null"
run_in_smoke_pod "curl -fsS http://${QUERY_SERVICE}/health >/dev/null"

if [[ "$OBSERVABILITY_ENABLED" == "true" ]]; then
  log "verifying Prometheus and Grafana endpoints"
  run_in_smoke_pod "curl -fsS http://${PROMETHEUS_SERVICE}/-/ready >/dev/null"
  run_in_smoke_pod "curl -fsS http://${GRAFANA_SERVICE}/api/health >/dev/null"
  retry_in_smoke_pod \
    "Prometheus service scrape targets" \
    "curl -fsS 'http://${PROMETHEUS_SERVICE}/api/v1/query?query=up%7Bjob%3D~%22sap-mock-api%7Cingestion-api%7Cevent-processor%7Cquery-api%22%7D' | grep -q 'sap-mock-api'" \
    20 \
    3
  retry_in_smoke_pod \
    "Grafana dashboard provisioning" \
    "curl -fsS 'http://${GRAFANA_SERVICE}/api/search?query=SAP%20Integration%20Platform' | grep -q 'sap-integration-platform-overview'" \
    20 \
    3
fi

log "sending business events through sap-mock-api"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/customers/update?dispatch=true' >/dev/null"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/sales-orders/create?dispatch=true' >/dev/null"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/invoices/issue?dispatch=true' >/dev/null"

log "waiting for Kafka processing and PostgreSQL projections"
retry_in_smoke_pod "customer read model" "curl -fsS http://${QUERY_SERVICE}/api/v1/customers/${CUSTOMER_ID} | grep -q '${CUSTOMER_ID}'"
retry_in_smoke_pod "order read model" "curl -fsS http://${QUERY_SERVICE}/api/v1/orders/${ORDER_ID} | grep -q '${ORDER_ID}'"
retry_in_smoke_pod "invoice read model" "curl -fsS http://${QUERY_SERVICE}/api/v1/invoices/${INVOICE_ID} | grep -q '${INVOICE_ID}'"

log "GKE smoke test passed"
printf '\nVerified flow:\n'
printf '  1. Kubernetes deployments rolled out\n'
printf '  2. service-to-service health checks passed inside the namespace\n'
printf '  3. sap-mock-api dispatched events to ingestion-api\n'
printf '  4. ingestion-api and event-processor used Kafka successfully\n'
printf '  5. event-processor persisted projections queried by query-api\n'
if [[ "$OBSERVABILITY_ENABLED" == "true" ]]; then
  printf '  6. Prometheus and Grafana were reachable and provisioned\n'
fi
