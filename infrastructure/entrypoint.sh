#!/bin/sh
set -e

: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_HOST:?POSTGRES_HOST is required}"

SSL_PARAM=""
if [ "${POSTGRES_ENABLE_SSL:-false}" = "false" ]; then
    SSL_PARAM="?sslmode=disable"
fi
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT:-5432}/${POSTGRES_DB}${SSL_PARAM}"

echo "==> running migrations"
/app/migrate -path /app/migrations -database "$DB_URL" up

echo "==> starting api on ${HTTP_ADDR:-0.0.0.0:3000}"
exec /app/api
