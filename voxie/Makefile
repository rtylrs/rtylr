.PHONY: clean generate install-onek install-js typecheck

# Onekit CLI version. Generated output is gitignored; see `make generate`.
ONK_VERSION := v0.19.0

# `go install` writes to GOPATH/bin, which is not guaranteed to be on PATH here.
GOBIN_DIR := $(shell go env GOPATH)/bin
export PATH := $(GOBIN_DIR):$(PATH)

# Clean generated code
clean:
	rm -rf gen

# Install the onek CLI used to generate Go/TypeScript from api/*.onk contracts.
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Install JS codegen tooling
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go, gen/ts, and gen/openapi from api/*.onk. Output is gitignored.
# onekit.toml sets api/ as the schema root, so generated import paths
# mirror the schema path exactly (gen/go/hub/business/v1, no "api/" prefix) -
# matching what every consumer already imports from the pre-onk proto days.
generate: install-onek
	rm -rf gen
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
