#!/bin/bash
set -xe

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
go build -o ./bin/api ./cmd/api/main.go

golangci-lint run       
