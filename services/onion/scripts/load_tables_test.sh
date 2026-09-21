#!/usr/bin/env bash
# scripts/load_tables_test.sh - regression test for `make load-tables`.
#
# Covers B03-R05: the loader must read the schema tree the repo actually ships
# (internal/platform/database/schemas, not the removed db/schemas), and it must
# never drop existing tables - the old orphan check compared live table names
# against schema *file basenames*, so it dropped every real table.
# Run locally: ./scripts/load_tables_test.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PASS=0
FAIL=0
ok() { PASS=$((PASS + 1)); echo "  ok: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1" >&2; }

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Fixture checkout holding only the schemas the repo proves it has, plus stub
# mysql clients. MYSQL_ARGV_LOG/MYSQL_STDIN_LOG record what the loader ran.
mkdir -p "$WORK/repo/internal/platform/database/schemas/crm" "$WORK/bin"
cat > "$WORK/repo/internal/platform/database/schemas/crm/schema.sql" <<'SQL'
CREATE TABLE IF NOT EXISTS crm_customers (id BIGINT NOT NULL PRIMARY KEY);
SQL

cat > "$WORK/bin/mysql" <<'SH'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$MYSQL_ARGV_LOG"
# Report one pre-existing table so any drop-orphans pass is observable.
case "$*" in
    *"SHOW TABLES;"*) echo "crm_customers" ;;
esac
if [ ! -t 0 ]; then
    cat >> "$MYSQL_STDIN_LOG"
fi
exit 0
SH
cat > "$WORK/bin/mysqladmin" <<'SH'
#!/usr/bin/env bash
exit 0
SH
chmod +x "$WORK/bin/mysql" "$WORK/bin/mysqladmin"
: > "$WORK/argv.log"
: > "$WORK/stdin.log"

status=0
output="$(
    cd "$WORK/repo" && PATH="$WORK/bin:$PATH" \
        MYSQL_ARGV_LOG="$WORK/argv.log" MYSQL_STDIN_LOG="$WORK/stdin.log" \
        DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD=onion_root_password DB_NAME=onion_test \
        bash "$SCRIPT_DIR/load_tables.sh" 2>&1
)" || status=$?

if [ "$status" -eq 0 ]; then
    ok "loader exits 0"
else
    fail "loader exits 0 (got $status: $(echo "$output" | tail -1))"
fi

case "$output" in
    *"Schema directory"*not\ found*) fail "no missing-schema-directory failure" ;;
    *) ok "no missing-schema-directory failure" ;;
esac

case "$output" in
    *"loaded successfully"*) ok "schema load completes" ;;
    *) fail "schema load completes" ;;
esac

if grep -q "crm_customers" "$WORK/stdin.log"; then
    ok "shipped schema file is applied"
else
    fail "shipped schema file is applied"
fi

if grep -qi "DROP TABLE" "$WORK/argv.log"; then
    fail "no table is dropped (argv: $(grep -i 'drop table' "$WORK/argv.log" | head -1))"
else
    ok "no table is dropped"
fi

echo
echo "  passed $PASS, failed $FAIL"
[ "$FAIL" -eq 0 ]
