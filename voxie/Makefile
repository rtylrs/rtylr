.PHONY: clean lint format breaking

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
