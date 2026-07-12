.PHONY: clean lint format breaking generate install-plugins install-onek install-js typecheck

ONEKIT_VERSION := v0.3.5
ONK_VERSION := v0.4.2

# Clean generated code
clean:
	rm -rf gen

# Lint proto files
lint:
	buf lint

# Format proto files
format:
	buf format -w

# Check for breaking changes
breaking:
	buf breaking --against '.git#branch=main'

# Install the onekit protoc plugins used for services still on .proto (see proto/).
# Being migrated off in favor of onek/.onk - see api/ and README.md.
install-plugins:
	go install github.com/corezio/onekit/cmd/protoc-gen-onekit-go-http@$(ONEKIT_VERSION)
	go install github.com/corezio/onekit/cmd/protoc-gen-onekit-ts-client@$(ONEKIT_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install the onek CLI used for services migrated to .onk (see api/).
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Install JS codegen tooling (ts-proto)
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go and gen/ts from ALL schema sources: proto/ (buf, being
# migrated off) and api/ (onek, the target). Output is committed to the repo.
generate: install-plugins install-onek install-js
	buf dep update
	rm -rf gen/go gen/ts
	buf generate
	python3 scripts/generate_validation_metadata.py --proto-root proto --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
