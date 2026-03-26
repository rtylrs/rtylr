#!/bin/bash

# Simple setup script for Onion
set -e

ENV_FILE=".env"
EXAMPLE_FILE=".env.example"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Setting up Onion project${NC}"

# Ensure env file exists
if [ ! -f "$ENV_FILE" ]; then
    cp "$EXAMPLE_FILE" "$ENV_FILE"
    echo -e "${GREEN}Created $ENV_FILE from $EXAMPLE_FILE${NC}"
    echo "Fill in the placeholder values in $ENV_FILE before running the app."
fi

if command -v pg_isready >/dev/null 2>&1; then
    echo -e "${BLUE}Checking PostgreSQL connectivity...${NC}"

    if [ -n "${DATABASE_URL:-}" ]; then
        until pg_isready -d "${DATABASE_URL}"; do
            echo "Waiting for PostgreSQL..."
            sleep 2
        done
    else
        until pg_isready \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "${DB_USER:-onion_user}" \
            -d "${DB_NAME:-onion_dev}"; do
            echo "Waiting for PostgreSQL..."
            sleep 2
        done
    fi

    echo -e "${GREEN}PostgreSQL is ready${NC}"
else
    echo -e "${YELLOW}pg_isready is not installed; skipping database connectivity check${NC}"
fi

# Generate sqlc code if available
if command -v sqlc >/dev/null 2>&1; then
    sqlc generate
fi

echo -e "${GREEN}Setup completed${NC}"
