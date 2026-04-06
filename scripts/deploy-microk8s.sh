#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/microk8s.sh"

K8S_NAMESPACE="${K8S_NAMESPACE:-sap-integration-local}"
HELM_RELEASE="${HELM_RELEASE:-sap-integration-platform}"
POSTGRES_SECRET_NAME="${POSTGRES_SECRET_NAME:-sap-integration-postgres-local}"
MICROK8S_IMAGE_TAG="${MICROK8S_IMAGE_TAG:-microk8s}"
KAFKA_K8S_PORT="${KAFKA_K8S_PORT:-9094}"
services=(
  sap-mock-api
  ingestion-api
  event-processor
  query-api
)
tmp_files=()

cleanup() {
  if [[ ${#tmp_files[@]} -gt 0 ]]; then
    rm -f "${tmp_files[@]}"
  fi
}

trap cleanup EXIT

require_command docker
require_command curl
require_command microk8s

log "detecting the MicroK8s node IP"
node_ip="$(detect_microk8s_node_ip)"
if [[ -z "$node_ip" ]]; then
  echo "could not detect MicroK8s node IP" >&2
  exit 1
fi
export LOCAL_K8S_NODE_IP="$node_ip"
export LOCAL_KAFKA_K8S_BOOTSTRAP_SERVER="${LOCAL_K8S_NODE_IP}:${KAFKA_K8S_PORT}"

log "ensuring local Kafka and PostgreSQL are running with a Kubernetes-accessible Kafka listener"
compose up -d --force-recreate kafka postgres

log "waiting for Kafka and PostgreSQL health checks"
for attempt in {1..40}; do
  kafka_ready=false
  postgres_ready=false

  if compose exec -T kafka cub kafka-ready -b localhost:9092 1 5 >/dev/null 2>&1; then
    kafka_ready=true
  fi

  if compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    postgres_ready=true
  fi

  if [[ "$kafka_ready" == true && "$postgres_ready" == true ]]; then
    break
  fi
  sleep 3
done

if [[ "${kafka_ready:-false}" != true || "${postgres_ready:-false}" != true ]]; then
  echo "local Kafka or PostgreSQL did not become healthy" >&2
  exit 1
fi

log "creating Kafka topics if they are missing"
"$SCRIPT_DIR/create-topics.sh"

for service in "${services[@]}"; do
  log "building Docker image ${service}:${MICROK8S_IMAGE_TAG}"
  docker build -t "${service}:${MICROK8S_IMAGE_TAG}" -f "services/${service}/Dockerfile" .

  image_archive="$(mktemp --suffix=.tar)"
  tmp_files+=("$image_archive")

  log "exporting ${service}:${MICROK8S_IMAGE_TAG} as an OCI archive"
  docker save -o "$image_archive" "${service}:${MICROK8S_IMAGE_TAG}"

  log "importing ${service}:${MICROK8S_IMAGE_TAG} into MicroK8s"
  microk8s_exec images import "$image_archive"

  rm -f "$image_archive"
done

postgres_url="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${LOCAL_K8S_NODE_IP}:${POSTGRES_PORT:-5432}/${POSTGRES_DB}?sslmode=disable"

log "creating namespace ${K8S_NAMESPACE} if needed"
if ! k8s_kubectl get namespace "$K8S_NAMESPACE" >/dev/null 2>&1; then
  k8s_kubectl create namespace "$K8S_NAMESPACE"
fi

secret_manifest="$(mktemp)"
tmp_files+=("$secret_manifest")
k8s_kubectl -n "$K8S_NAMESPACE" create secret generic "$POSTGRES_SECRET_NAME" \
  --from-literal=url="$postgres_url" \
  --dry-run=client \
  -o yaml >"$secret_manifest"
k8s_kubectl apply -f "$secret_manifest"

log "deploying the Helm release ${HELM_RELEASE} into ${K8S_NAMESPACE}"
k8s_helm upgrade --install "$HELM_RELEASE" "$REPO_ROOT/deploy/helm/platform" \
  --namespace "$K8S_NAMESPACE" \
  --create-namespace \
  -f "$REPO_ROOT/deploy/helm/platform/values.yaml" \
  -f "$REPO_ROOT/deploy/helm/platform/values-microk8s.yaml" \
  --set global.kafka.bootstrapServers="$LOCAL_KAFKA_K8S_BOOTSTRAP_SERVER" \
  --wait \
  --timeout 5m

log "waiting for deployments to become ready"
k8s_kubectl -n "$K8S_NAMESPACE" rollout status deployment/"$HELM_RELEASE"-sap-mock-api --timeout=180s
k8s_kubectl -n "$K8S_NAMESPACE" rollout status deployment/"$HELM_RELEASE"-ingestion-api --timeout=180s
k8s_kubectl -n "$K8S_NAMESPACE" rollout status deployment/"$HELM_RELEASE"-event-processor --timeout=180s
k8s_kubectl -n "$K8S_NAMESPACE" rollout status deployment/"$HELM_RELEASE"-query-api --timeout=180s

log "MicroK8s deployment is ready"
k8s_kubectl -n "$K8S_NAMESPACE" get deploy,svc,pods
printf '\n'
printf 'MicroK8s namespace: %s\n' "$K8S_NAMESPACE"
printf 'Kafka bootstrap for cluster workloads: %s\n' "$LOCAL_KAFKA_K8S_BOOTSTRAP_SERVER"
printf 'PostgreSQL for cluster workloads: %s\n' "$postgres_url"
