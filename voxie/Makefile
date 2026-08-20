.PHONY: clean generate install-onek install-js typecheck

# v0.6.0: fixes optional int64/uint64 fields (e.g. project_id, task_id on
# flow.time.v1.CreateWorkDayLineRequest) never getting the wire-string
# MarshalJSON/UnmarshalJSON treatment that required int64 fields already get -
# onek's needsInt64StringEncoding excluded any field with f.Optional set, so
# every consumer sending those fields as strings (the platform-wide int64
# convention documented in this repo's own CLAUDE.md) got a 400 back from Go's
# plain json.Unmarshal expecting a number. See internal/gengo/jsonmapping.go
# in 1homsi/onekit@7011894.
ONK_VERSION := 70118946d1c20146cd7f720365ef6ef11d8066f6

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
	onek build api

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
