.PHONY: clean generate install-onkt typecheck

# The onkt CLI, its TypeSpec emitter (@onkt/typespec) and the TypeSpec compiler
# are pinned in package.json/package-lock.json, so every machine and CI job
# generates with the same versions. Keep @onkt/cli and @onkt/typespec on the
# same version: they are released together. The @onkt/cli launcher downloads
# the release binary on first use and checks it against the release's
# checksums.txt. Generated output is gitignored; see `make generate`.

# Go/OpenAPI stay next to the contracts; the TypeScript half belongs to the
# clients workspace package that publishes it.
CONTRACTS_DIR := ../clients/packages/contracts

# Clean generated code
clean:
	rm -rf gen $(CONTRACTS_DIR)/src

# Install the pinned TypeSpec toolchain and onkt CLI (needs Node.js 22+).
install-onkt:
	npm ci --no-audit --no-fund

# Regenerate gen/go, gen/openapi and the contracts package's TypeScript from
# api/**/*.tsp. Output is gitignored. onkt.toml sets api/ as the schema root, so
# generated import paths mirror the schema path exactly (gen/go/hub/business/v1,
# no "api/" prefix) - matching what every consumer already imports.
# onkt requires output inside this project, so the TypeScript half is copied
# into the clients workspace package that exposes it.
generate: install-onkt
	rm -rf gen $(CONTRACTS_DIR)/src
	python3 scripts/generate_validation_metadata.py --proto-root api --out-dir gen/ts/validation
	npx --no-install onkt build
	mkdir -p $(CONTRACTS_DIR)/src
	cp -R gen/ts/. $(CONTRACTS_DIR)/src/

# Typecheck the generated TypeScript through its workspace package.
typecheck:
	pnpm -C ../clients --filter @voxire/voxie typecheck
