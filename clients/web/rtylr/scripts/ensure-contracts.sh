#!/usr/bin/env bash
set -euo pipefail

# Contract output is generated, not committed, so a fresh checkout (or one
# where it was cleaned) needs a generation pass before anything typechecks or
# builds against @voxire/voxie.
repo_root="$(git rev-parse --show-toplevel)"
voxie_dir="$repo_root/voxie"
contracts_src="$repo_root/clients/packages/contracts/src"

if [[ ! -f "$voxie_dir/onekit.toml" ]]; then
  echo "::error::in-tree contracts missing at $voxie_dir" >&2
  exit 1
fi

if [[ ! -d "$contracts_src" || ! -d "$voxie_dir/gen/go" ]]; then
  echo "Generating Voxie contracts..."
  make -C "$voxie_dir" generate
fi
