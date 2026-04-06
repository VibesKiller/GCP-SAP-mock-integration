#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

require_command curl

log "Seeding realistic SAP business events through sap-mock-api"

run_seed_call() {
  local label="$1"
  local url="$2"
  log "$label"
  curl -fsS -X POST "$url"
  printf '\n'
}

run_seed_call "customer.updated" "$SAP_MOCK_API_URL/api/v1/simulations/customers/update?dispatch=true"
run_seed_call "sales_order.created" "$SAP_MOCK_API_URL/api/v1/simulations/sales-orders/create?dispatch=true"
run_seed_call "sales_order.updated" "$SAP_MOCK_API_URL/api/v1/simulations/sales-orders/update?dispatch=true"
run_seed_call "invoice.issued" "$SAP_MOCK_API_URL/api/v1/simulations/invoices/issue?dispatch=true"

log "Seed flow completed"
