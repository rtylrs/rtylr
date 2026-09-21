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
# Two shapes are accepted, matching what the Go layer connects with
# (dbclient/mysql.go normalizeDSN):
#
#   native: [username[:password]@][protocol[(address)]]/dbname[?param=value]
#   URL:    mysql://username[:password]@host[:port]/dbname[?param=value]
#
# Parsing follows the driver's own rules, because a password is allowed to
# contain the characters a naive split would break on:
#
#   - in the native form the dbname starts after the LAST '/', and the query
#     string at the first '?' that follows it. A '?' earlier in the DSN (inside
#     a password) is not a delimiter: the driver scans for '/' first.
#   - within the part before the dbname, the credentials end at the LAST '@'
#   - within the credentials, the username ends at the FIRST ':'
#
# So `u:p@ss:w/rd?x@tcp(h:3306)/db` yields password `p@ss:w/rd?x`, which is what
# the driver does and what a naive cut on the first '@', ':' or '?' would get
# wrong.

# Percent-decode a URL-form user or password, the way url.Parse does in
# normalizeDSN before it rebuilds the native DSN. Anything that is not a valid
# '%XX' escape is kept verbatim, so a literal '%' survives.
percent_decode() {
    local input="$1"
    local output=""
    local hex

    while [[ "$input" == *%* ]]; do
        output+="${input%%\%*}"
        input="${input#*\%}"
        hex="${input:0:2}"
        input="${input:2}"
        if [[ "$hex" =~ ^[0-9A-Fa-f]{2}$ ]]; then
            output+="$(printf '%b' "\\x${hex}")"
        else
            output+="%${hex}"
        fi
    done

    printf '%s' "$output$input"
}

# Print `KEY=VALUE` lines for the DSN in $1. Returns non-zero on a malformed DSN.
parse_mysql_dsn() {
    local dsn="$1"

    if [ -z "$dsn" ]; then
        echo "parse_mysql_dsn: empty DSN" >&2
        return 1
    fi

    local url_form=false
    local credentials address dbname

    if [[ "$dsn" == *"://"* ]]; then
        # URL form. Only the scheme normalizeDSN rewrites is supported; any
        # other scheme is a misconfiguration and must not be exported as
        # credentials.
        local scheme="${dsn%%://*}"
        if [ "$scheme" != "mysql" ]; then
            echo "parse_mysql_dsn: unsupported DSN scheme '$scheme' (expected 'mysql://' or the native driver form)" >&2
            return 1
        fi

        local rest="${dsn#*://}"
        local authority="${rest%%/*}"
        if [ "$authority" = "$rest" ]; then
            echo "parse_mysql_dsn: no '/' separating the database name" >&2
            return 1
        fi

        # The dbname follows the authority and ends at the query string.
        local after_authority="${rest#*/}"
        dbname="${after_authority%%\?*}"

        # Credentials end at the last '@' in the authority.
        if [[ "$authority" == *@* ]]; then
            credentials="${authority%@*}"
            address="${authority##*@}"
        else
            credentials=""
            address="$authority"
        fi

        url_form=true
    else
        # Native form. The dbname starts after the LAST '/' (a password may
        # contain '/'), and the query string after the first '?' that follows
        # it, so a '?' before the dbname stays in the password.
        local head="${dsn%/*}"
        local tail="${dsn##*/}"

        if [ "$head" = "$dsn" ]; then
            echo "parse_mysql_dsn: no '/' separating the database name" >&2
            return 1
        fi

        dbname="${tail%%\?*}"

        # Credentials end at the last '@'; the address follows it.
        if [[ "$head" == *@* ]]; then
            credentials="${head%@*}"
            address="${head##*@}"
        else
            credentials=""
            address="$head"
        fi
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

    if [ "$url_form" = true ]; then
        user="$(percent_decode "$user")"
        password="$(percent_decode "$password")"
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
