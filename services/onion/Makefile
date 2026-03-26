.PHONY: onion setup build build-all generate lint test run run-api run-worker run-pay dev dev-api dev-worker dev-pay sql-check sql-case-check

onion: sql-check test build
	@echo "✅ Onion project is ready!"

setup:
	./scripts/setup.sh

build:
	./scripts/build.sh

# Build all service binaries
build-all:
	@echo "🔨 Building all services..."
	@mkdir -p ./bin
	go build -o ./bin/api ./cmd/api/main.go
	go build -o ./bin/worker ./cmd/worker/main.go
	go build -o ./bin/pay ./cmd/pay/main.go
	@echo "✅ All services built successfully!"

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

run-worker:
	go run ./cmd/worker/main.go

run-pay:
	go run ./cmd/pay/main.go

# Development with live reloading (air)
dev:
	air -c .air/api.toml

dev-worker:
	air -c .air/worker.toml

dev-pay:
	air -c .air/pay.toml

sql-check:
	./scripts/check_sql_case.sh

load-tables:
	./scripts/load_tables.sh

clean-generated:
	./scripts/clean_generated.sh
