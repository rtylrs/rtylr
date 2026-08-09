#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
recorded="$(git -C "$repo_root" ls-files -s -- voxie | awk '$1 == "160000" { print $2; exit }')"

if [[ -z "$recorded" ]]; then
  echo "::error::packages/voxie is not recorded as a git submodule. Contract output must remain owned by Voxie." >&2
  exit 1
fi

if ! actual="$(git -C "$repo_root/voxie" rev-parse --verify HEAD^{commit} 2>/dev/null)"; then
  echo "::error::packages/voxie is not initialized at a commit. Run git submodule update --init --recursive." >&2
  exit 1
fi

if [[ "$actual" != "$recorded" ]]; then
  echo "::error::packages/voxie is checked out at $actual, but packages records $recorded. Run git submodule update --init --recursive." >&2
  exit 1
fi

for generated_dir in gen/go gen/ts; do
  if [[ ! -d "$repo_root/voxie/$generated_dir" ]]; then
    echo "::error::packages/voxie@$actual is missing $generated_dir output." >&2
    exit 1
  fi
done

echo "packages/voxie is pinned to generated Voxie commit $actual"
