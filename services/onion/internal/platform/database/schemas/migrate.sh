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
SKIPPED=0
FAILED=0
OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$OUTPUT_FILE"' EXIT

deferred_migrations=(
    "hr_time/zz_flow_client_key_tenant_scope.sql"
    "shared/zz_role_access_v2_integrity.sql"
    "_perf/005_perf_indexes.sql"
)

is_deferred() {
    local filename="$1"
    local deferred
    for deferred in "${deferred_migrations[@]}"; do
        if [ "$filename" = "$deferred" ]; then
            return 0
        fi
    done
    return 1
}

mysql_cmd() {
    MYSQL_PWD="$DB_PASSWORD" mysql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        "$@" \
        "$DB_NAME"
}

mysql_cmd <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename VARCHAR(255) NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
SQL

already_applied() {
    local filename="$1"
    local count
    count=$(mysql_cmd -sN -e "SELECT COUNT(*) FROM schema_migrations WHERE filename = '$filename';" 2>/dev/null)
    [ "$count" -gt 0 ]
}

record_applied() {
    local filename="$1"
    mysql_cmd -e "INSERT INTO schema_migrations (filename) VALUES ('$filename');" 2>/dev/null
}

echo "Running migrations against $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

for dir in "${SCHEMA_MIGRATION_DIRS[@]}"; do
    schema_dir="$SCHEMAS_DIR/$dir"
    if [ ! -d "$schema_dir" ]; then
        continue
    fi

    for f in "$schema_dir"/*.sql; do
        [ -f "$f" ] || continue
        filename="$dir/$(basename "$f")"
        if is_deferred "$filename"; then
            continue
        fi
        if already_applied "$filename"; then
            echo "  ✓ $filename (already applied)"
            SKIPPED=$((SKIPPED + 1))
            continue
        fi
        echo "  → $filename"
        if mysql_cmd < "$f" >"$OUTPUT_FILE" 2>&1; then
            cat "$OUTPUT_FILE"
        else
            status=$?
            cat "$OUTPUT_FILE"
            echo "    ✗ $f failed; migration run stopped" >&2
            FAILED=$((FAILED + 1))
            echo ""
            echo "❌ Migrations stopped. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
            exit "$status"
        fi
        if ! record_applied "$filename"; then
            echo "    ✗ could not record $filename as applied" >&2
            FAILED=$((FAILED + 1))
            echo ""
            echo "❌ Migrations stopped. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
            exit 1
        fi
        APPLIED=$((APPLIED + 1))
    done
done

for filename in "${deferred_migrations[@]}"; do
    file_path="$SCHEMAS_DIR/$filename"
    if [ ! -f "$file_path" ]; then
        continue
    fi
    if already_applied "$filename"; then
        echo "  ✓ $filename (already applied)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi
    echo "  → $filename"
    if mysql_cmd < "$file_path" >"$OUTPUT_FILE" 2>&1; then
        cat "$OUTPUT_FILE"
    else
        status=$?
        cat "$OUTPUT_FILE"
        echo "    ✗ $file_path failed; migration run stopped" >&2
        FAILED=$((FAILED + 1))
        echo ""
        echo "❌ Migrations stopped. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
        exit "$status"
    fi
    if ! record_applied "$filename"; then
        echo "    ✗ could not record $filename as applied" >&2
        FAILED=$((FAILED + 1))
        echo ""
        echo "❌ Migrations stopped. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
        exit 1
    fi
    APPLIED=$((APPLIED + 1))
done

echo ""
echo "✅ Migrations complete. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
