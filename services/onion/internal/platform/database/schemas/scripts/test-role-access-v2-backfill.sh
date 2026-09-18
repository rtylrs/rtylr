#!/usr/bin/env bash

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
: "${DB_USER:?DB_USER is required}"
: "${DB_PASSWORD:?DB_PASSWORD is required}"

DB_NAME="schemas_role_access_v2_backfill_test_$$"
OUTPUT_FILE="$(mktemp)"
created=0

mysql_cmd() {
    MYSQL_PWD="$DB_PASSWORD" mysql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        "$@"
}

cleanup() {
    local status=$?
    rm -f "$OUTPUT_FILE"
    if ((created == 1)) && ! mysql_cmd -e "DROP DATABASE \`$DB_NAME\`"; then
        echo "failed to drop scratch database $DB_NAME" >&2
        if ((status == 0)); then
            status=1
        fi
    fi
    return "$status"
}
trap cleanup EXIT

if mysql_cmd -Nse "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '$DB_NAME'" | grep -qx "$DB_NAME"; then
    echo "scratch database already exists: $DB_NAME" >&2
    exit 1
fi

mysql_cmd -e "CREATE DATABASE \`$DB_NAME\`"
created=1

for schema_file in \
    shared/users.sql \
    shared/businesses.sql \
    shared/user_groups.sql \
    shared/user_apps.sql \
    shared/zz_role_access_v2.sql; do
    mysql_cmd "$DB_NAME" < "$schema_file"
done

mysql_cmd "$DB_NAME" <<'SQL'
INSERT INTO users (id, email, password_hash)
VALUES
    (1001, 'role-backfill-owner@example.com', 'test'),
    (1002, 'role-backfill-other@example.com', 'test');

INSERT INTO businesses (id, name, owner_id)
VALUES
    (2001, 'Role Backfill Owner Business', 1001),
    (2002, 'Role Backfill Other Business', 1002);

INSERT INTO user_app_permissions
    (id, business_id, owner_id, user_id, apps, role, is_active)
VALUES
    (3001, 2002, 1001, 1001, '["flow"]', 'owner', TRUE);
SQL

if mysql_cmd "$DB_NAME" < shared/zz_role_access_v2_backfill.sql >"$OUTPUT_FILE" 2>&1; then
    echo "backfill unexpectedly accepted a cross-owner business mapping" >&2
    cat "$OUTPUT_FILE" >&2
    exit 1
fi

if ! grep -q 'role_access_v2_backfill_abort' "$OUTPUT_FILE"; then
    echo "backfill failed for an unexpected reason" >&2
    cat "$OUTPUT_FILE" >&2
    exit 1
fi

counts=$(mysql_cmd "$DB_NAME" -Nse "SELECT CONCAT(
    (SELECT COUNT(*) FROM authorization_business_members), ':',
    (SELECT COUNT(*) FROM authorization_license_assignments), ':',
    (SELECT COUNT(*) FROM authorization_role_assignments)
)")
if [[ "$counts" != "0:0:0" ]]; then
    echo "failed backfill created normalized rows: $counts" >&2
    exit 1
fi

mysql_cmd "$DB_NAME" <<'SQL'
UPDATE user_app_permissions
SET business_id = 2001
WHERE id = 3001;
SQL

mysql_cmd "$DB_NAME" < shared/zz_role_access_v2_backfill.sql >"$OUTPUT_FILE" 2>&1

counts=$(mysql_cmd "$DB_NAME" -Nse "SELECT CONCAT(
    (SELECT COUNT(*) FROM authorization_business_members WHERE business_id = 2001 AND user_id = 1001), ':',
    (SELECT COUNT(*) FROM authorization_license_assignments WHERE business_id = 2001 AND user_id = 1001 AND app_id = 'flow'), ':',
    (SELECT COUNT(*) FROM authorization_role_assignments WHERE business_id = 2001 AND user_id = 1001 AND app_id = 'flow')
)")
if [[ "$counts" != "1:1:1" ]]; then
    echo "valid backfill produced unexpected normalized row counts: $counts" >&2
    exit 1
fi

echo "role access V2 backfill ownership regression passed"
