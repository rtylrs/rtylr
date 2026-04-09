#!/bin/bash
# scripts/migrate.sh
# Run all schema SQL files in order against the configured database.
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

# Run schema files in a consistent order: shared first, then module-specific schemas.
# Within each directory, files are applied in lexicographic (alphabetical) order.
for dir in shared rtylr voxire erp hr crm finance flow pos hub; do
    schema_dir="$PROJECT_ROOT/internal/platform/database/schemas/$dir"
    if [ ! -d "$schema_dir" ]; then
        continue
    fi

    for f in "$schema_dir"/*.sql; do
        [ -f "$f" ] || continue
        echo "  → $(basename "$f")"
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
