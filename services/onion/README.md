<img src="https://cdn.voxire.dev/adminupload/onion.png" alt="Onion" width="200">

# Onion 🧅

Onion is a full-featured backend framework that provides everything you need to build robust applications:

- **Modular Architecture**: Plug-and-play modules for any feature you need
- **Production Ready**: Built-in logging, monitoring, error handling, and recovery
- **Database First**: PostgreSQL with type-safe SQL queries and migrations
- **Authentication & Security**: JWT-based auth, middleware, and security features
- **Real-time Capabilities**: WebSocket support, SSE, and streaming
- **Extensible**: Easy to add new modules and functionality
- **Slack Integration**: Real-time monitoring and notifications
- **Auto-recovery**: Panic recovery with automatic server restart
- **Multi-Platform Support**: HTTP REST APIs for web + gRPC for mobile apps

## 📋 Prerequisites

- **Go 1.25.0+**
- **Docker & Docker Compose**
- **PostgreSQL 15+**

## 🛠️ Quick Start

### 1. Clone & Setup

```bash
git clone git@github.com:voxire/onion.git
cd onion
cp .env.example .env
# Edit .env with your configuration

# Initialize the proto submodule (grimoire)
git submodule update --init --recursive
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
# Start database and setup schema
make setup

# Run the application
make run
```

## 🗄️ Database Management

```bash
# Start database
make setup

# Stop database
docker-compose down

# Reset database
./scripts/reset-db.sh

# View database logs
docker-compose logs postgres

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

### Protobuf & gRPC

```bash
# Initialize or update the proto submodule
git submodule update --init --recursive
# Optional: pull latest from the remote for the submodule
git submodule update --remote proto

# Regenerate protobuf code
make proto
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
