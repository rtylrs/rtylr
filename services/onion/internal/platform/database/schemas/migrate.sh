#!/bin/bash
# migrate.sh — apply all Voxire schemas in dependency order.
#
# Required env vars:
#   DB_HOST     (default: localhost)
#   DB_PORT     (default: 3306)
#   DB_USER
#   DB_PASSWORD
#   DB_NAME

set -eo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:?DB_USER is required}"
DB_PASSWORD="${DB_PASSWORD:?DB_PASSWORD is required}"
DB_NAME="${DB_NAME:?DB_NAME is required}"

SCHEMAS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Keep this runner and onion/scripts/migrate.sh on the same schema coverage and
# dependency order.
# shellcheck source=migration_order.sh
source "$SCHEMAS_DIR/migration_order.sh"

APPLIED=0
FAILED=0
OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$OUTPUT_FILE"' EXIT

mysql_cmd() {
    mysql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        --password="$DB_PASSWORD" \
        "$DB_NAME"
}

echo "Running migrations against $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

for dir in "${SCHEMA_MIGRATION_DIRS[@]}"; do
    schema_dir="$SCHEMAS_DIR/$dir"
    if [ ! -d "$schema_dir" ]; then
        continue
    fi

    for f in "$schema_dir"/*.sql; do
        [ -f "$f" ] || continue
        echo "  → $dir/$(basename "$f")"
        if mysql_cmd < "$f" >"$OUTPUT_FILE" 2>&1; then
            grep -v "Warning:" "$OUTPUT_FILE" || true
            APPLIED=$((APPLIED + 1))
        else
            status=$?
            grep -v "Warning:" "$OUTPUT_FILE" || true
            echo "    ✗ $f failed; migration run stopped" >&2
            FAILED=$((FAILED + 1))
            echo ""
            echo "❌ Migrations stopped. Applied: $APPLIED, Skipped: 0, Failed: $FAILED."
            exit "$status"
        fi
    done
done

echo ""
echo "✅ Migrations complete. Applied: $APPLIED, Skipped: 0, Failed: $FAILED."
