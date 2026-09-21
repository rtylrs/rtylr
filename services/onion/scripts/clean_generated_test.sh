#!/usr/bin/env bash
# scripts/clean_generated_test.sh - regression test for `make clean-generated`.
#
# Covers B22-R06: the cleanup must delete generated sqlc output but never the
# hand-written, git-tracked schema supplement under
# internal/platform/database/schemas/sqlc, and a deletion that cannot succeed
# must not be reported as success. Run locally: ./scripts/clean_generated_test.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PASS=0
FAIL=0
ok() { PASS=$((PASS + 1)); echo "  ok: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1" >&2; }

WORK="$(mktemp -d)"
trap 'chmod -R u+w "$WORK" 2>/dev/null || true; rm -rf "$WORK"' EXIT

SUPPLEMENT="internal/platform/database/schemas/sqlc/flow/v8_time_tracking_columns.sql"

# Fixture tree with the generated-output layouts the cleanup targets, plus the
# tracked supplement that must survive.
mkdir -p "$WORK/repo/internal/modules/user/internal/infrastructure/persistence/sqlc"
mkdir -p "$WORK/repo/internal/platform/aicache/persistence/sqlc"
mkdir -p "$WORK/repo/internal/platform/database/schemas/sqlc/flow"
echo 'package sqlc' > "$WORK/repo/internal/modules/user/internal/infrastructure/persistence/sqlc/users.sql.go"
echo 'package sqlc' > "$WORK/repo/internal/modules/user/internal/infrastructure/persistence/sqlc/db.go"
echo 'package sqlc' > "$WORK/repo/internal/platform/aicache/persistence/sqlc/db.go"
echo '-- hand-written sqlc schema supplement' > "$WORK/repo/$SUPPLEMENT"

if ( cd "$WORK/repo" && printf 'y' | bash "$SCRIPT_DIR/clean_generated.sh" >/dev/null 2>&1 ); then
    ok "clean run exits 0"
else
    fail "clean run exits 0"
fi

if [ -f "$WORK/repo/$SUPPLEMENT" ]; then
    ok "git-tracked schema supplement is preserved"
else
    fail "git-tracked schema supplement is preserved"
fi

if [ ! -e "$WORK/repo/internal/modules/user/internal/infrastructure/persistence/sqlc" ]; then
    ok "module sqlc output removed"
else
    fail "module sqlc output removed"
fi

if [ ! -e "$WORK/repo/internal/platform/aicache/persistence/sqlc" ]; then
    ok "platform sqlc output removed"
else
    fail "platform sqlc output removed"
fi

# A deletion that cannot succeed must fail the script instead of printing
# success. Root ignores the directory permissions, so skip this case there.
if [ "$(id -u)" -ne 0 ]; then
    readonly_parent="$WORK/readonly/internal/modules/user/internal/infrastructure/persistence"
    mkdir -p "$readonly_parent/sqlc"
    echo 'package sqlc' > "$readonly_parent/sqlc/users.sql.go"
    chmod 500 "$readonly_parent"

    status=0
    output="$( cd "$WORK/readonly" && printf 'y' | bash "$SCRIPT_DIR/clean_generated.sh" 2>&1 )" || status=$?
    if [ "$status" -ne 0 ]; then
        ok "unremovable output exits non-zero"
    else
        fail "unremovable output exits non-zero"
    fi

    case "$output" in
        *"cleaned successfully"*) fail "no success message after a failed deletion" ;;
        *) ok "no success message after a failed deletion" ;;
    esac

    chmod 700 "$readonly_parent"
fi

# Every schema input referenced by a real sqlc.yaml must exist: that is what
# `make generate` (which the cleanup tells the developer to run next) needs.
missing="$(python3 - "$REPO_ROOT" <<'PY'
import os
import re
import sys

root = sys.argv[1]
missing = []
for dirpath, _, files in os.walk(os.path.join(root, "internal")):
    if "sqlc.yaml" not in files:
        continue
    config = os.path.join(dirpath, "sqlc.yaml")
    in_schema = False
    for raw in open(config):
        stripped = raw.strip()
        if stripped.startswith("schema:"):
            in_schema = True
            continue
        if not in_schema:
            continue
        match = re.match(r'^-\s*"([^"]+)"', stripped)
        if match:
            target = os.path.normpath(os.path.join(dirpath, match.group(1)))
            if not os.path.exists(target):
                missing.append(os.path.relpath(target, root))
        elif stripped and not stripped.startswith("#"):
            in_schema = False
print("\n".join(missing))
PY
)"
if [ -z "$missing" ]; then
    ok "every sqlc.yaml schema input exists"
else
    fail "sqlc.yaml schema inputs are missing: $missing"
fi

echo
echo "  passed $PASS, failed $FAIL"
[ "$FAIL" -eq 0 ]
