#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

ENVIRONMENT="${ENVIRONMENT:-dev}"
TERRAFORM_DIR="${TERRAFORM_DIR:-$REPO_ROOT/terraform/envs/$ENVIRONMENT}"
K8S_NAMESPACE="${K8S_NAMESPACE:-sap-integration-${ENVIRONMENT}}"
HELM_RELEASE="${HELM_RELEASE:-sap-integration-platform}"
POSTGRES_SECRET_NAME="${POSTGRES_SECRET_NAME:-sap-integration-postgres-${ENVIRONMENT}}"
IMAGE_TAG="${IMAGE_TAG:-gke-${ENVIRONMENT}-$(git -C "$REPO_ROOT" rev-parse --short HEAD)}"
DISABLE_INGRESS="${DISABLE_INGRESS:-true}"
MIGRATION_POD_NAME="${MIGRATION_POD_NAME:-sap-integration-db-migrate}"

require_command terraform
require_command gcloud
require_command docker
require_command helm
require_command kubectl
require_command jq

tf_output_raw() {
  terraform -chdir="$TERRAFORM_DIR" output -raw "$1"
}

tf_output_raw_or_default() {
  local output_name="$1"
  local default_value="$2"
  tf_output_raw "$output_name" 2>/dev/null || printf '%s' "$default_value"
}

tf_output_json() {
  terraform -chdir="$TERRAFORM_DIR" output -json "$1"
}

service_account_email() {
  local key="$1"
  tf_output_json service_account_emails | jq -r --arg key "$key" '.[$key]'
}

cleanup_migration_pod() {
  kubectl -n "$K8S_NAMESPACE" delete pod "$MIGRATION_POD_NAME" --ignore-not-found >/dev/null 2>&1 || true
}

PROJECT_ID="$(sed -n 's/^project_id[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$TERRAFORM_DIR/terraform.tfvars" | head -n 1)"
REGION="$(sed -n 's/^region[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$TERRAFORM_DIR/terraform.tfvars" | head -n 1)"
ARTIFACT_REPO_URL="$(tf_output_raw artifact_registry_repository_url)"
CLOUDSQL_IP="$(tf_output_raw cloudsql_private_ip_address)"
CLOUDSQL_DB="$(tf_output_raw_or_default cloudsql_database_name integration)"
CLOUDSQL_USER="$(tf_output_raw_or_default cloudsql_app_username integration_app)"
POSTGRES_SECRET_ID="$(tf_output_json secret_ids | jq -r '.postgresql_app_password')"
POSTGRES_PASSWORD="$(gcloud secrets versions access latest --project "$PROJECT_ID" --secret "$POSTGRES_SECRET_ID")"
POSTGRES_PASSWORD_ENCODED="$(jq -nr --arg value "$POSTGRES_PASSWORD" '$value|@uri')"
POSTGRES_URL="postgres://${CLOUDSQL_USER}:${POSTGRES_PASSWORD_ENCODED}@${CLOUDSQL_IP}:5432/${CLOUDSQL_DB}?sslmode=disable"

SAP_MOCK_API_GSA="$(service_account_email sap-mock-api)"
INGESTION_API_GSA="$(service_account_email ingestion-api)"
EVENT_PROCESSOR_GSA="$(service_account_email event-processor)"
QUERY_API_GSA="$(service_account_email query-api)"

log "configuring kubectl for GKE"
eval "$(tf_output_raw gke_get_credentials_command)"

log "resolving Google Managed Kafka bootstrap address"
KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS:-$(eval "$(tf_output_raw managed_kafka_bootstrap_address_lookup_command)")}"

log "configuring Docker authentication for Artifact Registry"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

log "building and pushing application images with tag ${IMAGE_TAG}"
for service in sap-mock-api ingestion-api event-processor query-api; do
  image="${ARTIFACT_REPO_URL}/${service}:${IMAGE_TAG}"
  docker build -f "$REPO_ROOT/services/${service}/Dockerfile" -t "$image" "$REPO_ROOT"
  docker push "$image"
done

log "creating namespace and PostgreSQL secret"
kubectl create namespace "$K8S_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$K8S_NAMESPACE" create secret generic "$POSTGRES_SECRET_NAME" \
  --from-literal=url="$POSTGRES_URL" \
  --dry-run=client -o yaml | kubectl apply -f -

