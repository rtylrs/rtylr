#!/usr/bin/env bash
# scripts/dsn_test.sh — exercises parse_mysql_dsn against the shapes a real DSN takes.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./dsn.sh
source "$SCRIPT_DIR/dsn.sh"

PASS=0
FAIL=0

expect() {
    local name="$1" dsn="$2" field="$3" want="$4"
    local got
    got="$(parse_mysql_dsn "$dsn" 2>/dev/null | grep "^${field}=" | cut -d= -f2-)"
    if [ "$got" = "$want" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        printf '  FAIL %-46s %s\n       want %-24s got %s\n' "$name" "$field" "'$want'" "'$got'"
    fi
}

expect_error() {
    local name="$1" dsn="$2"
    if parse_mysql_dsn "$dsn" >/dev/null 2>&1; then
        FAIL=$((FAIL + 1)); printf '  FAIL %-46s expected a non-zero exit\n' "$name"
    else
        PASS=$((PASS + 1))
    fi
}

PLAIN='onion:secret@tcp(db.internal:3306)/onion_prod?parseTime=true&loc=UTC'
expect "plain: host"     "$PLAIN" DB_HOST     "db.internal"
expect "plain: port"     "$PLAIN" DB_PORT     "3306"
expect "plain: user"     "$PLAIN" DB_USER     "onion"
expect "plain: password" "$PLAIN" DB_PASSWORD "secret"
expect "plain: name"     "$PLAIN" DB_NAME     "onion_prod"

# A password holding '@', ':' and '/' is the case a naive cut gets wrong.
NASTY='onion:p@ss:w/rd@tcp(db.internal:3306)/onion_prod?parseTime=true'
expect "password with @ : /: user"     "$NASTY" DB_USER     "onion"
expect "password with @ : /: password" "$NASTY" DB_PASSWORD 'p@ss:w/rd'
expect "password with @ : /: host"     "$NASTY" DB_HOST     "db.internal"
expect "password with @ : /: name"     "$NASTY" DB_NAME     "onion_prod"

# The real production endpoint shape.
RDS='onion:pw@tcp(voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com:3306)/voxire?charset=utf8mb4'
expect "rds: host" "$RDS" DB_HOST "voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com"
expect "rds: name" "$RDS" DB_NAME "voxire"

NOPARAMS='u:p@tcp(h:3307)/d'
expect "no query string: port" "$NOPARAMS" DB_PORT "3307"
expect "no query string: name" "$NOPARAMS" DB_NAME "d"

NOPORT='u:p@tcp(h)/d'
expect "address without a port defaults to 3306" "$NOPORT" DB_PORT "3306"
expect "address without a port: host"            "$NOPORT" DB_HOST "h"

NOPASS='u@tcp(h:3306)/d'
expect "user without a password: user"     "$NOPASS" DB_USER     "u"
expect "user without a password: password" "$NOPASS" DB_PASSWORD ""

expect_error "empty DSN"              ""
expect_error "no database name"       "u:p@tcp(h:3306)"

# export_db_vars_from_dsn must not clobber an explicit override.
(
    export DATABASE_URL="$PLAIN" DB_NAME="override_db"
    export_db_vars_from_dsn
    [ "$DB_NAME" = "override_db" ] && [ "$DB_HOST" = "db.internal" ]
) && PASS=$((PASS + 1)) || { FAIL=$((FAIL + 1)); echo "  FAIL explicit DB_NAME must win over the DSN"; }

# With no DATABASE_URL it is a silent success, leaving DB_* untouched.
(
    unset DATABASE_URL
    export DB_HOST="already-set"
    export_db_vars_from_dsn
    [ "$DB_HOST" = "already-set" ]
) && PASS=$((PASS + 1)) || { FAIL=$((FAIL + 1)); echo "  FAIL absent DATABASE_URL must be a no-op"; }

echo
echo "  passed $PASS, failed $FAIL"
[ "$FAIL" -eq 0 ]
