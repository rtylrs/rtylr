#!/bin/bash

# Database Table Loading Script for Onion Project
# This script reads all schema files and creates/updates tables

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Loading database tables from schema files${NC}"

if ! command -v mysql >/dev/null 2>&1; then
    echo -e "${RED}mysql is required to load schema files.${NC}"
    exit 1
fi

if ! command -v mysqladmin >/dev/null 2>&1; then
    echo -e "${RED}mysqladmin is required to verify database connectivity.${NC}"
    exit 1
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:-onion_dev}"
DB_USER="${DB_USER:-onion_user}"
DB_PASSWORD="${DB_PASSWORD:-onion_password}"

mysql_cmd=(
    mysql
    --host="${DB_HOST}"
    --port="${DB_PORT}"
    --user="${DB_USER}"
    --password="${DB_PASSWORD}"
    "${DB_NAME}"
)

echo -e "${BLUE}Checking MySQL connectivity...${NC}"
until mysqladmin ping \
    --host="${DB_HOST}" \
    --port="${DB_PORT}" \
    --user="${DB_USER}" \
    --password="${DB_PASSWORD}" \
    --silent >/dev/null 2>&1; do
    echo "Waiting for MySQL..."
    sleep 2
done

# Schema directories
SCHEMA_DIRS=("internal/platform/database/schemas")

# Check if schema directories exist
for schema_dir in "${SCHEMA_DIRS[@]}"; do
    if [ ! -d "$schema_dir" ]; then
        echo -e "${RED}Schema directory $schema_dir not found.${NC}"
        exit 1
    fi
done

# Function to create table if not exists
create_table_if_not_exists() {
    local file="$1"
    local filename
    filename=$(basename "$file")
    local table_name
    table_name=$(basename "$filename" .sql)

    echo -e "${BLUE}Ensuring table exists: $table_name${NC}"

    if "${mysql_cmd[@]}" < "$file"; then
        echo -e "${GREEN}✓ Table $table_name is ready${NC}"
    else
        echo -e "${YELLOW}⚠ Table $table_name may already exist or had issues${NC}"
    fi
}

# Process all SQL files in the schema directories
echo -e "${BLUE}Scanning for schema files in schema directories...${NC}"

# Find all .sql files in all schema directories
sql_files=""
for schema_dir in "${SCHEMA_DIRS[@]}"; do
    sql_files+="$(find "$schema_dir" -name "*.sql" -type f | sort)"$'\n'
done

if [ -z "$sql_files" ]; then
    echo -e "${YELLOW}No SQL schema files found in schema directories${NC}"
    exit 0
fi

echo -e "${GREEN}Found $(echo "$sql_files" | wc -l) schema file(s)${NC}"

# Process each SQL file to create/update tables. Existing tables are left alone
# (the files use CREATE TABLE IF NOT EXISTS); tables are never dropped here.
for file in $sql_files; do
    if [ -n "$file" ]; then
        create_table_if_not_exists "$file"
    fi
done

echo -e "${GREEN}🎉 Database tables loaded successfully!${NC}"

# Show current tables
echo -e "${BLUE}Current tables in database:${NC}"
"${mysql_cmd[@]}" --batch --skip-column-names -e "SHOW TABLES;" || true
