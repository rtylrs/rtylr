#!/usr/bin/env bash
#
# Regression test for zz_finance_tax_config_canonical.sql.
#
# Legacy tax configs are owner-scoped (owner_id UNIQUE), so one business can
# hold several rows, including several sharing (business_id, name, type) and
# several flagged is_default. The migration must dedupe those tuples before
# adding the unique key, and must leave exactly one default per business:
# repair a business that has configs but none default, without promoting the
# other rows of a business that already has a default.

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
: "${DB_USER:?DB_USER is required}"
: "${DB_PASSWORD:?DB_PASSWORD is required}"

DB_NAME="schemas_finance_tax_config_canonical_test_$$"
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

mysql_cmd "$DB_NAME" < finance/schema.sql

mysql_cmd "$DB_NAME" <<'SQL'
INSERT INTO finance_tax_config
    (id, owner_id, business_id, name, type, rate, tax_name, tax_rate, is_default)
VALUES
    -- Duplicate (business_id, name, type) written by two users: the default
    -- row keeps the name, the non-default row is renamed.
    (10, 1001, 2001, 'VAT', 'vat', 20.00, NULL, NULL, TRUE),
    (11, 1002, 2001, 'VAT', 'vat', 5.00, NULL, NULL, FALSE),
    -- Names normalized from the legacy columns: row 12 becomes 'Standard VAT'
    -- and is the lowest id, so it is promoted when no default exists.
    (12, 1003, 2002, '', 'vat', 0.00, 'Standard VAT', 20.00, FALSE),
    (13, 1004, 2002, 'Reduced', 'vat', 5.00, NULL, NULL, FALSE),
    -- An existing default (15) must stay the only default; the duplicate with
    -- the lower id is renamed.
    (14, 1005, 2003, 'VAT', 'vat', 5.00, NULL, NULL, FALSE),
    (15, 1006, 2003, 'VAT', 'vat', 20.00, NULL, NULL, TRUE),
    -- Legacy rows that were already flagged default twice: the lowest id (16)
    -- stays the only default and the other two duplicates are renamed.
    (16, 1007, 2004, 'VAT', 'vat', 20.00, NULL, NULL, TRUE),
    (17, 1008, 2004, 'VAT', 'vat', 5.00, NULL, NULL, TRUE),
    (18, 1009, 2004, 'VAT', 'vat', 0.00, NULL, NULL, FALSE),
    -- Two empty names normalize to the same 'Default tax' tuple.
    (19, 1010, 2005, '', 'vat', 0.00, NULL, NULL, FALSE),
    (20, 1011, 2005, '', 'vat', 0.00, NULL, NULL, FALSE);
SQL

if ! mysql_cmd "$DB_NAME" < finance/zz_finance_tax_config_canonical.sql >"$OUTPUT_FILE" 2>&1; then
    echo "tax config canonical migration failed" >&2
    cat "$OUTPUT_FILE" >&2
    exit 1
fi

state=$(mysql_cmd "$DB_NAME" -Nse "SELECT GROUP_CONCAT(CONCAT(id, ':', name, ':', is_default) ORDER BY id) FROM finance_tax_config")
if [[ "$state" != "10:VAT:1,11:VAT-11:0,12:Standard VAT:1,13:Reduced:0,14:VAT-14:0,15:VAT:1,16:VAT:1,17:VAT-17:0,18:VAT-18:0,19:Default tax:1,20:Default tax-20:0" ]]; then
    echo "unexpected tax config rows after migration: $state" >&2
    cat "$OUTPUT_FILE" >&2
    exit 1
fi

duplicates=$(mysql_cmd "$DB_NAME" -Nse "SELECT COUNT(*) FROM (
    SELECT 1 FROM finance_tax_config GROUP BY business_id, name, type HAVING COUNT(*) > 1
) AS duplicate_groups")
if [[ "$duplicates" != "0" ]]; then
    echo "duplicate (business_id, name, type) tuples remain: $duplicates" >&2
    exit 1
fi

bad_defaults=$(mysql_cmd "$DB_NAME" -Nse "SELECT COUNT(*) FROM (
    SELECT business_id FROM finance_tax_config GROUP BY business_id HAVING SUM(is_default = TRUE) <> 1
) AS bad_default_businesses")
if [[ "$bad_defaults" != "0" ]]; then
    echo "businesses without exactly one default: $bad_defaults" >&2
    exit 1
fi

indexes=$(mysql_cmd "$DB_NAME" -Nse "SELECT CONCAT(
    (SELECT COUNT(DISTINCT INDEX_NAME) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_tax_config' AND INDEX_NAME = 'owner_id'), ':',
    (SELECT COUNT(DISTINCT INDEX_NAME) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_tax_config' AND INDEX_NAME = 'uq_finance_tax_config_business_name_type')
)")
if [[ "$indexes" != "0:1" ]]; then
    echo "unexpected tax config indexes (owner_id:unique_key): $indexes" >&2
    exit 1
fi

echo "finance tax config canonical migration regression passed"
