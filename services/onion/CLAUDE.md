# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

```bash
make build          # Full build pipeline (sql check + sqlc + vet + build + lint)
make run            # Run API server (port 9999)
make run-worker     # Run worker server (port 9998)
make run-pay        # Run pay server (port 10000)
make dev            # Dev mode with live reload (air)
make generate       # Regenerate SQLC code
make test           # Run tests
make lint           # Run linters
make onion          # Full check (sql + tests + build)
```

## Architecture

Onion is a modular Go backend using Clean Architecture with platform-based organization.

### Project Structure

```
internal/
├── modules/              # shared across all platforms
│   ├── config/           # site config, feature flags, themes
│   ├── translation/      # i18n (en, ar, fr)
│   └── user/             # authentication (email/password/JWT)
├── platforms/            # platform-specific modules
│   ├── rtylr/            # Rtylr platform
│   │   ├── echo/         # AI chat, SSE streaming
│   │   └── memory/       # conversation sessions, search, summarization
│   └── voxire/           # Voxire platform
│       └── careers/      # career listings
└── platform/             # shared infrastructure
    ├── auth/             # JWT, TokenValidator interface, device middleware
    ├── database/         # PostgreSQL connection, SQLC, schemas
    ├── server/http/      # Chi server, middleware, module registry
    ├── cache/            # Redis + in-memory (BigCache)
    ├── snips/            # AI client abstraction (OpenAI, Gemini)
    ├── mailer/           # email with DKIM signing
    ├── logger/           # structured logging
    ├── errors/           # error types
    └── utils/            # JWT, password hashing, converters
```

### Services

| Service | Port | Entry Point | Description |
|---------|------|-------------|-------------|
| API     | 9999 | `cmd/api/`  | Main API with all modules |
| Worker  | 9998 | `cmd/worker/` | Background worker |
| Pay     | 10000 | `cmd/pay/` | Payment service |

### Module Pattern

Every module follows Clean Architecture layers:

- `domain/` — entities with `FromRows()` to map from SQLC types
- `application/` — business logic, validation
- `infrastructure/` — repository (SQLC queries), external services
- `routes/handlers/` — HTTP handlers using local request/response structs
- `routes/routes.go` — route registration with Chi router
- `module.go` — entry point, implements `RegisterRoutes` + `GetAnonymousRoutes`

### Platform Pattern

Platforms group modules under `internal/platforms/<name>/`:
- Each platform has a `platform.go` with a `Platform` struct containing its modules
- The module registry auto-discovers and recurses into platform structs via reflection
- Platforms that have no `RegisterRoutes`/`GetAnonymousRoutes` are treated as containers

### Database Schemas

PostgreSQL schemas are used for platform isolation. Tables live in their respective schemas on RDS:

```
internal/platform/database/schemas/
├── shared/     # users, config, feature flags, themes, translations
├── voxire/     # vox_users, permissions, careers, clients, expenses, faq
└── rtylr/      # chat_sessions, conversations, user_memory
```

### Code Generation

- **SQLC**: Each module has `sql/sqlc.yaml` + `sql/queries/*.sql`. The build script auto-discovers all `sqlc.yaml` files under `internal/modules/` and `internal/platforms/`. Each sqlc.yaml references only the schema dirs it needs.

### Auth

- JWT-based with `TokenValidator` interface in `internal/platform/auth/`
- The `user` module implements `TokenValidator` — any platform can provide its own
- Device detection middleware extracts platform, language, user agent from requests
- Anonymous routes declared via `GetAnonymousRoutes()` on each module

## Documentation

See `docs/` for detailed guides:

- [Creating Endpoints](docs/creating-endpoints.md) — full flow from handler to route registration

## Key Technologies

- **Go 1.25** with Chi router
- **PostgreSQL** with SQLC for type-safe queries
- **JWT** (HMAC-SHA256) for authentication
- **SSE** for real-time streaming
- **OpenAI / Gemini** for AI features
- **Redis** + BigCache for caching

## Build Output

After `make build`, binaries are in `./bin/` (api, worker, pay).

## Dependencies

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.6
```

## Example Module

Use the `config` module (`internal/modules/config/`) as the reference implementation for the clean architecture pattern.
