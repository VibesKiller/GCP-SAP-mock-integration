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
CONSUMER_GROUPS=(
  sap-integration.event-processor.v1
)

log "Kafka bootstrap servers"
printf '  external: %s\n' "$LOCAL_KAFKA_BOOTSTRAP_SERVER"
printf '  internal: %s\n' "$LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER"

log "Describing local Kafka topics"
for topic in "${TOPICS[@]}"; do
  printf '\n[%s]\n' "$topic"
  compose exec -T "$LOCAL_KAFKA_SERVICE" kafka-topics \
    --bootstrap-server "$LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER" \
    --describe \
    --topic "$topic"
done

log "Current topic end offsets"
for topic in "${TOPICS[@]}"; do
  printf '  %-24s %s\n' "$topic" "$(topic_offset_sum "$topic")"
done

log "Consumer group inspection"
for group in "${CONSUMER_GROUPS[@]}"; do
  printf '\n[%s]\n' "$group"
  compose exec -T "$LOCAL_KAFKA_SERVICE" kafka-consumer-groups \
    --bootstrap-server "$LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER" \
    --describe \
    --group "$group" || true
done
