.PHONY: clean generate install-onek install-js typecheck

# Onekit v0.19.0: latest release; main CI regenerates and commits output.
ONK_VERSION := 627f76e4fd75d45275833e02e483b276af97ddf7

# Clean generated code
clean:
	rm -rf gen

# Install the onek CLI used to generate Go/TypeScript from api/*.onk contracts.
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Install JS codegen tooling
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go, gen/ts, and gen/openapi from api/*.onk. Output is committed.
# onekit.toml sets api/ as the schema root, so generated import paths
# mirror the schema path exactly (gen/go/hub/business/v1, no "api/" prefix) -
# matching what every consumer already imports from the pre-onk proto days.
generate: install-onek install-js
	rm -rf gen/go gen/ts
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
