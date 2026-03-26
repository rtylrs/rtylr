<img src="https://cdn.voxire.dev/adminupload/onion.png" alt="Onion" width="200">

# Onion 🧅

Onion is a full-featured backend framework that provides everything you need to build robust applications:

- **Modular Architecture**: Plug-and-play modules for any feature you need
- **Production Ready**: Built-in logging, monitoring, error handling, and recovery
- **Database First**: PostgreSQL with type-safe SQL queries and migrations
- **Authentication & Security**: JWT-based auth, middleware, and security features
- **Real-time Capabilities**: WebSocket support, SSE, and streaming
- **Extensible**: Easy to add new modules and functionality
- **Auto-recovery**: Panic recovery with automatic server restart
- **Multi-Platform Support**: HTTP REST APIs for web and internal services

## 📋 Prerequisites

- **Go 1.25.0+**
- **PostgreSQL 15+**
- **`psql` and `pg_isready` available locally**

## 🛠️ Quick Start

### 1. Clone & Setup

```bash
git clone git@github.com:voxire/onion.git
cd onion
cp .env.example .env
# Edit .env so DATABASE_URL or DB_* points to a running PostgreSQL instance
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
# Validate database connectivity and run local setup checks
make setup

# Run the application
make run
```

## 🗄️ Database Management

```bash
# Validate the configured database connection
make setup

# Reset database
./scripts/reset-db.sh

# Generate SQL code for all modules
make generate

# Check SQL identifier casing
make sql-check

# Optionally load tables from schema directories (dev tooling)
make load-tables
```

## 🔧 Development

### Building

```bash
# Build the project
make build

# The binary will be at ./bin/onionexec
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
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── app/            # Application setup and configuration
│   ├── modules/        # Feature modules (AI, Memory, User, Config)
│   └── platform/       # Shared platform code
├── db/                 # Database schemas and migrations
├── scripts/            # Build and setup scripts
└── test/               # Test files and examples
```

## 🚀 Deployment

### Render (Automatic)

Every push to `main` triggers an automatic build and deployment on Render.

### Manual Deployment

```bash
# Build for production
make build

# Run the binary
./bin/onionexec
```
