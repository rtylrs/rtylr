.PHONY: clean generate install-onek install-js typecheck

# Onekit generates OpenAPI 3.1 specs in both YAML and JSON.
ONK_VERSION := 3bbec1c808ffd381f4b35d3e2b427d283d677a0f

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
