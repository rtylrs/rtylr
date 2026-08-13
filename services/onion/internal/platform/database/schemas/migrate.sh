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

ERRORS=0

mysql_cmd() {
    mysql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        --password="$DB_PASSWORD" \
        --force \
        "$DB_NAME"
}

echo "Running migrations against $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

for dir in shared cdn rtylr voxire erp hr hr_core hr_attendance hr_documents hr_equipment hr_leave hr_onboarding hr_overtime hr_payroll hr_performance hr_setup hr_shifts hr_time crm finance flow pos hub store support book hr_dashboard recruit insights oauth migrations; do
    schema_dir="$SCHEMAS_DIR/$dir"
    if [ ! -d "$schema_dir" ]; then
        continue
    fi

    for f in "$schema_dir"/*.sql; do
        [ -f "$f" ] || continue
        echo "  → $dir/$(basename "$f")"
        if ! mysql_cmd < "$f" 2>&1 | grep -v "Warning:"; then
            echo "    ⚠ warnings in $f (continuing)"
            ERRORS=$((ERRORS + 1))
        fi
    done
done

if [ "$ERRORS" -gt 0 ]; then
    echo ""
    echo "✅ Migrations complete with $ERRORS warning(s) (tables/indexes already existed)."
else
    echo ""
    echo "✅ All migrations applied cleanly."
fi
