#!/usr/bin/env bash
set -euo pipefail

git -C voxie fetch --quiet origin main

expected="$(git -C voxie rev-parse origin/main)"
actual="$(git -C voxie rev-parse HEAD)"

if [[ "$actual" != "$expected" ]]; then
  echo "::error::onion/voxie points to $actual, but Voxie main is $expected. Update the submodule pointer in Onion." >&2
  exit 1
fi

echo "onion/voxie is pinned to Voxie main at $actual"
