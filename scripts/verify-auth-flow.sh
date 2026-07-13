#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.auth-verify.yml"
PROJECT_NAME=acmhot100-auth-verify

cleanup() {
  status=$?
  if [[ $status -ne 0 ]]; then
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" logs api mailpit >&2 || true
  fi
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
  return "$status"
}
trap cleanup EXIT

cleanup
docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d --wait
for _ in $(seq 1 180); do
  if curl -fsS http://127.0.0.1:18080/api/v1/health >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS http://127.0.0.1:18080/api/v1/health >/dev/null

python3 "$ROOT_DIR/scripts/verify-auth-flow.py" \
  --api http://127.0.0.1:18080/api/v1 \
  --mailpit http://127.0.0.1:18025
