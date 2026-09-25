#!/usr/bin/env bash
# Exercise the Role and Access Control V2 migration and its scratch restore
# procedure from a database state containing only the legacy authorization.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_PW="${ONION_DB_ROOT_PW:-onion_root_password}"
DB_NAME="onion_test"
CONTAINER="${ONION_DB_CONTAINER:-onion-env-db-1}"

export DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME
export ONION_DB_ROOT_PW="$ROOT_PW"

case "$DB_NAME" in
    *_test|*_scratch) ;;
    *) echo "rollback check requires a scratch database" >&2; exit 1 ;;
esac

TEST_DB_SETUP_ONLY=1 "$PROJECT_ROOT/scripts/test-db.sh"

dbq() {
    docker exec -i -e "MYSQL_PWD=$ROOT_PW" "$CONTAINER" mysql -uroot "$DB_NAME" "$@"
}

dbq <<'SQL'
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS
    authorization_role_assignment_scopes,
    authorization_role_permissions,
    authorization_role_assignments,
    authorization_license_assignments,
    authorization_business_members,
    authorization_dash_app_defaults,
    authorization_audit_events,
    authorization_roles,
    authorization_permission_catalog,
    user_mfa_enrollments,
    user_provisioning_states,
    user_email_verification_tokens;
SET FOREIGN_KEY_CHECKS = 1;
DELETE FROM schema_migrations
WHERE filename IN (
    'shared/zz_role_access_v2.sql',
    'shared/zz_role_access_v2_backfill.sql',
    'shared/zz_role_access_v2_email_verification.sql',
    'shared/zz_role_access_v2_integrity.sql',
    'shared/zz_role_access_v2_provisioning.sql',
    'shared/zz_role_access_v2_usernames.sql'
);
INSERT INTO user_app_permissions
    (business_id, owner_id, user_id, apps, role, is_active)
VALUES
    (1001, 1001, 1002, JSON_ARRAY('flow'), 'staff', TRUE)
ON DUPLICATE KEY UPDATE
    business_id = VALUES(business_id),
    apps = VALUES(apps),
    role = VALUES(role),
    is_active = VALUES(is_active);
SQL

SNAPSHOT_DIR="$(mktemp -d -t role-access-v2-rollback.XXXXXX)"
trap 'rm -rf "$SNAPSHOT_DIR"' EXIT
SNAPSHOT_FILE="$SNAPSHOT_DIR/pre-migration.sql"

DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$DB_NAME" \
    "$PROJECT_ROOT/scripts/rollback-role-access-control-v2.sh" snapshot "$SNAPSHOT_FILE"

DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$DB_NAME" \
    bash "$PROJECT_ROOT/scripts/migrate.sh"

post_table_count="$(dbq -N -e "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN ('authorization_permission_catalog','authorization_roles','authorization_role_permissions','authorization_business_members','authorization_license_assignments','authorization_role_assignments','authorization_role_assignment_scopes','authorization_dash_app_defaults','authorization_audit_events');")"
[ "$post_table_count" -eq 9 ] || { echo "expected 9 normalized authorization tables after migration, got $post_table_count" >&2; exit 1; }

ROLE_ACCESS_V2_ROLLBACK_CONFIRM=I_UNDERSTAND \
    DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$DB_NAME" \
    "$PROJECT_ROOT/scripts/rollback-role-access-control-v2.sh" restore "$SNAPSHOT_FILE"

ROLE_ACCESS_V2_ROLLBACK_CHECK=1 AUTHORIZATION_V2_MODE=legacy \
    DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$DB_NAME" \
    go test ./internal/modules/team/internal/application -run '^TestRoleAccessRollbackRestoresLegacyAuthorization$' -count=1

echo "role access V2 apply and rollback check passed"
