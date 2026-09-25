#!/usr/bin/env bash
# Snapshot and restore a database around the additive Role and Access Control
# V2 migration. This utility is intentionally limited to local scratch
# databases because restore replaces the selected database.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-onion_root_password}"
DB_NAME="${DB_NAME:-onion_test}"
CONTAINER="${ONION_DB_CONTAINER:-onion-env-db-1}"

V2_TABLES=(
    authorization_permission_catalog
    authorization_roles
    authorization_role_permissions
    authorization_business_members
    authorization_license_assignments
    authorization_role_assignments
    authorization_role_assignment_scopes
    authorization_dash_app_defaults
    authorization_audit_events
)

V2_MIGRATIONS=(
    shared/zz_role_access_v2.sql
    shared/zz_role_access_v2_backfill.sql
    shared/zz_role_access_v2_email_verification.sql
    shared/zz_role_access_v2_integrity.sql
    shared/zz_role_access_v2_provisioning.sql
    shared/zz_role_access_v2_usernames.sql
)

die() {
    echo "role access rollback: $1" >&2
    exit 1
}

if ! [[ "$DB_NAME" =~ ^[A-Za-z0-9_]+$ ]]; then
    die "DB_NAME contains characters that are unsafe for a database identifier"
fi
case "$DB_NAME" in
    *_test|*_scratch) ;;
    *) die "refusing to operate on non-scratch database '$DB_NAME'" ;;
esac
case "$DB_HOST" in
    127.0.0.1|localhost|::1|db|onion-env-db-1) ;;
    *) die "refusing to operate on remote database host '$DB_HOST'" ;;
esac

mysql_mode=""
if command -v mysql >/dev/null 2>&1 && command -v mysqldump >/dev/null 2>&1; then
    mysql_mode="host"
elif docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
    mysql_mode="container"
else
    die "mysql and mysqldump are unavailable, and database container '$CONTAINER' is not running"
fi

mysql_exec() {
    if [ "$mysql_mode" = "host" ]; then
        MYSQL_PWD="$DB_PASSWORD" mysql \
            --default-character-set=utf8mb4 \
            --host="$DB_HOST" --port="$DB_PORT" --user="$DB_USER" \
            "$DB_NAME" "$@"
    else
        docker exec -i -e "MYSQL_PWD=$DB_PASSWORD" "$CONTAINER" mysql \
            --default-character-set=utf8mb4 -uroot "$DB_NAME" "$@"
    fi
}

mysql_admin() {
    if [ "$mysql_mode" = "host" ]; then
        MYSQL_PWD="$DB_PASSWORD" mysql \
            --default-character-set=utf8mb4 \
            --host="$DB_HOST" --port="$DB_PORT" --user="$DB_USER" "$@"
    else
        docker exec -i -e "MYSQL_PWD=$DB_PASSWORD" "$CONTAINER" mysql \
            --default-character-set=utf8mb4 -uroot "$@"
    fi
}

dump_exec() {
    if [ "$mysql_mode" = "host" ]; then
        MYSQL_PWD="$DB_PASSWORD" mysqldump \
            --default-character-set=utf8mb4 \
            --host="$DB_HOST" --port="$DB_PORT" --user="$DB_USER" \
            --single-transaction --routines --triggers --events --hex-blob \
            --set-gtid-purged=OFF --no-tablespaces "$DB_NAME"
    else
        docker exec -i -e "MYSQL_PWD=$DB_PASSWORD" "$CONTAINER" mysqldump \
            --default-character-set=utf8mb4 -uroot \
            --single-transaction --routines --triggers --events --hex-blob \
            --set-gtid-purged=OFF --no-tablespaces "$DB_NAME"
    fi
}

v2_table_count() {
    mysql_exec -N -e "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN ('authorization_permission_catalog','authorization_roles','authorization_role_permissions','authorization_business_members','authorization_license_assignments','authorization_role_assignments','authorization_role_assignment_scopes','authorization_dash_app_defaults','authorization_audit_events');"
}

