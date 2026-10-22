#!/usr/bin/env bash
set -euo pipefail

# Compiles onion release binaries on the runner instead of in the Dockerfile's
# builder stage, so the Go caches that actions/cache restores apply:
#
#   build-onion-binaries.sh <dockerfile> <out-dir> <package>...
#
# Each package (relative to the Dockerfile's directory, e.g. ./cmd/api) lands
# at <out-dir>/app/<last path element>, where the Dockerfile's final stage
# copies it from the builder stage that <out-dir> replaces. The onek and sqlc
# versions come from the Dockerfile, so both paths generate the same code.

dockerfile="${1:?usage: build-onion-binaries.sh <dockerfile> <out-dir> <package>...}"
out="${2:?usage: build-onion-binaries.sh <dockerfile> <out-dir> <package>...}"
shift 2
[ "$#" -gt 0 ] || { echo "No packages to build" >&2; exit 1; }
[ "$(uname -m)" = x86_64 ] || { echo "This downloads x86_64 tools; run it on an x86_64 runner" >&2; exit 1; }

root="$(git rev-parse --show-toplevel)"
module="$(cd "$(dirname "$dockerfile")" && pwd)"
mkdir -p "$out/app"
out="$(cd "$out" && pwd)"
onek_version="$(sed -n 's|.*github.com/1homsi/onekit/cmd/onek@v\([0-9][0-9.]*\).*|\1|p' "$dockerfile" | head -n 1)"
sqlc_version="$(sed -n 's|.*/sqlc_\([0-9][0-9.]*\)_linux_amd64.*|\1|p' "$dockerfile" | head -n 1)"
if [ -z "$onek_version" ] || [ -z "$sqlc_version" ]; then
  echo "Could not read the onek and sqlc versions from $dockerfile" >&2
  exit 1
fi

tools="$(mktemp -d)"
trap 'rm -rf "$tools"' EXIT
fetch() { curl --fail --silent --show-error --location --retry 5 --retry-all-errors --retry-delay 2 "$@"; }

# Prebuilt releases rather than `go install`, which compiles them. onek
# publishes checksums; sqlc doesn't, and the Dockerfile trusts its download too.
onek_archive="onekit_${onek_version}_Linux_x86_64.tar.gz"
onek_release="https://github.com/1homsi/onekit/releases/download/v$onek_version"
fetch -o "$tools/$onek_archive" "$onek_release/$onek_archive"
fetch -o "$tools/onek-checksums.txt" "$onek_release/checksums.txt"
(cd "$tools" && grep " $onek_archive\$" onek-checksums.txt | sha256sum --check --status)
tar -xzf "$tools/$onek_archive" -C "$tools" onek

sqlc_archive="sqlc_${sqlc_version}_linux_amd64.tar.gz"
fetch -o "$tools/$sqlc_archive" "https://github.com/sqlc-dev/sqlc/releases/download/v$sqlc_version/$sqlc_archive"
tar -xzf "$tools/$sqlc_archive" -C "$tools" sqlc
export PATH="$tools:$PATH"

(cd "$root/voxie" && onek build .)
(cd "$module" && bash scripts/generate_sqlc.sh)

for package in "$@"; do
  (cd "$module" && TARGETARCH=amd64 bash scripts/build-binary.sh "$package" "$out/app/${package##*/}")
done