log "applying database migration from inside the cluster"
cleanup_migration_pod
trap cleanup_migration_pod EXIT
cat <<YAML | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${MIGRATION_POD_NAME}
  namespace: ${K8S_NAMESPACE}
  labels:
    app.kubernetes.io/name: sap-integration-db-migrate
    app.kubernetes.io/managed-by: script
spec:
  restartPolicy: Never
  containers:
    - name: psql
      image: postgres:16
      command: ["sh", "-c", "sleep 600"]
      env:
        - name: POSTGRES_URL
          valueFrom:
            secretKeyRef:
              name: ${POSTGRES_SECRET_NAME}
              key: url
YAML
kubectl -n "$K8S_NAMESPACE" wait --for=condition=Ready "pod/${MIGRATION_POD_NAME}" --timeout=240s >/dev/null
kubectl -n "$K8S_NAMESPACE" cp "$REPO_ROOT/platform/database/migrations/001_initial_schema.sql" "${MIGRATION_POD_NAME}:/tmp/001_initial_schema.sql"
kubectl -n "$K8S_NAMESPACE" exec "$MIGRATION_POD_NAME" -- sh -ec 'psql "$POSTGRES_URL" -f /tmp/001_initial_schema.sql >/dev/null'
cleanup_migration_pod
trap - EXIT

log "rendering GKE-specific Helm values"
VALUES_FILE="$(mktemp)"
INGRESS_ENABLED="true"
if [[ "$DISABLE_INGRESS" == "true" ]]; then
  INGRESS_ENABLED="false"
fi
trap 'rm -f "$VALUES_FILE"' EXIT
cat >"$VALUES_FILE" <<YAML
global:
  projectId: ${PROJECT_ID}
  namespace:
    create: false
    name: ${K8S_NAMESPACE}
  ingress:
    enabled: ${INGRESS_ENABLED}
  kafka:
    bootstrapServers: ${KAFKA_BOOTSTRAP_SERVERS}
    authMode: google_access_token
    tls:
      enabled: true
      insecureSkipVerify: false
services:
  sapMockApi:
    image:
      repository: ${ARTIFACT_REPO_URL}/sap-mock-api
      tag: ${IMAGE_TAG}
    serviceAccount:
      name: sap-mock-api
      annotations:
        iam.gke.io/gcp-service-account: ${SAP_MOCK_API_GSA}
  ingestionApi:
    image:
      repository: ${ARTIFACT_REPO_URL}/ingestion-api
      tag: ${IMAGE_TAG}
    kafkaPrincipalEmail: ${INGESTION_API_GSA}
    serviceAccount:
      name: ingestion-api
      annotations:
        iam.gke.io/gcp-service-account: ${INGESTION_API_GSA}
  eventProcessor:
    image:
      repository: ${ARTIFACT_REPO_URL}/event-processor
      tag: ${IMAGE_TAG}
    kafkaPrincipalEmail: ${EVENT_PROCESSOR_GSA}
    serviceAccount:
      name: event-processor
      annotations:
        iam.gke.io/gcp-service-account: ${EVENT_PROCESSOR_GSA}
    extraSecretEnv:
      - name: POSTGRES_URL
        secretName: ${POSTGRES_SECRET_NAME}
        secretKey: url
        optional: false
  queryApi:
    image:
      repository: ${ARTIFACT_REPO_URL}/query-api
      tag: ${IMAGE_TAG}
    serviceAccount:
      name: query-api
      annotations:
        iam.gke.io/gcp-service-account: ${QUERY_API_GSA}
    extraSecretEnv:
      - name: POSTGRES_URL
        secretName: ${POSTGRES_SECRET_NAME}
        secretKey: url
        optional: false
YAML

log "deploying Helm release ${HELM_RELEASE} to ${K8S_NAMESPACE}"
helm upgrade --install "$HELM_RELEASE" "$REPO_ROOT/deploy/helm/platform" \
  --namespace "$K8S_NAMESPACE" \
  -f "$REPO_ROOT/deploy/helm/platform/values.yaml" \
  -f "$REPO_ROOT/deploy/helm/platform/values-dev.yaml" \
  -f "$VALUES_FILE" \
  --wait \
  --timeout 10m

log "GKE deployment completed"
printf '\nNext checks:\n'
printf '  make gke-status\n'
printf '  make gke-smoke\n'