v2_migration_count() {
    mysql_exec -N -e "SELECT COUNT(*) FROM schema_migrations WHERE filename IN ('shared/zz_role_access_v2.sql','shared/zz_role_access_v2_backfill.sql','shared/zz_role_access_v2_email_verification.sql','shared/zz_role_access_v2_integrity.sql','shared/zz_role_access_v2_provisioning.sql','shared/zz_role_access_v2_usernames.sql');"
}

legacy_signature() {
    mysql_exec -N -e "SELECT CONCAT((SELECT COUNT(*) FROM businesses), ':', (SELECT COUNT(*) FROM users), ':', (SELECT COUNT(*) FROM user_app_permissions), ':', COALESCE((SELECT SHA2(GROUP_CONCAT(CONCAT_WS('#', id, owner_id, user_id, COALESCE(CAST(business_id AS CHAR), ''), CAST(apps AS CHAR), role, IF(is_active, '1', '0')) ORDER BY id SEPARATOR '|'), 256) FROM user_app_permissions), 'none'));"
}

snapshot_database() {
    local snapshot_file="$1"
    local table_count migration_count signature metadata_file

    [ -n "$snapshot_file" ] || die "snapshot path is required"
    [ -d "$(dirname "$snapshot_file")" ] || die "snapshot directory does not exist"

    table_count="$(v2_table_count)"
    migration_count="$(v2_migration_count)"
    if [ "$table_count" -ne 0 ] || [ "$migration_count" -ne 0 ]; then
        die "database is not pre-migration: found $table_count V2 tables and $migration_count V2 migration records"
    fi

    signature="$(legacy_signature)"
    dump_exec > "$snapshot_file"
    metadata_file="${snapshot_file}.meta"
    printf '%s\n' "$signature" > "$metadata_file"
    echo "snapshot created: $snapshot_file"
}

restore_database() {
    local snapshot_file="$1"
    local metadata_file expected_signature actual_signature table_count migration_count

    [ -f "$snapshot_file" ] || die "snapshot file does not exist: $snapshot_file"
    metadata_file="${snapshot_file}.meta"
    [ -f "$metadata_file" ] || die "snapshot metadata does not exist: $metadata_file"
    [ "${ROLE_ACCESS_V2_ROLLBACK_CONFIRM:-}" = "I_UNDERSTAND" ] || die "set ROLE_ACCESS_V2_ROLLBACK_CONFIRM=I_UNDERSTAND to replace the scratch database"

    expected_signature="$(< "$metadata_file")"
    table_count="$(v2_table_count)"
    migration_count="$(v2_migration_count)"
    if [ "$table_count" -ne 9 ] || [ "$migration_count" -ne 6 ]; then
        die "database is not fully post-migration: found $table_count of 9 V2 tables and $migration_count of 6 V2 migration records"
    fi

    mysql_admin -e "DROP DATABASE IF EXISTS \`$DB_NAME\`; CREATE DATABASE \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    mysql_exec < "$snapshot_file"

    actual_signature="$(legacy_signature)"
    [ "$actual_signature" = "$expected_signature" ] || die "legacy authorization signature changed during restore"
    table_count="$(v2_table_count)"
    migration_count="$(v2_migration_count)"
    if [ "$table_count" -ne 0 ] || [ "$migration_count" -ne 0 ]; then
        die "restore retained $table_count V2 tables and $migration_count V2 migration records"
    fi
    echo "scratch database restored: $DB_NAME"
}

if [ "$#" -ne 2 ]; then
    echo "usage: $0 snapshot|restore SNAPSHOT_FILE" >&2
    exit 2
fi

case "$1" in
    snapshot) snapshot_database "$2" ;;
    restore) restore_database "$2" ;;
    *) die "unknown action '$1'" ;;
esac
