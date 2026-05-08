#!/usr/bin/env bash
# Bring up the dev database from a clean slate:
#   1. start (or refresh) the docker compose db service and wait for healthcheck
#   2. drop + recreate the database
#   3. run all sqlx migrations
#   4. apply seed data
#
# Safe to re-run; intended for local dev only.

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
    echo "error: .env not found. Copy .env.example to .env first." >&2
    exit 1
fi

set -a
source .env
set +a

echo "==> docker compose up..."
docker compose up -d --wait

echo "==> sqlx database reset (drop + create + migrate)..."
sqlx database reset -y

echo "==> seeding..."
psql "$DATABASE_URL" -f seeds/seed.sql

echo "==> done."
