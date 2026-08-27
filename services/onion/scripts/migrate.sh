#!/bin/bash
# scripts/migrate.sh
# Run all schema SQL files in order against the configured database.
# Each file is recorded in schema_migrations once applied — re-runs are skipped.
#
# Required environment variables:
#   DB_HOST     - Database host (default: localhost)
#   DB_PORT     - Database port (default: 3306)
#   DB_USER     - Database user
#   DB_PASSWORD - Database password
#   DB_NAME     - Database name

set -eo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:?DB_USER is required}"
DB_PASSWORD="${DB_PASSWORD:?DB_PASSWORD is required}"
DB_NAME="${DB_NAME:?DB_NAME is required}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCHEMAS_DIR="$PROJECT_ROOT/internal/platform/database/schemas"
# The schemas repository owns the canonical dependency order used by both
# migration entry points.
# shellcheck source=../internal/platform/database/schemas/migration_order.sh
source "$SCHEMAS_DIR/migration_order.sh"

mysql_cmd() {
    # The CLI otherwise negotiates latin1, which cannot represent the emoji
    # defaults used by flow/schema.sql and other schemas.
    MYSQL_PWD="$DB_PASSWORD" mysql \
        --default-character-set=utf8mb4 \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        "$@" \
        "$DB_NAME"
}

echo "Running migrations against $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

# Bootstrap the tracking table on first run.
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
    mysql_cmd -e "INSERT IGNORE INTO schema_migrations (filename) VALUES ('$filename');" 2>/dev/null
}

APPLIED=0
SKIPPED=0
FAILED=0
OUTPUT_FILE="$(mktemp)"
trap 'rm -f "$OUTPUT_FILE"' EXIT

deferred_migrations=(
    "hr_time/zz_flow_client_key_tenant_scope.sql"
    "shared/zz_role_access_v2_integrity.sql"
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

apply_migration() {
    local filename="$1"
    local file_path="$2"

    if already_applied "$filename"; then
        echo "  ✓ $filename (already applied)"
        SKIPPED=$((SKIPPED + 1))
        return
    fi

    echo "  → $filename"
    if mysql_cmd < "$file_path" >"$OUTPUT_FILE" 2>&1; then
        # Keep affected-row output and server warnings visible to the operator.
        # Warnings are part of the migration evidence and must not be hidden.
        cat "$OUTPUT_FILE"
    else
        local status=$?
        cat "$OUTPUT_FILE"
        echo "    ✗ $filename failed; it was not recorded as applied" >&2
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
}

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

        apply_migration "$filename" "$f"
    done
done

# Module-owned migrations live beside their sqlc input. Apply them after the
# canonical schema set so module changes can build on shared tables.
shopt -s nullglob
for f in "$PROJECT_ROOT"/internal/modules/*/sql/migration_*.sql; do
    apply_migration "${f#"$PROJECT_ROOT/"}" "$f"
done

# Apply cross-schema migrations after all owning schemas exist. Their
# application-level preflight checks still fail closed on invalid data.
for filename in "${deferred_migrations[@]}"; do
    file_path="$SCHEMAS_DIR/$filename"
    if [ -f "$file_path" ]; then
        apply_migration "$filename" "$file_path"
    fi
done

echo ""
echo "✅ Migrations complete. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
