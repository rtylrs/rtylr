#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <module_name>"
  exit 1
fi

module="$1"
dir="internal/modules/${module}"

if [ -d "$dir" ]; then
  echo "Module already exists: $dir" >&2
  exit 1
fi

mkdir -p \
  "$dir/internal/application" \
  "$dir/internal/domain" \
  "$dir/internal/infrastructure" \
  "$dir/internal/routes" \
  "$dir/sql/queries" \
  "$dir/sql/schema"

cat > "$dir/module.go" <<EOF
package ${module}

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/${module}/internal/application"
	"github.com/voxire/onion/internal/modules/${module}/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}
EOF

cat > "$dir/internal/application/service.go" <<EOF
package application

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/${module}/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}
EOF

cat > "$dir/internal/infrastructure/repository.go" <<EOF
package infrastructure

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
EOF

cat > "$dir/internal/routes/routes.go" <<EOF
package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/${module}/internal/application"
)

func RegisterRoutes(_ context.Context, _ *chi.Mux, _ *application.Service) {}
EOF

cat > "$dir/internal/domain/${module}.go" <<EOF
package domain
EOF

cat > "$dir/sql/queries/${module}.sql" <<EOF
-- name: Ping :one
SELECT 1;
EOF

cat > "$dir/sql/sqlc.yaml" <<EOF
version: "2"
sql:
  - engine: "mysql"
    queries: "./queries"
    schema: "./schema"
    gen:
      go:
        package: "sqlc"
        out: "../internal/infrastructure/persistence/sqlc"
EOF

cat > "$dir/sql/schema/schema.sql" <<EOF
-- Add this bounded context's owned schema here.
EOF

echo "Generated $dir"
