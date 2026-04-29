#!/usr/bin/env bash
set -euo pipefail

SPICEDB_TOKEN="${SPICEDB_TOKEN:-ci-integration-test-key}"
SPICEDB_PORT="${SPICEDB_PORT:-50051}"
SPICEDB_IMAGE="${SPICEDB_IMAGE:-authzed/spicedb:latest}"
CONTAINER_NAME="spicedb"

# Stop and remove any existing container (no-op if it doesn't exist)
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm   "$CONTAINER_NAME" 2>/dev/null || true

docker run -d --name "$CONTAINER_NAME" \
  -p "${SPICEDB_PORT}:50051" \
  "$SPICEDB_IMAGE" serve \
  --grpc-preshared-key "$SPICEDB_TOKEN" \
  --datastore-engine memory

for i in $(seq 1 30); do
  if grpc_health_probe -addr="127.0.0.1:${SPICEDB_PORT}" 2>/dev/null; then
    echo "SpiceDB is ready"
    exit 0
  fi
  sleep 2
done

docker logs "$CONTAINER_NAME"
exit 1
