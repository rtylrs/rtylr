.PHONY: onion setup build build-all generate lint test run run-api dev sql-check sql-case-check local-env-up local-env-down local-env-logs

onion: sql-check test build
	@echo "✅ Onion project is ready!"

setup:
	./scripts/setup.sh

build:
	./scripts/build.sh

# Build all service binaries
build-all:
	@echo "🔨 Building API binary..."
	@mkdir -p ./bin
	go build -o ./bin/api ./cmd/api/main.go
	@echo "✅ API binary built successfully!"

generate:
	./scripts/generate_sqlc.sh

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh

# Default run (backwards compatibility)
run:
	go run ./cmd/api/main.go

# Run specific services
run-api:
	go run ./cmd/api/main.go

# Development with live reloading (air)
dev:
	air -c .air/api.toml

local-env-up:
	docker compose -p onion-env -f local-env/docker-compose.yml up -d

local-env-down:
	docker compose -p onion-env -f local-env/docker-compose.yml down

local-env-logs:
	docker compose -p onion-env -f local-env/docker-compose.yml logs -f

sql-check:
	./scripts/check_sql_case.sh

load-tables:
	./scripts/load_tables.sh

clean-generated:
	./scripts/clean_generated.sh
