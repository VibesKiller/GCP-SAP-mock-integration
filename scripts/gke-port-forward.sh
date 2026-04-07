#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_PATH="$SCRIPT_DIR/$(basename "${BASH_SOURCE[0]}")"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib/local-env.sh"

KUBECTL="${KUBECTL:-kubectl}"
K8S_NAMESPACE="${K8S_NAMESPACE:-sap-integration-dev}"
HELM_RELEASE="${HELM_RELEASE:-sap-integration-platform}"
STATE_DIR="${PORT_FORWARD_STATE_DIR:-$REPO_ROOT/.cache/port-forward}"
TARGET="${1:-status}"

require_command "$KUBECTL"

service_for_target() {
  case "$1" in
    grafana)
      printf '%s %s %s %s\n' "${HELM_RELEASE}-grafana" "3000" "80" "http://localhost:3000/d/sap-integration-platform-overview/sap-integration-platform-overview"
      ;;
    prometheus)
      printf '%s %s %s %s\n' "${HELM_RELEASE}-prometheus" "9090" "80" "http://localhost:9090/targets"
      ;;
    *)
      echo "unknown port-forward target: $1" >&2
      exit 1
      ;;
  esac
}

pid_file() {
  printf '%s/%s.pid' "$STATE_DIR" "$1"
}

log_file() {
  printf '%s/%s.log' "$STATE_DIR" "$1"
}

unit_name() {
  printf 'sap-integration-%s-port-forward.service' "$1"
}

systemd_user_available() {
  command -v systemd-run >/dev/null 2>&1 &&
    command -v systemctl >/dev/null 2>&1 &&
    systemctl --user show-environment >/dev/null 2>&1
}

systemd_unit_pid() {
  local target="$1"
  systemctl --user show "$(unit_name "$target")" \
    --property=MainPID \
    --value 2>/dev/null || true
}

is_running() {
  local target="$1"
  local file pid
  file="$(pid_file "$target")"
  if [[ -f "$file" ]] && kill -0 "$(cat "$file")" >/dev/null 2>&1; then
    return 0
  fi

  if systemd_user_available; then
    pid="$(systemd_unit_pid "$target")"
    if [[ -n "$pid" && "$pid" != "0" ]] && kill -0 "$pid" >/dev/null 2>&1; then
      printf '%s\n' "$pid" >"$file"
      return 0
    fi
  fi

  return 1
}

run_foreground() {
  local target="$1"
  local service local_port remote_port url
  read -r service local_port remote_port url < <(service_for_target "$target")

  exec "$KUBECTL" -n "$K8S_NAMESPACE" \
    port-forward "svc/${service}" "${local_port}:${remote_port}"
}

start_with_systemd() {
  local target="$1"
  local pid log_path unit
  log_path="$(log_file "$target")"
  unit="$(unit_name "$target")"

  systemctl --user stop "$unit" >/dev/null 2>&1 || true
  : >"$log_path"

  if ! systemd-run --user \
    --unit "$unit" \
    --collect \
    --property="StandardOutput=append:${log_path}" \
    --property="StandardError=append:${log_path}" \
    "$SCRIPT_PATH" "run-${target}" >/dev/null 2>&1; then
    return 1
  fi

  sleep 2
  pid="$(systemd_unit_pid "$target")"
  if [[ -n "$pid" && "$pid" != "0" ]] && kill -0 "$pid" >/dev/null 2>&1; then
    printf '%s\n' "$pid" >"$(pid_file "$target")"
    return 0
  fi

  return 1
}

start_forward() {
  local target="$1"
  local service local_port remote_port url pid log_path
  read -r service local_port remote_port url < <(service_for_target "$target")
  mkdir -p "$STATE_DIR"

  if is_running "$target"; then
    pid="$(cat "$(pid_file "$target")")"
    log "${target} port-forward already running with PID ${pid}"
    printf 'URL: %s\n' "$url"
    printf 'Log: %s\n' "$(log_file "$target")"
    return 0
  fi

  "$KUBECTL" -n "$K8S_NAMESPACE" get service "$service" >/dev/null

  log_path="$(log_file "$target")"
  if systemd_user_available && start_with_systemd "$target"; then
    pid="$(cat "$(pid_file "$target")")"
  elif command -v setsid >/dev/null 2>&1; then
    setsid "$KUBECTL" -n "$K8S_NAMESPACE" port-forward "svc/${service}" "${local_port}:${remote_port}" </dev/null >"$log_path" 2>&1 &
    pid="$!"
    printf '%s\n' "$pid" >"$(pid_file "$target")"
  else
    nohup "$KUBECTL" -n "$K8S_NAMESPACE" port-forward "svc/${service}" "${local_port}:${remote_port}" </dev/null >"$log_path" 2>&1 &
    pid="$!"
    printf '%s\n' "$pid" >"$(pid_file "$target")"
  fi

  sleep 2
  if ! kill -0 "$pid" >/dev/null 2>&1; then
    echo "failed to start ${target} port-forward; log follows:" >&2
    cat "$log_path" >&2 || true
    rm -f "$(pid_file "$target")"
    exit 1
  fi
  disown "$pid" >/dev/null 2>&1 || true

  log "${target} port-forward started with PID ${pid}"
  printf 'URL: %s\n' "$url"
  printf 'Log: %s\n' "$log_path"
}

stop_forward() {
  local target="$1"
  local file pid stopped_with_systemd
  file="$(pid_file "$target")"
  stopped_with_systemd=0

  if systemd_user_available; then
    if systemctl --user is-active --quiet "$(unit_name "$target")"; then
      systemctl --user stop "$(unit_name "$target")" >/dev/null 2>&1 || true
      stopped_with_systemd=1
    else
      systemctl --user stop "$(unit_name "$target")" >/dev/null 2>&1 || true
    fi
  fi

  if [[ ! -f "$file" ]]; then
    log "${target} port-forward is not tracked"
    return 0
  fi

  pid="$(cat "$file")"
  if kill -0 "$pid" >/dev/null 2>&1; then
    kill "$pid"
    log "stopped ${target} port-forward with PID ${pid}"
  elif [[ "$stopped_with_systemd" == "1" ]]; then
    log "stopped ${target} port-forward systemd unit"
  else
    log "${target} port-forward PID ${pid} was not running"
  fi
  rm -f "$file"
}

status_forward() {
  local target pid
  for target in grafana prometheus; do
    if is_running "$target"; then
      pid="$(cat "$(pid_file "$target")")"
      log "${target}: running with PID ${pid}"
    else
      log "${target}: stopped"
    fi
  done
}

case "$TARGET" in
  grafana|prometheus)
    start_forward "$TARGET"
    ;;
  stop)
    stop_forward "${2:-grafana}"
    if [[ "${2:-}" == "" ]]; then
      stop_forward prometheus
    fi
    ;;
  stop-grafana)
    stop_forward grafana
    ;;
  stop-prometheus)
    stop_forward prometheus
    ;;
  run-grafana)
    run_foreground grafana
    ;;
  run-prometheus)
    run_foreground prometheus
    ;;
  status)
    status_forward
    ;;
  *)
    echo "usage: $0 {grafana|prometheus|status|stop|stop-grafana|stop-prometheus}" >&2
    exit 1
    ;;
esac
