#!/usr/bin/env bash
set -euo pipefail

git -C packages fetch --quiet origin main

expected_packages="$(git -C packages rev-parse origin/main)"
actual_packages="$(git -C packages rev-parse HEAD)"

if [[ "$actual_packages" != "$expected_packages" ]]; then
  echo "::error::rtylr/packages points to $actual_packages, but packages main is $expected_packages. Update the submodule pointer in rtylr." >&2
  exit 1
fi

git -C packages/voxie fetch --quiet origin main

expected_voxie="$(git -C packages/voxie rev-parse origin/main)"
actual_voxie="$(git -C packages/voxie rev-parse HEAD)"

if [[ "$actual_voxie" != "$expected_voxie" ]]; then
  echo "::error::packages/voxie points to $actual_voxie, but Voxie main is $expected_voxie. Update the nested submodule pointer in packages." >&2
  exit 1
fi

echo "rtylr/packages is pinned to packages main at $actual_packages"
echo "packages/voxie is pinned to Voxie main at $actual_voxie"
