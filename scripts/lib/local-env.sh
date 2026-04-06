#!/usr/bin/env bash

LOCAL_ENV_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$LOCAL_ENV_DIR/../.." && pwd)"

ENV_FILE="${ENV_FILE:-$REPO_ROOT/.env}"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

COMPOSE_FILE="${COMPOSE_FILE:-$REPO_ROOT/deploy/local/docker-compose.yml}"
LOCAL_KAFKA_SERVICE="${LOCAL_KAFKA_SERVICE:-kafka}"
LOCAL_POSTGRES_SERVICE="${LOCAL_POSTGRES_SERVICE:-postgres}"
LOCAL_K8S_NODE_IP="${LOCAL_K8S_NODE_IP:-127.0.0.1}"
LOCAL_KAFKA_BOOTSTRAP_SERVER="${LOCAL_KAFKA_BOOTSTRAP_SERVER:-localhost:9092}"
KAFKA_K8S_PORT="${KAFKA_K8S_PORT:-9094}"
LOCAL_KAFKA_K8S_BOOTSTRAP_SERVER="${LOCAL_KAFKA_K8S_BOOTSTRAP_SERVER:-${LOCAL_K8S_NODE_IP}:${KAFKA_K8S_PORT}}"
LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER="${LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER:-kafka:29092}"
POSTGRES_DB="${POSTGRES_DB:-integration}"
POSTGRES_USER="${POSTGRES_USER:-integration}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-integration}"
LOCAL_POSTGRES_URL="${LOCAL_POSTGRES_URL:-postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable}"
SAP_MOCK_API_URL="${SAP_MOCK_API_URL:-http://localhost:8080}"
INGESTION_API_URL="${INGESTION_API_URL:-http://localhost:8081}"
EVENT_PROCESSOR_URL="${EVENT_PROCESSOR_URL:-http://localhost:8082}"
QUERY_API_URL="${QUERY_API_URL:-http://localhost:8083}"
KAFKA_UI_URL="${KAFKA_UI_URL:-http://localhost:8085}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

log() {
  printf '==> %s\n' "$*"
}

warn() {
  printf 'warning: %s\n' "$*" >&2
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "required command not found: $1" >&2
    exit 1
  fi
}

wait_for_http() {
  local url="$1"
  local name="$2"
  local attempts="${3:-60}"
  local sleep_seconds="${4:-2}"

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      log "$name is ready at $url"
      return 0
    fi
    sleep "$sleep_seconds"
  done

  echo "$name did not become ready at $url" >&2
  return 1
}

extract_json_string() {
  local file="$1"
  local key="$2"
  sed -n "s/.*\"$key\": \"\([^\"]*\)\".*/\1/p" "$file" | head -n 1
}

postgres_query() {
  local sql="$1"
  compose exec -T "$LOCAL_POSTGRES_SERVICE" \
    psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atqc "$sql"
}

topic_offset_sum() {
  local topic="$1"
  compose exec -T "$LOCAL_KAFKA_SERVICE" \
    kafka-get-offsets --bootstrap-server "$LOCAL_KAFKA_INTERNAL_BOOTSTRAP_SERVER" --topic "$topic" 2>/dev/null \
    | awk -F: '{sum += $3} END {print sum + 0}'
}
