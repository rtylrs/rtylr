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

echo -e "${BLUE}🗄️  Loading database tables from schema files${NC}"

# Check if Docker Compose is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo -e "${YELLOW}PostgreSQL container is not running. Starting it...${NC}"
    docker-compose up -d postgres
    
    # Wait for PostgreSQL to be ready
    echo -e "${BLUE}Waiting for PostgreSQL to be ready...${NC}"
    until docker-compose exec -T postgres pg_isready -U onion_user -d onion_dev; do
        echo "Waiting for PostgreSQL..."
        sleep 2
    done
fi

# Schema directories
SCHEMA_DIRS=("db/schemas" "internal/platform/database/schemas")

# Check if schema directories exist
for schema_dir in "${SCHEMA_DIRS[@]}"; do
    if [ ! -d "$schema_dir" ]; then
        echo -e "${RED}Schema directory $schema_dir not found.${NC}"
        exit 1
    fi
done

# Function to drop tables not in schema
drop_orphaned_tables() {
    local schema_tables="$1"
    local DB_USER="onion_user"
    local DB_NAME="onion_dev"
    
    echo -e "${BLUE}Checking for orphaned tables...${NC}"
    
    # Get list of existing tables in the database
    local existing_tables=$(docker-compose exec -T postgres psql -U $DB_USER -d $DB_NAME -t -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';" | tr -d ' ')
    
    # Check each existing table
    for table in $existing_tables; do
        # Skip if table is in our schema list
        if echo "$schema_tables" | grep -q "^$table$"; then
            echo -e "${GREEN}✓ Table $table is in schema${NC}"
        else
            echo -e "${YELLOW}🗑️  Dropping orphaned table: $table${NC}"
            docker-compose exec -T postgres psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS $table CASCADE;"
        fi
    done
}

# Function to create table if not exists
create_table_if_not_exists() {
    local file="$1"
    local filename=$(basename "$file")
    local table_name=$(basename "$filename" .sql)
    
    echo -e "${BLUE}Ensuring table exists: $table_name${NC}"
    
    # Use hardcoded DB credentials
    local DB_USER="onion_user"
    local DB_NAME="onion_dev"
    
    # Read the schema file content and execute it directly
    if docker-compose exec -T postgres psql -U $DB_USER -d $DB_NAME -c "$(cat "$file")"; then
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

# Collect schema table names
schema_tables=""
for file in $sql_files; do
    if [ -n "$file" ]; then
        table_name=$(basename "$file" .sql)
        schema_tables="$schema_tables$table_name"$'\n'
    fi
done

# Drop orphaned tables first
drop_orphaned_tables "$schema_tables"

# Process each SQL file to create/update tables
for file in $sql_files; do
    if [ -n "$file" ]; then
        create_table_if_not_exists "$file"
    fi
done

echo -e "${GREEN}🎉 Database tables loaded successfully!${NC}"

# Show current tables
echo -e "${BLUE}Current tables in database:${NC}"
docker-compose exec -T postgres psql -U onion_user -d onion_dev -c "\dt" || true