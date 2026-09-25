#!/bin/bash

# Simple setup script for Onion
set -e

ENV_FILE=".env"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Setting up Onion project${NC}"

# Ensure env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}$ENV_FILE not found.${NC}"
    echo "Create $ENV_FILE and configure DATABASE_URL or DB_* before running the app."
fi

if command -v mysqladmin >/dev/null 2>&1; then
    echo -e "${BLUE}Checking MySQL connectivity...${NC}"

    until mysqladmin ping \
        --host="${DB_HOST:-localhost}" \
        --port="${DB_PORT:-3306}" \
        --user="${DB_USER:-onion_user}" \
        --password="${DB_PASSWORD:-onion_password}" \
        --silent >/dev/null 2>&1; do
        echo "Waiting for MySQL..."
        sleep 2
    done

    echo -e "${GREEN}MySQL is ready${NC}"
else
    echo -e "${YELLOW}mysqladmin is not installed; skipping database connectivity check${NC}"
fi

# Generate sqlc code if available
if command -v sqlc >/dev/null 2>&1; then
    sqlc generate
fi

echo -e "${GREEN}Setup completed${NC}"
