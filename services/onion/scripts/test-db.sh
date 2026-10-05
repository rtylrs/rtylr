#!/bin/bash
# scripts/test-db.sh — run the database-backed integration tests against the
# local-env MySQL container.
#
# It resets a dedicated `onion_test` database to a known state on every run
# (drop → recreate → migrate → seed fixtures) so the integration tests, which
# create rows, stay repeatable. Unit tests (no DB_HOST) run regardless; the
# integration tests opt in via the DB_HOST this script exports.
#
# Usage:
#   ./scripts/test-db.sh                 # reset DB + run ./... tests
#   ./scripts/test-db.sh ./internal/...  # reset DB + run a package subset
#
# Requires: the local-env stack running (make local-env-up). MySQL client is
# NOT required on the host — all SQL runs inside the db container.

set -eo pipefail

CONTAINER="${ONION_DB_CONTAINER:-onion-env-db-1}"
ROOT_PW="${ONION_DB_ROOT_PW:-onion_root_password}"
TEST_DB="onion_test"
TEST_DB_PORT="${ONION_TEST_DB_PORT:-3306}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/internal/testutil/fixtures"

GREEN='\033[0;32m'; BLUE='\033[0;34m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${BLUE}[test-db]${NC} $1"; }
ok()   { echo -e "${GREEN}[test-db]${NC} $1"; }
die()  { echo -e "${RED}[test-db]${NC} $1" >&2; exit 1; }

# mysql in the container. MYSQL_PWD avoids a password warning without hiding
# SQL errors from the caller. Extra mysql arguments are forwarded for the
# compatibility-layer metadata query below.
dbq() {
    local database="$1"
    shift
    docker exec -i -e "MYSQL_PWD=${ROOT_PW}" "$CONTAINER" mysql -uroot "$database" "$@"
}

if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER}$"; then
    die "db container '${CONTAINER}' not running. Start it with: make local-env-up"
fi

info "resetting ${TEST_DB} (drop → recreate)"
dbq "" <<SQL
DROP DATABASE IF EXISTS ${TEST_DB};
CREATE DATABASE ${TEST_DB} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
SQL

info "applying schema"
DB_HOST=127.0.0.1 DB_PORT="$TEST_DB_PORT" DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$TEST_DB" \
    MYSQL_PWD="$ROOT_PW" bash "$PROJECT_ROOT/scripts/migrate.sh"

info "seeding fixtures"
for f in "$FIXTURES_DIR"/*.sql; do
    [ -f "$f" ] || continue
    if [ "$(basename "$f")" = "70_direct_authorization.sql" ] && [ "${DIRECT_AUTHORIZATION_EVIDENCE:-}" != "1" ]; then
        continue
    fi
    dbq "$TEST_DB" < "$f" >/dev/null
done

if [ "${TEST_DB_LEGACY_SCOPE:-}" = "1" ]; then
    # Legacy fixture compatibility is opt-in.  These triggers rewrite zero
    # business IDs and can hide missing tenant scoping in tests.
    info "installing opt-in legacy fixture compatibility"
    compatibility_sql=$(dbq "$TEST_DB" -sN -e "
        SELECT CONCAT(
            'CREATE TRIGGER ', CHAR(96), 'trg_test_business_scope_', TABLE_NAME, CHAR(96), ' BEFORE INSERT ON ', CHAR(96),
            TABLE_NAME,
            CHAR(96), ' FOR EACH ROW SET NEW.business_id = IF(NEW.business_id = 0, NEW.owner_id, NEW.business_id), NEW.owner_id = IF(NEW.owner_id = 0, NEW.business_id, NEW.owner_id);'
        )
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME IN (
              SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
              WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
          )
          AND TABLE_NAME <> 'finance_invoices'
          AND COLUMN_NAME IN ('owner_id', 'business_id')
        GROUP BY TABLE_NAME
        HAVING COUNT(DISTINCT COLUMN_NAME) = 2
        ORDER BY TABLE_NAME;
    ")
    if [ -n "$compatibility_sql" ]; then
        dbq "$TEST_DB" <<< "$compatibility_sql" >/dev/null
    fi
else
    info "legacy fixture compatibility disabled"
fi

ok "database ready — running tests"
export DB_HOST=127.0.0.1 DB_PORT="$TEST_DB_PORT" DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$TEST_DB"
go clean -testcache
if [ "${TEST_DB_SETUP_ONLY:-}" != "1" ]; then
    go test -p 1 "${@:-./...}"
fi
