#!/usr/bin/env bash

MICROK8S_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$MICROK8S_LIB_DIR/local-env.sh"

microk8s_exec() {
  local cmd=(microk8s "$@")

  if groups | tr ' ' '\n' | grep -qx 'microk8s'; then
    "${cmd[@]}"
    return
  fi

  local quoted=()
  local arg
  for arg in "${cmd[@]}"; do
    quoted+=("$(printf '%q' "$arg")")
  done

  sg microk8s -c "${quoted[*]}"
}

k8s_kubectl() {
  microk8s_exec kubectl "$@"
}

k8s_helm() {
  microk8s_exec helm3 "$@"
}

detect_microk8s_node_ip() {
  k8s_kubectl get nodes -o jsonpath='{range .items[0].status.addresses[?(@.type=="InternalIP")]}{.address}{end}'
}
