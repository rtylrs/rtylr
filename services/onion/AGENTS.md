# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

Refer to [CLAUDE.md](CLAUDE.md) for architecture overview, commands, and project structure.

Refer to [docs/creating-endpoints.md](docs/creating-endpoints.md) for the full flow to create an endpoint from proto to handler.

## Quick Reference

### Commands

```bash
make build          # Full build pipeline
make run            # Run API server (port 9999)
make dev            # Dev mode with live reload
make proto          # Update grimoire submodule + regenerate protos
make generate       # Regenerate SQLC code
make test           # Run tests
```

### Key Patterns

- Modules follow Clean Architecture: `domain/` → `application/` → `infrastructure/` → `routes/`
- Shared modules live in `internal/modules/`
- Platform-specific modules live in `internal/platforms/<platform>/`
- Proto definitions are in the `grimoire` submodule — do not edit `proto/` directly
- Each module has its own `sql/sqlc.yaml` for SQLC generation
- The module registry auto-discovers routes via reflection
- Auth uses `TokenValidator` interface — platform-agnostic

### Example Module

Use `internal/modules/config/` as the reference implementation.
