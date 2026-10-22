#!/bin/bash
set -euo pipefail

# SQLC Generation Script
# Generates sqlc code for every module and platform package with a sqlc.yaml
# (SQL case checking lives in a separate script). The configs are independent,
# so they generate in parallel.

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Generating sqlc files for all modules...${NC}"

configs="$(find internal/modules internal/platform -name "sqlc.yaml" -type f)"
if [ -z "$configs" ]; then
    echo "  No modules with sqlc.yaml found!"
    exit 1
fi
module_count="$(printf '%s\n' "$configs" | wc -l | tr -d ' ')"
jobs="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"

# xargs exits non-zero when any generation fails.
printf '%s\n' "$configs" | xargs -n 1 -P "$jobs" sh -c '
    dir="$(dirname "$1")"
    if ! (cd "$dir" && sqlc generate); then
        echo "      ✗ Failed to generate sqlc for $dir" >&2
        exit 1
    fi
' sh

echo -e "${GREEN}Successfully generated sqlc code for $module_count modules!${NC}"
