#!/bin/bash

# Clean Generated Files Script for Onion Project
# This script removes generated SQLC files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Cleaning generated files...${NC}"

# Function to count files before deletion
count_files() {
    local pattern="$1"
    local count=$(find . -path "$pattern" -type f 2>/dev/null | wc -l)
    echo $count
}

list_sqlc_output_dirs() {
    find . -name "sqlc" -type d -print
}

# Count files before deletion
echo -e "${BLUE}Counting files to be removed...${NC}"
sqlc_count=$(find . -name "dbmodels" -type d | wc -l)
sqlc_go_count=$(find . -name "*.sql.go" -type f 2>/dev/null | wc -l)
persistence_count=$(list_sqlc_output_dirs | wc -l)

total_count=$((sqlc_count + sqlc_go_count + persistence_count))

if [ $total_count -eq 0 ]; then
    echo -e "${YELLOW}No generated files found to clean.${NC}"
    exit 0
fi

echo -e "${BLUE}Found generated items:${NC}"
echo -e "  • SQLC dbmodels folders: ${sqlc_count}"
echo -e "  • SQLC Go files: ${sqlc_go_count} files"
echo -e "  • Persistence sqlc folders: ${persistence_count}"
echo -e "  • Total items: ${total_count}"

echo ""
echo -e "${YELLOW}⚠️  This will remove all generated files.${NC}"
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Operation cancelled.${NC}"
    exit 0
fi

echo -e "${BLUE}Removing generated files...${NC}"

# Remove SQLC generated files
if [ $sqlc_count -gt 0 ]; then
    echo -e "${BLUE}Removing SQLC dbmodels...${NC}"
    find . -name "dbmodels" -type d -exec rm -rf {} \;
fi

if [ $sqlc_go_count -gt 0 ]; then
    echo -e "${BLUE}Removing SQLC Go files...${NC}"
    find . -name "*.sql.go" -type f -delete
fi

# Remove persistence files
if [ $persistence_count -gt 0 ]; then
    echo -e "${BLUE}Removing persistence files...${NC}"
    while IFS= read -r sqlc_dir; do
        rm -rf "$sqlc_dir"
    done < <(list_sqlc_output_dirs)
fi

# Remove empty directories (cleanup any remaining empty dirs)
echo -e "${BLUE}Cleaning up any remaining empty directories...${NC}"
find . -type d -empty -delete 2>/dev/null || true

# Verify the generated sqlc directories are really gone: a failed rm (read-only
# mount, root-owned files) must fail the script, not be reported as success.
remaining=$(list_sqlc_output_dirs | wc -l)
if [ "$remaining" -ne 0 ]; then
    echo -e "${RED}❌ ${remaining} generated sqlc directories could not be removed:${NC}" >&2
    list_sqlc_output_dirs >&2
    exit 1
fi

echo -e "${GREEN}✅ Generated files cleaned successfully!${NC}"
echo ""
echo -e "${BLUE}To regenerate the files, run:${NC}"
echo -e "  • SQLC: ${GREEN}make generate${NC}"
echo ""
echo -e "${YELLOW}Note: You'll need to regenerate these files before building the project.${NC}"
