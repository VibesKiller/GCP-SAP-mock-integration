#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/microk8s.sh"

K8S_NAMESPACE="${K8S_NAMESPACE:-sap-integration-local}"
SMOKE_POD_NAME="${SMOKE_POD_NAME:-sap-integration-smoke}"
SAP_SERVICE="${SAP_SERVICE:-sap-integration-platform-sap-mock-api}"
QUERY_SERVICE="${QUERY_SERVICE:-sap-integration-platform-query-api}"
customer_id="CUST-100045"
order_id="SO-2026-000184"
invoice_id="INV-2026-000091"

run_in_smoke_pod() {
  k8s_kubectl -n "$K8S_NAMESPACE" exec "$SMOKE_POD_NAME" -- sh -ec "$1"
}

log "ensuring the application deployments are available in MicroK8s"
k8s_kubectl -n "$K8S_NAMESPACE" get deploy "$SAP_SERVICE" "$QUERY_SERVICE" >/dev/null

log "starting a disposable curl pod inside ${K8S_NAMESPACE}"
k8s_kubectl -n "$K8S_NAMESPACE" delete pod "$SMOKE_POD_NAME" --ignore-not-found >/dev/null
k8s_kubectl -n "$K8S_NAMESPACE" run "$SMOKE_POD_NAME" \
  --image=curlimages/curl:8.7.1 \
  --restart=Never \
  --command \
  -- sh -c 'sleep 300' >/dev/null
k8s_kubectl -n "$K8S_NAMESPACE" wait --for=condition=Ready pod/"$SMOKE_POD_NAME" --timeout=180s >/dev/null
trap 'k8s_kubectl -n "$K8S_NAMESPACE" delete pod "$SMOKE_POD_NAME" --ignore-not-found >/dev/null 2>&1 || true' EXIT

log "verifying health endpoints through the cluster services"
run_in_smoke_pod "curl -fsS http://${SAP_SERVICE}/health >/dev/null"
run_in_smoke_pod "curl -fsS http://sap-integration-platform-ingestion-api/ready >/dev/null"
run_in_smoke_pod "curl -fsS http://sap-integration-platform-event-processor/ready >/dev/null"
run_in_smoke_pod "curl -fsS http://${QUERY_SERVICE}/health >/dev/null"

log "sending business events through sap-mock-api"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/customers/update?dispatch=true' >/dev/null"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/sales-orders/create?dispatch=true' >/dev/null"
run_in_smoke_pod "curl -fsS -X POST 'http://${SAP_SERVICE}/api/v1/simulations/invoices/issue?dispatch=true' >/dev/null"

log "waiting for the event processor to persist the projections"
sleep 8

log "querying the read model through query-api"
run_in_smoke_pod "curl -fsS http://${QUERY_SERVICE}/api/v1/customers/${customer_id} | grep -q '${customer_id}'"
run_in_smoke_pod "curl -fsS http://${QUERY_SERVICE}/api/v1/orders/${order_id} | grep -q '${order_id}'"
run_in_smoke_pod "curl -fsS http://${QUERY_SERVICE}/api/v1/invoices/${invoice_id} | grep -q '${invoice_id}'"

log "MicroK8s smoke test passed"
