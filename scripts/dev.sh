#!/usr/bin/env bash
# Run the full dev stack: postgres (docker), backend (air), frontend (vite).
# Ctrl-C stops backend + frontend; the db container is left running.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> starting postgres..."
docker compose up -d --wait

trap 'trap - INT TERM EXIT; kill 0 2>/dev/null || true; wait 2>/dev/null || true' INT TERM EXIT

air &
npm --prefix web run dev &

wait -n
