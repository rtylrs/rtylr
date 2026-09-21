#!/usr/bin/env bash
# local-env/mindsdb/bootstrap_test.sh — regression test for bootstrap.sh.
#
# Covers B60-R04: MindsDB answers POST /api/sql/query with HTTP 200 and a body
# such as {"type":"error","error_message":"..."} when a statement is rejected.
# The bootstrap must exit non-zero instead of printing "✅ ... complete", stop
# applying further files, and also fail when MindsDB never becomes ready
# (instead of falling through to the init loop).
# Run locally: ./local-env/mindsdb/bootstrap_test.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"

PASS=0
FAIL=0
ok() { PASS=$((PASS + 1)); echo "  ok: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1" >&2; }

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Fixture copy of the script under test with two init statements, plus stub
# curl/sleep clients so the test never talks to a real MindsDB.
mkdir -p "$WORK/mindsdb/init" "$WORK/bin"
cp "$BOOTSTRAP" "$WORK/mindsdb/bootstrap.sh"
cat > "$WORK/mindsdb/init/01_first.sql" <<'SQL'
CREATE DATABASE IF NOT EXISTS onion_db;
SQL
cat > "$WORK/mindsdb/init/02_second.sql" <<'SQL'
SELECT 1;
SQL

cat > "$WORK/bin/curl" <<'SH'
#!/usr/bin/env bash
url=""
for arg in "$@"; do
    case "$arg" in
        http://*|https://*) url="$arg" ;;
    esac
done
printf '%s\n' "$*" >> "$CURL_LOG"
case "$url" in
    */api/status)
        exit "$(cat "$STUB_STATUS_EXIT")"
        ;;
    */api/sql/query)
        printf '%s\n' "$*" >> "$CURL_QUERY_LOG"
        cat "$STUB_QUERY_RESPONSE"
        ;;
    *)
        exit 7
        ;;
esac
SH
cat > "$WORK/bin/sleep" <<'SH'
#!/usr/bin/env bash
exit 0
SH
chmod +x "$WORK/bin/curl" "$WORK/bin/sleep"

run_bootstrap() {
    : > "$WORK/query.log"
    CURL_LOG="$WORK/curl.log" CURL_QUERY_LOG="$WORK/query.log" \
        STUB_STATUS_EXIT="$WORK/status_exit" STUB_QUERY_RESPONSE="$WORK/query_response" \
        PATH="$WORK/bin:$PATH" MINDSDB_URL="http://mindsdb.invalid" \
        bash "$WORK/mindsdb/bootstrap.sh" 2>&1
}
query_posts() { grep -o '/api/sql/query' "$WORK/query.log" 2>/dev/null | wc -l | tr -d ' '; }

# 1. MindsDB never becomes ready -> non-zero, no SQL is sent.
echo 7 > "$WORK/status_exit"
echo '{"type":"ok"}' > "$WORK/query_response"
status=0
output="$(run_bootstrap)" || status=$?
if [ "$status" -ne 0 ]; then
    ok "unready MindsDB exits non-zero"
else
    fail "unready MindsDB exits non-zero (got 0: $(echo "$output" | tail -1))"
fi
case "$output" in
    *"timed out"*) ok "unready MindsDB reports a timeout" ;;
    *) fail "unready MindsDB reports a timeout" ;;
esac
if [ "$(query_posts)" -eq 0 ]; then
    ok "no SQL sent while MindsDB is unready"
else
    fail "no SQL sent while MindsDB is unready (got $(query_posts) posts)"
fi

# 2. HTTP 200 with an error payload -> non-zero, later files are not applied.
echo 0 > "$WORK/status_exit"
echo '{"type":"error","error_message":"Bindings not found"}' > "$WORK/query_response"
status=0
output="$(run_bootstrap)" || status=$?
if [ "$status" -ne 0 ]; then
    ok "rejected statement exits non-zero"
else
    fail "rejected statement exits non-zero (got 0: $(echo "$output" | tail -1))"
fi
case "$output" in
    *"Bindings not found"*) ok "rejection reason is reported" ;;
    *) fail "rejection reason is reported" ;;
esac
case "$output" in
    *"complete"*) fail "no success banner after a rejected statement" ;;
    *) ok "no success banner after a rejected statement" ;;
esac
if [ "$(query_posts)" -eq 1 ]; then
    ok "bootstrap stops at the first rejected statement"
else
    fail "bootstrap stops at the first rejected statement (got $(query_posts) posts, want 1)"
fi

# 3. Healthy run -> zero, every file applied.
echo '{"type":"ok"}' > "$WORK/query_response"
status=0
output="$(run_bootstrap)" || status=$?
if [ "$status" -eq 0 ]; then
    ok "healthy MindsDB exits zero"
else
    fail "healthy MindsDB exits zero (got $status: $(echo "$output" | tail -1))"
fi
case "$output" in
    *"complete"*) ok "healthy run prints the success banner" ;;
    *) fail "healthy run prints the success banner" ;;
esac
if [ "$(query_posts)" -eq 2 ]; then
    ok "healthy run applies every init file"
else
    fail "healthy run applies every init file (got $(query_posts) posts, want 2)"
fi

echo
echo "  passed $PASS, failed $FAIL"
[ "$FAIL" -eq 0 ]
