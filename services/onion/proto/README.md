# Grimoire

A centralized repository for Protocol Buffer definitions shared across Voxire projects.

## Overview

This repository contains all the `.proto` files used across Voxire projects, providing a single source of truth for API contracts and data models.

## Structure

```
grimoire/
├── proto/                  # Source .proto definitions
│   ├── api/
│   │   ├── account/v1/     # Auth, profile, registration
│   │   ├── config/v1/      # Site config, news, FAQ, themes, careers
│   │   ├── core/v1/        # Core services (e.g., Slack logging)
│   │   └── snips/v1/       # AI chat, memory, translation
│   ├── common/v1/          # Shared types (errors, pagination, enums)
│   └── tooling/v1/         # Internal tooling (vox users, permissions)
├── generated/              # Generated code outputs (git-ignored except structure)
│   └── go/
├── buf.yaml                # Buf lint/module config
└── buf.gen.yaml            # Code generation config
```

## Usage

### As a Git Submodule

Add this repository as a submodule to your project. In this example, we add it under `grimoire/`:

```bash
git submodule add https://github.com/Voxire/grimoire.git grimoire
```

The `.proto` files are located at `grimoire/proto/` in your project.

### Generate Code

From this repository (or from the submodule directory in your project), run:

```bash
# Install buf if not already installed
go install github.com/bufbuild/buf/cmd/buf@latest

# Lint .proto files
buf lint

# Generate code as configured in buf.gen.yaml (e.g., Go to generated/go)
buf generate
```

### Updating

To update the submodule to the latest version:

```bash
git submodule update --remote grimoire
```

## Projects Using Grimoire

- [onion](https://github.com/Voxire/onion) - Go backend service
- Additional projects will be listed here as they adopt grimoire

## Contributing

When adding new proto files or modifying existing ones:

1. Make your changes to the `.proto` files
2. Test the changes by generating code
3. Commit and push your changes
4. Update dependent projects to use the new version
