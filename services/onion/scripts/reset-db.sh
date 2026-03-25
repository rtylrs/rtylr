#!/bin/bash

# Database Reset Script for Onion Project
# This script resets the database for development purposes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DB_NAME="onion_dev"
DB_USER="onion_user"
DB_PASSWORD="onion_password"
DB_HOST="localhost"
DB_PORT="5432"

echo -e "${YELLOW}⚠️  Database Reset Script${NC}"
echo -e "${YELLOW}This will DROP and RECREATE the database. All data will be lost!${NC}"

read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Database reset cancelled.${NC}"
    exit 0
fi

echo -e "${BLUE}🔄 Resetting database...${NC}"

# Drop and recreate database
echo -e "${BLUE}Dropping database...${NC}"
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/postgres" -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || echo "Database drop failed (might not exist)"

echo -e "${BLUE}Creating database...${NC}"
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/postgres" -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"

echo -e "${BLUE}Running migrations...${NC}"

# Apply schema files
for schema_file in db/schemas/*.sql; do
    if [ -f "$schema_file" ]; then
        echo "Applying schema: $(basename "$schema_file")"
        psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -f "$schema_file"
    fi
done

echo -e "${GREEN}✓ Database reset completed successfully!${NC}"
echo -e "${BLUE}Database URL: postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME${NC}" 