#!/bin/bash

# Simple setup script for Onion
set -e

ENV_FILE=".env"
EXAMPLE_FILE=".env.example"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${BLUE}🐳 Setting up Onion project${NC}"

# Ensure env file exists
if [ ! -f "$ENV_FILE" ]; then
    cp "$EXAMPLE_FILE" "$ENV_FILE"
    echo -e "${GREEN}Created $ENV_FILE from $EXAMPLE_FILE${NC}"
    echo "Fill in the placeholder values in $ENV_FILE before running the app."
fi

echo "Installing buf..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install bufbuild/tap/buf
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    curl -sSL \
        "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
        -o "$(go env GOPATH)/bin/buf" && \
    chmod +x "$(go env GOPATH)/bin/buf"
fi

# Start database
echo -e "${BLUE}Starting PostgreSQL...${NC}"
docker-compose up -d postgres

# Wait for database
until docker-compose exec -T postgres pg_isready -U ${DB_USER:-onion_user} -d ${DB_NAME:-onion_dev}; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

echo -e "${GREEN}✓ PostgreSQL is ready${NC}"

# Generate sqlc code if available
if command -v sqlc >/dev/null 2>&1; then
    sqlc generate
fi

echo -e "${GREEN}Setup completed${NC}"
