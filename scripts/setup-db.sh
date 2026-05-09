#!/usr/bin/env bash
# Bring up the dev database from a clean slate:
#   1. start (or refresh) the docker compose db service and wait for healthcheck
#   2. drop + recreate the database
#   3. run all golang-migrate migrations
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

# Make go-installed tools (e.g. migrate) reachable when invoked directly.
if command -v go >/dev/null 2>&1; then
    export PATH="$(go env GOPATH)/bin:$PATH"
fi

for tool in docker psql migrate; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "error: required tool '$tool' is not on PATH." >&2
        if [[ "$tool" == "migrate" ]]; then
            echo "       install with:" >&2
            echo "         go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest" >&2
            echo "       or run ./scripts/setup-workspace.sh which installs it for you." >&2
        fi
        exit 1
    fi
done

echo "==> docker compose up..."
docker compose up -d --wait

echo "==> dropping + recreating database..."
PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -p "${POSTGRES_PORT:-5432}" -U "$POSTGRES_USER" -d postgres \
    -v ON_ERROR_STOP=1 \
    -c "DROP DATABASE IF EXISTS \"$POSTGRES_DB\";" \
    -c "CREATE DATABASE \"$POSTGRES_DB\";"

SSL_PARAM=""
if [[ "${POSTGRES_ENABLE_SSL:-true}" == "false" ]]; then
    SSL_PARAM="?sslmode=disable"
fi
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}/${POSTGRES_DB}${SSL_PARAM}"

echo "==> running migrations..."
migrate -path migrations -database "$DB_URL" up

if [[ "${POSTGRES_SEED_DB:-false}" == "true" ]]; then
    echo "==> seeding..."
    psql "$DB_URL" -f seeds/seed.sql
else
    echo "==> skipping seed (POSTGRES_SEED_DB is not true)"
fi

echo "==> done."
