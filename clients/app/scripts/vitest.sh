#!/usr/bin/env bash
# Runs this workspace's Vitest suite from the repository root.
#
# Vitest only loads test files that live under the project they belong to, and
# the shared packages live at the repository root, outside this workspace
# (apps/rtylr). Rooting the run at the repository with this workspace's config
# lets both the "apps" and "packages" projects resolve.
set -euo pipefail

workspace="$(cd "$(dirname "$0")/.." && pwd)"
root="$(git -C "$workspace" rev-parse --show-toplevel)"
cd "$root"

exec "$workspace/node_modules/.bin/vitest" --config "$workspace/vitest.config.ts" "$@"
