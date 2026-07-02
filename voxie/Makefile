.PHONY: clean lint format breaking generate install-plugins fetch-onekit-proto

ONEKIT_VERSION := v0.2.0

# Clean generated code
clean:
	rm -rf gen

# Fetch onekit's HTTP annotation .proto files for a pinned version.
# Not vendored/committed — fetched on demand, same as installing the plugin binary.
fetch-onekit-proto:
	rm -rf .onekit
	mkdir -p .onekit
	buf export "https://github.com/1homsi/onekit.git#tag=$(ONEKIT_VERSION),subdir=proto" -o .onekit/proto

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
	go install github.com/1homsi/onekit/cmd/protoc-gen-onekit-go-http@$(ONEKIT_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Regenerate gen/go from proto sources. Output is committed to the repo.
generate: install-plugins fetch-onekit-proto
	buf dep update
	rm -rf gen/go
	buf generate
