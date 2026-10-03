# Onion 🧅

Onion is a full-featured backend framework that provides everything you need to build robust applications:

- **Modular Architecture**: Plug-and-play modules for any feature you need
- **Production Ready**: Built-in logging, monitoring, error handling, and recovery
- **Database First**: MySQL with type-safe SQL queries and migrations
- **Authentication & Security**: JWT-based auth, middleware, and security features
- **Real-time Capabilities**: WebSocket support, SSE, and streaming
- **Extensible**: Easy to add new modules and functionality
- **Auto-recovery**: Panic recovery with automatic server restart
- **Multi-Platform Support**: HTTP REST APIs for web and internal services

## 📋 Prerequisites

- **Go 1.26+**
- **MySQL-compatible database**. Local dev uses Percona Server 5.7.23 to match the current HostGator database.
- **Docker** for the optional local development environment.
- `mysql` and `mysqladmin` available locally if you run database scripts from your machine.

## 🛠️ Quick Start

### 1. Clone & Setup

```bash
git clone git@github.com:voxire/onion.git
cd onion
# Create .env and point DATABASE_URL or DB_* to a running MySQL instance
```

### 2. Install Dependencies

```bash
# Install sqlc for SQL code generation
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0

# Install golangci-lint for code quality
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.6
```

### 3. Start Development Environment

```bash
# Start local MySQL, Redis, and Floci
make local-env-up

# Copy .env.example to .env if needed. The local Docker environment
# matches the DB_* defaults in .env.example.
# Set FLOCI_ENABLED=true to route object storage through Floci.

# Validate database connectivity and run local setup checks
make setup

# Run the application
make run
```

### AI providers

Gemini is the default provider. Set `GEMINI_API_KEY` and optionally
`OPENCODE_API_KEY` for [OpenCode Go](https://opencode.ai/docs/go/) fallback.
The default order is Gemini → OpenCode Go → DeepSeek → OpenAI; providers
without keys are skipped. `AI_PRIMARY_PROVIDER` can select a configured provider.

OpenCode Go uses `deepseek-v4.1-flash` at `/zen/go/v1`, identifies Caret with
its own user agent and sends a stable opaque `x-opencode-session` per chat.
Calls without a conversation receive an isolated routing ID shared across
retries and failover. DeepSeek thinking is disabled to reserve the output
budget for the visible answer. Ordinary and JSON completions, streaming,
usage accounting, cooldowns and request timeouts use the shared provider layer.
Streaming switches providers only before visible response text is emitted.

Go is designed primarily for coding-agent traffic; review the provider's
current usage terms and limits before expanding its use. Hosted provider keys
come from environment configuration, not local OpenCode CLI credentials.

## 🗄️ Database Management

```bash
# Validate the configured MySQL connection
make setup

# Start/stop local MySQL, Redis, and Floci
make local-env-up
make local-env-down

# Reset database
./scripts/reset-db.sh

# Generate SQL code for all modules
make generate

# Check SQL identifier casing
make sql-check

# Optionally load tables from schema directories (dev tooling)
make load-tables

# Seed synthetic POS/ERP volume and inspect EXPLAIN plans before launch
make db-prelaunch-check
```

## 🔧 Development

### Building

```bash
# Build the project
make build

# The binary will be at ./bin/api
```

### Code Quality

```bash
# Run linters
make lint

# Run tests
make test

# Check SQL identifier casing
make sql-check
```

### Project Structure

```
onion/
├── cmd/api/             # API entry point
├── internal/            # Private application code
│   ├── modules/        # Feature modules
│   └── platform/       # Shared platform code, server, database, auth, cache
├── scripts/            # Build and setup scripts
├── docs/               # Project documentation
└── voxie/              # Local proto dependency and generated code
```

## 🚀 Deployment

### GitHub Actions

Pushes to `main` or `master` trigger the repository's Docker image build workflow.

### Manual Deployment

```bash
# Build for production
make build

# Run the binary
./bin/api
```
