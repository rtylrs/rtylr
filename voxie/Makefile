.PHONY: clean lint format breaking generate install-plugins fetch-onekit-proto install-js typecheck

ONEKIT_VERSION := v0.3.0

# Clean generated code
clean:
	rm -rf gen

# Fetch onekit's HTTP annotation .proto files for a pinned version.
# Not vendored/committed — fetched on demand, same as installing the plugin binary.
fetch-onekit-proto:
	rm -rf .onekit
	mkdir -p .onekit
	buf export "https://github.com/stackxio/onekit.git#tag=$(ONEKIT_VERSION),subdir=proto" -o .onekit/proto

# Lint proto files
lint: fetch-onekit-proto
	buf lint

# Format proto files
format:
	buf format -w

# Check for breaking changes
breaking: fetch-onekit-proto
	buf breaking --against '.git#branch=main'

# Install the onekit protoc plugins used for generation
install-plugins:
	go install github.com/stackxio/onekit/cmd/protoc-gen-onekit-go-http@$(ONEKIT_VERSION)
	go install github.com/stackxio/onekit/cmd/protoc-gen-onekit-ts-client@$(ONEKIT_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install JS codegen tooling (ts-proto)
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go and gen/ts from proto sources. Output is committed to the repo.
generate: install-plugins install-js fetch-onekit-proto
	buf dep update
	rm -rf gen/go gen/ts
	buf generate
	python3 scripts/generate_validation_metadata.py --proto-root proto --out-dir gen/ts/validation

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
