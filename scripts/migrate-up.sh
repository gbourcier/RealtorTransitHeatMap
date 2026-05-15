#!/usr/bin/env bash
# Apply all pending golang-migrate migrations.

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
    echo "error: .env not found. Copy .env.example to .env first." >&2
    exit 1
fi

set -a
source .env
set +a

if command -v go >/dev/null 2>&1; then
    export PATH="$(go env GOPATH)/bin:$PATH"
fi

if ! command -v migrate >/dev/null 2>&1; then
    echo "error: 'migrate' is not on PATH." >&2
    echo "       install with:" >&2
    echo "         go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest" >&2
    exit 1
fi

SSL_PARAM=""
if [[ "${POSTGRES_ENABLE_SSL:-true}" == "false" ]]; then
    SSL_PARAM="?sslmode=disable"
fi
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}/${POSTGRES_DB}${SSL_PARAM}"

echo "==> applying pending migrations..."
migrate -path migrations -database "$DB_URL" up
echo "==> done."
