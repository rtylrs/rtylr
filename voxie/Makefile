.PHONY: clean generate install-onek install-js typecheck

ONK_VERSION := v0.4.4

# Clean generated code
clean:
	rm -rf gen

# Install the onek CLI used to generate Go/TypeScript from api/*.onk contracts.
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Install JS codegen tooling
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go and gen/ts from api/*.onk. Output is committed to the repo.
generate: install-onek install-js
	rm -rf gen/go gen/ts
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
