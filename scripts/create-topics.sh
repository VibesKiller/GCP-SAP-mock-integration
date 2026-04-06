#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

require_command docker

TOPICS=(
  sap.sales-orders.v1
  sap.customers.v1
  sap.invoices.v1
  sap.integration.dlq.v1
)

declare -A PARTITIONS=(
  [sap.sales-orders.v1]=6
  [sap.customers.v1]=3
  [sap.invoices.v1]=3
  [sap.integration.dlq.v1]=3
)

log "Using Kafka bootstrap server $LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER"
ready=false
for attempt in $(seq 1 30); do
  if compose exec -T "$LOCAL_KAFKA_SERVICE" cub kafka-ready -b localhost:9092 1 30 >/dev/null 2>&1; then
    ready=true
    break
  fi
  sleep 2
done

if [[ "$ready" != "true" ]]; then
  echo "Kafka broker did not become ready in time" >&2
  exit 1
fi

for topic in "${TOPICS[@]}"; do
  compose exec -T "$LOCAL_KAFKA_SERVICE" kafka-topics \
    --bootstrap-server "$LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER" \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions "${PARTITIONS[$topic]}" \
    --replication-factor 1 >/dev/null
  log "ensured topic $topic"
done
