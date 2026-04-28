#!/bin/bash
# Smoke-test the PATCH/MINOR/MAJOR rollover logic from .github/workflows/build-and-push.yml.
# Run locally: ./scripts/test_version_bump.sh

set -euo pipefail

bump() {
    local LATEST="$1"
    if [ -z "$LATEST" ]; then
        echo "0.1.1"; return
    fi
    IFS='.' read -r MAJOR MINOR PATCH <<< "$LATEST"
    if [ "$PATCH" -ge 99 ]; then
        PATCH=0
        if [ "$MINOR" -ge 99 ]; then
            MINOR=0
            MAJOR=$((MAJOR + 1))
        else
            MINOR=$((MINOR + 1))
        fi
    else
        PATCH=$((PATCH + 1))
    fi
    echo "${MAJOR}.${MINOR}.${PATCH}"
}

assert_eq() {
    local got="$1" want="$2" label="$3"
    if [ "$got" != "$want" ]; then
        echo "FAIL: $label — got $got, want $want" >&2
        exit 1
    fi
    echo "ok: $label ($got)"
}

assert_eq "$(bump '')"        "0.1.1" "empty -> seed"
assert_eq "$(bump '0.1.0')"   "0.1.1" "patch increment"
assert_eq "$(bump '0.1.50')"  "0.1.51" "mid-patch increment"
assert_eq "$(bump '0.1.98')"  "0.1.99" "approach patch ceiling"
assert_eq "$(bump '0.1.99')"  "0.2.0"  "patch rollover -> minor"
assert_eq "$(bump '0.99.99')" "1.0.0"  "minor + patch rollover -> major"
assert_eq "$(bump '1.99.99')" "2.0.0"  "major bump"
assert_eq "$(bump '5.10.50')" "5.10.51" "non-zero major patch"

echo "all version bump cases passed"
