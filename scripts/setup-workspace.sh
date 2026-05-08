#!/usr/bin/env bash
# Full workspace setup for a fresh checkout:
#   1. activate tracked git hooks
#   2. install required Go tooling (golang-migrate)
#   3. download Go module dependencies
#   4. set up the database (drop + create + migrate + seed)
#   5. build all Go packages
#
# Idempotent: safe to re-run.

set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not installed. Install Go (>= 1.24) and re-run." >&2
    exit 1
fi

# Make sure tools installed via `go install` are on PATH for the rest of this run.
export PATH="$(go env GOPATH)/bin:$PATH"

echo "==> activating tracked git hooks..."
git config core.hooksPath .githooks

echo "==> ensuring golang-migrate is installed..."
if ! command -v migrate >/dev/null 2>&1; then
    echo "    installing github.com/golang-migrate/migrate/v4/cmd/migrate ..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

echo "==> go mod download..."
go mod download

echo "==> setting up database..."
./scripts/setup-db.sh

echo "==> go build..."
go build ./...

echo "==> workspace ready."
