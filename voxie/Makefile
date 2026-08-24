.PHONY: clean generate install-onek install-js typecheck

# v0.8.0: frontend platform release - opt-in [generate.ts-client] artifacts
# (zod schemas, React Query hooks, MSW handlers) and the stdlib `onek mock`
# dev server. No output changes for our current config; flags stay off until
# a consumer needs them. See 1homsi/onekit@e526fb5.
ONK_VERSION := 6186effb3818a5dd8cbd4668b285c843b69b5e1a

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
