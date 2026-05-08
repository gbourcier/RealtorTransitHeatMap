#!/usr/bin/env bash
# Full workspace setup: database + Rust build.
# Intended for fresh checkouts or when you want a clean rebuild from scratch.

set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> setting up database..."
./scripts/setup-db.sh

echo "==> cargo fetch..."
cargo fetch

echo "==> cargo build..."
cargo build

echo "==> workspace ready."
