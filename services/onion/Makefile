.PHONY: onion setup build build-all build-media-engine generate generate-module lint test test-db direct-authorization-evidence role-access-control-v2-rollback-check route-contract-check json-tags-check mutation-validation-check run run-api run-upload dev sql-check sql-case-check db-prelaunch-check local-env-up local-env-down local-env-logs local-env-mindsdb-init

onion: sql-check test build
	@echo "✅ Onion project is ready!"

setup:
	./scripts/setup.sh

build:
	./scripts/build.sh

# Build all service binaries
build-all:
	./scripts/check-voxie-ref.sh
	@echo "Building API binary..."
	@mkdir -p ./bin
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	 COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	 BUILD_TIME=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	 go build \
	   -ldflags="-s -w \
	     -X 'github.com/voxire/onion/internal/platform/server/http.Version=$$VERSION' \
	     -X 'github.com/voxire/onion/internal/platform/server/http.Commit=$$COMMIT' \
	     -X 'github.com/voxire/onion/internal/platform/server/http.BuildTime=$$BUILD_TIME'" \
	   -o ./bin/api ./cmd/api/main.go
	@echo "All binaries built successfully!"

# Build the C++ media-engine binary the upload runmode shells out to.
# Requires libvips + cmake (see media-engine/Dockerfile for the full list).
build-media-engine:
	cmake -S media-engine -B media-engine/build -DCMAKE_BUILD_TYPE=Release
	cmake --build media-engine/build --parallel

generate:
	./scripts/generate_sqlc.sh

generate-module:
	@test -n "$(NAME)" || (echo "usage: make generate-module NAME=<module_name>" && exit 1)
	./scripts/generate-module.sh $(NAME)

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh

# Run the database-backed integration tests against the local-env MySQL
# container. Resets a dedicated onion_test DB (drop → migrate → seed) each
# run so the tests stay repeatable. Requires `make local-env-up` first.
# Pass a package subset with PKG=, e.g. `make test-db PKG=./internal/modules/pos_users/...`
test-db:
	./scripts/test-db.sh $(PKG)

direct-authorization-evidence:
	./scripts/run-role-access-control-v2-direct-coverage.sh

role-access-control-v2-rollback-check:
	./scripts/test-role-access-control-v2-rollback.sh

route-contract-check:
	python3 ./scripts/check-route-contract-drift.py

json-tags-check:
	python3 ./scripts/check-json-tags.py

mutation-validation-check:
	python3 ./scripts/check-mutation-validation.py

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

local-env-mindsdb-init:
	./local-env/mindsdb/bootstrap.sh

sql-check:
	./scripts/check_sql_case.sh

load-tables:
	./scripts/load_tables.sh

db-prelaunch-check:
	./scripts/db-prelaunch-check.sh

clean-generated:
	./scripts/clean_generated.sh
