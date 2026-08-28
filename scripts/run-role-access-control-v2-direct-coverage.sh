#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
onion_dir="${ONION_DIR:-$(cd "$script_dir/../../onion" && pwd)}"

cd "$onion_dir"

packages=()
while IFS= read -r package_dir; do
	packages+=("./${package_dir}")
done < <(
	rg --files internal/modules \
		| rg '/direct_(authorization_matrix|sensitive_authorization)_test\.go$' \
		| sed 's#/[^/]*$##' \
		| sort -u
)

if ((${#packages[@]} == 0)); then
	echo "No direct authorization test packages found" >&2
	exit 1
fi

export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-3306}"
export DB_USER="${DB_USER:-root}"
export DB_PASSWORD="${DB_PASSWORD:-onion_root_password}"
export DB_NAME="${DB_NAME:-onion_test}"
export DIRECT_AUTHORIZATION_EVIDENCE=1

if [[ "${ROLE_ACCESS_DIRECT_DB_READY:-}" != "1" ]]; then
	export ROLE_ACCESS_DIRECT_DB_READY=1
	TEST_DB_SETUP_ONLY=1 "$onion_dir/scripts/test-db.sh" "${packages[@]}"
fi

go test -p 1 -count=1 -run '^(TestDirectAuthorizationMatrix|TestDirectSensitiveAuthorization)$' "${packages[@]}"
