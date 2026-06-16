#!/usr/bin/env bash
# Full workspace setup for a fresh checkout:
#   1. activate tracked git hooks
#   2. install required Go tooling (golang-migrate)
#   3. download Go module dependencies
#   4. install frontend npm dependencies
#   5. build frontend assets for go:embed
#   6. set up the database (drop + create + migrate + seed)
#   7. build all Go packages
#   8. precompute transit stops from GTFS feeds
#
# Idempotent: safe to re-run.

set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not installed. Install Go (>= 1.25) and re-run." >&2
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

echo "==> installing frontend dependencies..."
if ! command -v npm >/dev/null 2>&1; then
    echo "error: npm is not installed. Install Node.js (>= 20) and re-run." >&2
    exit 1
fi
npm --prefix web install

echo "==> building frontend assets..."
npm --prefix web run build

echo "==> setting up database..."
./scripts/setup-db.sh

echo "==> go build..."
go build ./...

echo "==> precomputing transit stops (downloads GTFS feeds on first run)..."
PRECOMPUTE_FLAGS=""
if [[ -d data/gtfs ]] && compgen -G "data/gtfs/*.zip" > /dev/null; then
    PRECOMPUTE_FLAGS="-skip-download"
fi
go run ./cmd/gtfs-precompute $PRECOMPUTE_FLAGS

echo "==> workspace ready."
