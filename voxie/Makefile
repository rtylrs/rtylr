.PHONY: clean generate install-onek install-js typecheck

# v0.7.2: introduces schema_root so api/*.onk can stay in a subdirectory while
# outputs anchor at the repo root (v0.7.1's output-containment check rejected
# the previous "../gen/go" out paths). Also carries v0.7.0-v0.7.1 generator
# fixes: errors.Is/errors.As matching and usage-gated validation helpers in Go,
# guarded JSON.parse + escaped validation messages in TS, epoch-zero timestamp
# round-trip for non-optional @encode(unix_*). See 1homsi/onekit@5d18b2f.
ONK_VERSION := 5d18b2f5c869c204a34a03be5b45f7cdb6c3fbb5

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
# onekit.toml lives in api/ (the schema root), so generated import paths
# mirror the schema path exactly (gen/go/hub/business/v1, no "api/" prefix) -
# matching what every consumer already imports from the pre-onk proto days.
generate: install-onek install-js
	rm -rf gen/go gen/ts
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
