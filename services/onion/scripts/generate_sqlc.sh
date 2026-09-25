#!/bin/bash
set -e

# SQLC Generation Script
# This script generates sqlc code for all modules (SQL case checking moved to separate script)

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Generating sqlc files for all modules...${NC}"

# Find all sqlc.yaml files in modules and platform packages and generate code
module_count=0
for config_file in $(find internal/modules internal/platform -name "sqlc.yaml" -type f); do
    config_dir=$(dirname "$config_file")
    module_name=$(basename "$(dirname "$(dirname "$config_dir")")")
    
    echo "  [$((++module_count))] Generating sqlc for module: $module_name ($config_dir)"
    
    # Change to the config directory and run sqlc generate
    if (cd "$config_dir" && sqlc generate); then
        echo "      ✓ Successfully generated sqlc for $module_name"
    else
        echo "      ✗ Failed to generate sqlc for $module_name"
        exit 1
    fi
done

if [ $module_count -eq 0 ]; then
    echo "  No modules with sqlc.yaml found!"
    exit 1
fi

echo -e "${GREEN}Successfully generated sqlc code for $module_count modules!${NC}"