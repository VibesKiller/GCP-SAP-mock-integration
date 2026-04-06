#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

require_command curl
require_command docker

wait_for_growth() {
  local label="$1"
  local before="$2"
  local command="$3"
  local attempts="${4:-30}"

  for ((i = 1; i <= attempts; i++)); do
    local current
    current="$(eval "$command")"
    if [[ "$current" =~ ^[0-9]+$ ]] && (( current > before )); then
      log "$label increased from $before to $current"
      return 0
    fi
    sleep 2
  done

  echo "$label did not increase after smoke actions" >&2
  return 1
}

assert_contains() {
  local body="$1"
  local expected="$2"
  local label="$3"
  if ! grep -q "$expected" <<<"$body"; then
    echo "$label did not contain expected value: $expected" >&2
    exit 1
  fi
}

CUSTOMER_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/customer-update.json"
ORDER_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/sales-order-create.json"
INVOICE_SAMPLE="$REPO_ROOT/services/sap-mock-api/sample-data/invoice-issued.json"

CUSTOMER_ID="$(extract_json_string "$CUSTOMER_SAMPLE" customer_id)"
ORDER_ID="$(extract_json_string "$ORDER_SAMPLE" sales_document_id)"
INVOICE_ID="$(extract_json_string "$INVOICE_SAMPLE" billing_document_id)"

log "Waiting for local services"
wait_for_http "$SAP_MOCK_API_URL/health" "sap-mock-api"
wait_for_http "$INGESTION_API_URL/health" "ingestion-api"
wait_for_http "$EVENT_PROCESSOR_URL/health" "event-processor"
wait_for_http "$QUERY_API_URL/health" "query-api"

log "Ensuring topics exist before smoke run"
"$SCRIPT_DIR/create-topics.sh" >/dev/null

CUSTOMERS_OFFSET_BEFORE="$(topic_offset_sum sap.customers.v1)"
ORDERS_OFFSET_BEFORE="$(topic_offset_sum sap.sales-orders.v1)"
INVOICES_OFFSET_BEFORE="$(topic_offset_sum sap.invoices.v1)"
PROCESSED_EVENTS_BEFORE="$(postgres_query 'select count(*) from processed_events;')"

log "Executing seed flow"
"$SCRIPT_DIR/seed-local.sh" >/dev/null

wait_for_growth "customer topic offsets" "$CUSTOMERS_OFFSET_BEFORE" "topic_offset_sum sap.customers.v1"
wait_for_growth "sales-order topic offsets" "$ORDERS_OFFSET_BEFORE" "topic_offset_sum sap.sales-orders.v1"
wait_for_growth "invoice topic offsets" "$INVOICES_OFFSET_BEFORE" "topic_offset_sum sap.invoices.v1"
wait_for_growth "processed_events rows" "$PROCESSED_EVENTS_BEFORE" "postgres_query 'select count(*) from processed_events;'"

log "Verifying read model API"
customer_body="$(curl -fsS "$QUERY_API_URL/api/v1/customers/$CUSTOMER_ID")"
order_body="$(curl -fsS "$QUERY_API_URL/api/v1/orders/$ORDER_ID")"
invoice_body="$(curl -fsS "$QUERY_API_URL/api/v1/invoices/$INVOICE_ID")"

assert_contains "$customer_body" "\"customer_id\":\"$CUSTOMER_ID\"" "customer query"
assert_contains "$order_body" "\"order_id\":\"$ORDER_ID\"" "order query"
assert_contains "$invoice_body" "\"invoice_id\":\"$INVOICE_ID\"" "invoice query"

log "Smoke test completed successfully"
printf '\nVerified flow:\n'
printf '  1. health endpoints reachable\n'
printf '  2. mock SAP event dispatch accepted\n'
printf '  3. Kafka topic offsets increased\n'
printf '  4. event processor persisted processed_events\n'
printf '  5. query-api returned customer/order/invoice projections\n'
