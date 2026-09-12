#!/usr/bin/env bash
# scripts/dsn.sh
#
# Derive DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME from a Go MySQL DSN.
#
# Production sets only DATABASE_URL; it has no DB_* variables at all. See the
# comment on resolveMainDSN in internal/platform/database/dbclient/mysql.go,
# which says so directly. scripts/migrate.sh needs DB_*, so without this the
# migration container would either refuse to start ("DB_USER is required") or,
# worse, fall back to the localhost defaults and report success against nothing.
#
# The format is the go-sql-driver DSN, which is NOT a URL:
#
#   [username[:password]@][protocol[(address)]]/dbname[?param=value]
#
# Parsing follows the driver's own rules, because a password is allowed to
# contain the characters a naive split would break on:
#
#   - the dbname starts after the LAST '/' that precedes any '?'
#   - within the part before it, the credentials end at the LAST '@'
#   - within the credentials, the username ends at the FIRST ':'
#
# So `u:p@ss:w/rd@tcp(h:3306)/db` yields password `p@ss:w/rd`, which is what the
# driver does and what a naive cut on the first '@' or ':' would get wrong.

# Print `KEY=VALUE` lines for the DSN in $1. Returns non-zero on a malformed DSN.
parse_mysql_dsn() {
    local dsn="$1"

    if [ -z "$dsn" ]; then
        echo "parse_mysql_dsn: empty DSN" >&2
        return 1
    fi

    # Strip query parameters; a '?' inside a password is not legal unescaped.
    local without_params="${dsn%%\?*}"

    # dbname is everything after the last '/'.
    local dbname="${without_params##*/}"
    local prefix="${without_params%/*}"

    if [ "$prefix" = "$without_params" ]; then
        echo "parse_mysql_dsn: no '/' separating the database name" >&2
        return 1
    fi

    # Credentials end at the last '@'; the address follows it.
    local credentials address
    if [[ "$prefix" == *@* ]]; then
        credentials="${prefix%@*}"
        address="${prefix##*@}"
    else
        credentials=""
        address="$prefix"
    fi

    # Username ends at the first ':'; everything after is the password.
    local user password
    if [[ "$credentials" == *:* ]]; then
        user="${credentials%%:*}"
        password="${credentials#*:}"
    else
        user="$credentials"
        password=""
    fi

    # address is protocol(host:port), e.g. tcp(db.example.com:3306).
    local hostport="$address"
    if [[ "$address" == *"("*")" ]]; then
        hostport="${address#*(}"
        hostport="${hostport%)}"
    fi

    local host port
    if [[ "$hostport" == *:* ]]; then
        host="${hostport%:*}"
        port="${hostport##*:}"
    else
        host="$hostport"
        port="3306"
    fi

    if [ -z "$host" ] || [ -z "$dbname" ]; then
        echo "parse_mysql_dsn: could not read a host and database name" >&2
        return 1
    fi

    printf 'DB_HOST=%s\n' "$host"
    printf 'DB_PORT=%s\n' "$port"
    printf 'DB_USER=%s\n' "$user"
    printf 'DB_PASSWORD=%s\n' "$password"
    printf 'DB_NAME=%s\n' "$dbname"
}

# Export DB_* from DATABASE_URL, without overwriting anything already set.
# Silent and successful when DATABASE_URL is absent, so a caller that already
# has DB_* is unaffected.
export_db_vars_from_dsn() {
    [ -n "${DATABASE_URL:-}" ] || return 0

    local parsed
    parsed="$(parse_mysql_dsn "$DATABASE_URL")" || return 1

    local line key value
    while IFS= read -r line; do
        key="${line%%=*}"
        value="${line#*=}"
        # An explicit DB_* wins, so an operator can still override one field.
        if [ -z "${!key:-}" ]; then
            export "$key=$value"
        fi
    done <<< "$parsed"
}
