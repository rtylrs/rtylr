#!/bin/bash

# Database Reset Script for Onion Project
# This script resets the database for development purposes

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration follows .env / the same defaults as the application.
DB_NAME="${DB_NAME:-onion_dev}"
DB_USER="${DB_USER:-onion_user}"
DB_PASSWORD="${DB_PASSWORD:-onion_password}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"

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
mysql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    -e "DROP DATABASE IF EXISTS \`$DB_NAME\`;" 2>/dev/null || echo "Database drop failed (might not exist)"

echo -e "${BLUE}Creating database...${NC}"
mysql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    -e "CREATE DATABASE \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo -e "${BLUE}Running migrations...${NC}"

# Use the canonical runner so reset and deploy apply the same ordered schema
# set. It also stops on failures and tracks applied files.
DB_HOST="$DB_HOST" DB_PORT="$DB_PORT" DB_USER="$DB_USER" \
DB_PASSWORD="$DB_PASSWORD" DB_NAME="$DB_NAME" \
    bash "$PROJECT_ROOT/scripts/migrate.sh"

echo -e "${GREEN}✓ Database reset completed successfully!${NC}"
echo -e "${BLUE}Database DSN: $DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME?parseTime=true${NC}"
