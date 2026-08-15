#!/usr/bin/env bash
set -euo pipefail

# Release tags created before this script used github.run_number as the patch
# component (for example 0.0.1112). Those legacy values are deliberately
# migrated to the first bounded-semver minor release, 0.1.0.

bump_version() {
  local current="$1"
  local major minor patch

  if [[ ! "$current" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    echo "Invalid version: $current" >&2
    return 1
  fi

  IFS=. read -r major minor patch <<< "$current"
  major=$((10#$major))
  minor=$((10#$minor))
  patch=$((10#$patch))

  if (( minor > 9 || patch > 9 )); then
    printf '0.1.0\n'
    return 0
  fi

  patch=$((patch + 1))
  if (( patch > 9 )); then
    patch=0
    minor=$((minor + 1))
  fi
  if (( minor > 9 )); then
    minor=0
    major=$((major + 1))
  fi

  printf '%d.%d.%d\n' "$major" "$minor" "$patch"
}

if [[ "${1:-}" == '--from' ]]; then
  [[ $# -eq 2 ]] || { echo 'Usage: next-app-version.sh --from MAJOR.MINOR.PATCH' >&2; exit 1; }
  bump_version "$2"
  exit 0
fi

app="${1:?Usage: next-app-version.sh APP [INITIAL_VERSION]}"
[[ "$app" =~ ^[a-z0-9-]+$ ]] || { echo "Invalid app: $app" >&2; exit 1; }
initial="${2:-0.0.1}"
[[ "$initial" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "Invalid initial version: $initial" >&2; exit 1; }

latest="$({
  git tag --list "${app}-v*" |
    sed -nE "s/^${app}-v([0-9]+\.[0-9]+\.[0-9]+)$/\1/p" |
    sort -V |
    tail -1
} || true)"

if [[ -z "$latest" ]]; then
  printf '%s\n' "$initial"
else
  bump_version "$latest"
fi
