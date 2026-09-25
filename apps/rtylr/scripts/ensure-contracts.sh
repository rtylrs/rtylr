#!/usr/bin/env bash
set -euo pipefail

# Generated contract output is gitignored, so a fresh checkout (or one where
# gen/ was cleaned) needs a generation pass before anything typechecks or
# builds against @voxire/voxie.
repo_root="$(git rev-parse --show-toplevel)"
voxie_dir="$repo_root/voxie"

if [[ ! -f "$voxie_dir/onekit.toml" ]]; then
  echo "::error::in-tree contracts missing at $voxie_dir" >&2
  exit 1
fi

if [[ ! -d "$voxie_dir/gen/ts" || ! -d "$voxie_dir/gen/go" ]]; then
  echo "Generating Voxie contracts into $voxie_dir/gen ..."
  make -C "$voxie_dir" generate
fi
