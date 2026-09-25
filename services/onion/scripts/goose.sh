#!/bin/bash
# scripts/goose.sh
# Compatibility wrapper for the canonical schema migration runner.
#
# Usage:
#   ./scripts/goose.sh up

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
if [ "${1:-}" = "up" ] && [ "$#" -eq 1 ]; then
    exec "$PROJECT_ROOT/scripts/migrate.sh"
fi

echo "goose.sh no longer manages migrations: use './scripts/migrate.sh' (only 'up' is supported)" >&2
exit 2
