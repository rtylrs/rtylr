#!/bin/bash
set -xe

# Voxie's generated Go output is gitignored. Regenerate it so Go compiles
# against the current .onk contracts, not a stale local generation.
make -C ../../voxie generate

for dir in ./db/dbmodels/*/; do
 [ -d "$dir" ] || continue
 find "$dir" -type f ! -name 'queries.go' ! -name 'utils.go' ! -name 'queries_input.go' -delete
done

# Check sql case
./scripts/check_sql_case.sh

# Generate SQL code for each module (includes SQL case checking)
./scripts/generate_sqlc.sh

go vet ./...

# Build all service binaries
mkdir -p ./bin
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build \
  -ldflags="-s -w \
    -X 'github.com/voxire/onion/internal/platform/server/http.Version=${VERSION}' \
    -X 'github.com/voxire/onion/internal/platform/server/http.Commit=${COMMIT}' \
    -X 'github.com/voxire/onion/internal/platform/server/http.BuildTime=${BUILD_TIME}'" \
  -o ./bin/api ./cmd/api

golangci-lint run
