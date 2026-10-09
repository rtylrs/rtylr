.PHONY: clean generate install-onek typecheck

# Onekit CLI version. Generated output is gitignored; see `make generate`.
ONK_VERSION := v0.30.0

# `go install` writes to GOPATH/bin, which is not guaranteed to be on PATH here.
GOBIN_DIR := $(shell go env GOPATH)/bin
export PATH := $(GOBIN_DIR):$(PATH)

# Go/OpenAPI stay next to the contracts; the TypeScript half belongs to the
# clients workspace package that publishes it.
CONTRACTS_DIR := ../clients/packages/contracts

# Clean generated code
clean:
	rm -rf gen $(CONTRACTS_DIR)/src

# Install the onek CLI used to generate Go/TypeScript from api/*.onk contracts.
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Regenerate gen/go, gen/openapi and the contracts package's TypeScript from
# api/*.onk. Output is gitignored. onekit.toml sets api/ as the schema root, so
# generated import paths mirror the schema path exactly (gen/go/hub/business/v1,
# no "api/" prefix) - matching what every consumer already imports.
# onek requires output inside this project, so the TypeScript half is copied
# into the clients workspace package that exposes it.
generate: install-onek
	rm -rf gen $(CONTRACTS_DIR)/src
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	onek build .
	mkdir -p $(CONTRACTS_DIR)/src
	cp -R gen/ts/. $(CONTRACTS_DIR)/src/

# Typecheck the generated TypeScript through its workspace package.
typecheck:
	pnpm -C ../clients --filter @voxire/voxie typecheck
