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

mysql_cmd() {
    MYSQL_PWD="$DB_PASSWORD" mysql \
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

for dir in shared rtylr voxire erp hr hr_core hr_attendance hr_documents hr_equipment hr_leave hr_onboarding hr_overtime hr_payroll hr_performance hr_setup hr_shifts hr_time crm finance flow pos hub store support book hr_dashboard recruit insights oauth ai migrations; do
    schema_dir="$SCHEMAS_DIR/$dir"
    if [ ! -d "$schema_dir" ]; then
        continue
    fi

    for f in "$schema_dir"/*.sql; do
        [ -f "$f" ] || continue
        filename="$dir/$(basename "$f")"

        if already_applied "$filename"; then
            echo "  ✓ $filename (already applied)"
            SKIPPED=$((SKIPPED + 1))
            continue
        fi

        echo "  → $filename"
        if mysql_cmd < "$f" >"$OUTPUT_FILE" 2>&1; then
            grep -v "Warning:" "$OUTPUT_FILE" || true
        else
            status=$?
            grep -v "Warning:" "$OUTPUT_FILE" || true
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
    done
done

echo ""
echo "✅ Migrations complete. Applied: $APPLIED, Skipped: $SKIPPED, Failed: $FAILED."
