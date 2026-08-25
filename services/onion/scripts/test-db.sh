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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FIXTURES_DIR="$PROJECT_ROOT/scripts/fixtures"

GREEN='\033[0;32m'; BLUE='\033[0;34m'; RED='\033[0;31m'; NC='\033[0m'
info() { echo -e "${BLUE}[test-db]${NC} $1"; }
ok()   { echo -e "${GREEN}[test-db]${NC} $1"; }
die()  { echo -e "${RED}[test-db]${NC} $1" >&2; exit 1; }

# mysql in the container, quiet about the password warning.
dbq() { docker exec -i "$CONTAINER" sh -c "mysql -uroot -p${ROOT_PW} $1" 2>&1 | sed '/Using a password/d'; }

if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER}$"; then
    die "db container '${CONTAINER}' not running. Start it with: make local-env-up"
fi

info "resetting ${TEST_DB} (drop → recreate)"
dbq "" <<SQL
DROP DATABASE IF EXISTS ${TEST_DB};
CREATE DATABASE ${TEST_DB} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
SQL

info "applying schema"
DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$TEST_DB" \
    MYSQL_PWD="$ROOT_PW" bash "$PROJECT_ROOT/scripts/migrate.sh"

info "seeding fixtures"
for f in "$FIXTURES_DIR"/*.sql; do
    [ -f "$f" ] || continue
    dbq "$TEST_DB" < "$f" >/dev/null
done

ok "database ready — running tests"
export DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD="$ROOT_PW" DB_NAME="$TEST_DB"
go clean -testcache
go test "${@:-./...}"
