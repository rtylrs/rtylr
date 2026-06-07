.PHONY: generate clean lint

# Generate Go + TypeScript code from proto files
generate:
	@echo "Generating code from proto files..."
	buf generate
	python3 scripts/generate_validation_metadata.py
	@echo "Done. Output in gen/go and gen/ts"

# Clean generated code
clean:
	rm -rf gen/go gen/ts gen/validation
	mkdir -p gen/go gen/ts gen/validation

# Lint proto files
lint:
	buf lint

# Format proto files
format:
	buf format -w

# Check for breaking changes
breaking:
	buf breaking --against '.git#branch=main'
