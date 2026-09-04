#!/usr/bin/env bash
# Run every operation-level authorization matrix against a fresh database.
# The matrix includes allowed mutations, so ordinary integration packages are
# intentionally excluded from this database run.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

package_args=()
while IFS= read -r package_dir; do
    package_args+=("./${package_dir#"$PROJECT_ROOT/"}")
done < <(
    find "$PROJECT_ROOT/internal" -name 'direct_authorization_matrix_test.go' -print \
        | xargs -n1 dirname \
        | sort -u
)

if [ "${#package_args[@]}" -eq 0 ]; then
    echo "No direct authorization matrix packages found" >&2
    exit 1
fi

cd "$PROJECT_ROOT"
DIRECT_AUTHORIZATION_EVIDENCE=1 TEST_DB_LEGACY_SCOPE=0 \
    "$PROJECT_ROOT/scripts/test-db.sh" "${package_args[@]}"
