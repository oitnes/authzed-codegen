#!/usr/bin/env bash
set -euo pipefail

GRPC_HEALTH_PROBE_VERSION="${GRPC_HEALTH_PROBE_VERSION:-v0.4.28}"
INSTALL_DIR="${GRPC_HEALTH_PROBE_INSTALL_DIR:-/usr/local/bin}"

if command -v grpc_health_probe &>/dev/null; then
  echo "grpc_health_probe already installed: $(command -v grpc_health_probe)"
  exit 0
fi

if [[ "${CI:-false}" != "true" ]]; then
  echo "Not in CI and grpc_health_probe is not installed — skipping"
  exit 0
fi

wget -qO "${INSTALL_DIR}/grpc_health_probe" \
  "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64"
chmod +x "${INSTALL_DIR}/grpc_health_probe"
echo "grpc_health_probe ${GRPC_HEALTH_PROBE_VERSION} installed to ${INSTALL_DIR}"
