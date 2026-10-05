.PHONY: onion setup build build-all build-media-engine generate contracts generate-module lint vet architecture-check test test-db authorization-evidence route-contract-check json-tags-check mutation-validation-check run run-api run-upload dev sql-check sql-case-check db-prelaunch-check local-env-up local-env-down local-env-logs local-env-mindsdb-init

onion: sql-check test build
	@echo "✅ Onion project is ready!"

setup:
	./scripts/setup.sh

# SQLC output is gitignored, so a checkout advanced across commits keeps
# whatever was generated for the older revision. Every target below that
# compiles Go for verification takes `generate` as a prerequisite, so stale
# output cannot reach the compiler. Make builds a prerequisite at most once
# per invocation, so `make onion` still generates a single time.
#
# build.sh generates again on its own so it stays self-contained when invoked
# directly. The repeat costs about seven seconds against a multi-minute target,
# which is worth keeping both entry points correct on their own terms.
build: generate contracts
	./scripts/build.sh

# Build all service binaries
build-all: generate contracts
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
	   -o ./bin/api ./cmd/api
	@echo "All binaries built successfully!"

# Build the C++ media-engine binary the upload runmode shells out to.
# Requires libvips + cmake (see media-engine/Dockerfile for the full list).
build-media-engine:
	cmake -S media-engine -B media-engine/build -DCMAKE_BUILD_TYPE=Release
	cmake --build media-engine/build --parallel

generate:
	./scripts/generate_sqlc.sh

# Voxie's generated Go output is gitignored; regenerate before Go compiles.
contracts:
	$(MAKE) -C ../../voxie generate

generate-module:
	@test -n "$(NAME)" || (echo "usage: make generate-module NAME=<module_name>" && exit 1)
	./scripts/generate-module.sh $(NAME)

lint: generate contracts
	./scripts/lint.sh

vet: generate contracts
	go vet ./...

architecture-check:
	go test . -run '^TestArchitecture($$|Scanner$$)' -count=1

test: generate contracts
	./scripts/test.sh

# Run the database-backed integration tests against the local-env MySQL
# container. Resets a dedicated onion_test DB (drop → migrate → seed) each
# run so the tests stay repeatable. Requires `make local-env-up` first.
# Pass a package subset with PKG=, e.g. `make test-db PKG=./internal/modules/pos/users/...`
test-db: generate contracts
	./scripts/test-db.sh $(PKG)

authorization-evidence:
	./scripts/run-authorization-evidence.sh

route-contract-check:
	python3 ./scripts/check-route-contract-drift.py

json-tags-check:
	python3 ./scripts/check-json-tags.py

mutation-validation-check:
	python3 ./scripts/check-mutation-validation.py

# Default run (backwards compatibility)
run:
	go run ./cmd/api

# Run specific services
run-api:
	go run ./cmd/api

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

db-prelaunch-check:
	./scripts/db-prelaunch-check.sh

clean-generated:
	./scripts/clean_generated.sh
