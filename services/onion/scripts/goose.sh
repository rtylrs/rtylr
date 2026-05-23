#!/bin/bash
# scripts/goose.sh
# Run Goose migrations. If goose is not installed locally, it runs via 'go run'.
#
# Usage:
#   ./scripts/goose.sh status
#   ./scripts/goose.sh up
#   ./scripts/goose.sh down
#   ./scripts/goose.sh create add_some_table sql

set -eo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:?DB_USER is required}"
DB_PASSWORD="${DB_PASSWORD:?DB_PASSWORD is required}"
DB_NAME="${DB_NAME:?DB_NAME is required}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/internal/platform/database/migrations"

mkdir -p "$MIGRATIONS_DIR"

DB_DSN="${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"

# Check if goose binary is installed, otherwise use go run
if command -v goose &> /dev/null; then
    goose -dir "$MIGRATIONS_DIR" mysql "$DB_DSN" "$@"
else
    echo "Goose is not installed locally, running via 'go run'..."
    go run github.com/pressly/goose/v3/cmd/goose@latest -dir "$MIGRATIONS_DIR" mysql "$DB_DSN" "$@"
fi
