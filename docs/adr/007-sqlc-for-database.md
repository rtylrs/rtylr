# ADR-007: sqlc for Database Access (not GORM or raw SQL)

**Status:** Accepted  
**Date:** 2026-02-01  
**Decision:** All database queries use sqlc. No raw SQL in handler or service code. No ORM.

## Context

The Go backend needs to interact with PostgreSQL. Options:

1. **GORM** - ORM with auto-migration, associations, hooks
2. **sqlx** - Lightweight SQL mapper with struct scanning
3. **sqlc** - Generates type-safe Go code from SQL queries
4. **Raw database/sql** - Standard library, manual scanning

## Decision

sqlc. Write SQL queries in `.sql` files, run `sqlc generate`, get type-safe Go functions.

```sql
-- queries/tasks.sql
-- name: ListTasksByProject :many
SELECT id, title, status, assignee_id FROM tasks WHERE project_id = $1;
```

Generates:

```go
func (q *Queries) ListTasksByProject(ctx context.Context, projectID int64) ([]Task, error)
```

## Rationale

- **Type safety**: sqlc validates queries against the database schema at generation time. A typo in a column name is caught before compilation, not at runtime.
- **No magic**: Unlike GORM, there's no implicit behavior (N+1 queries, lazy loading, auto-migration). What you write in SQL is what executes.
- **Performance**: Generated code uses `database/sql` directly. No reflection, no query building overhead. Equivalent to hand-written SQL with zero runtime cost.
- **SQL is the API**: Complex queries (joins, CTEs, window functions) are written in SQL, not in a Go DSL. SQL is the right language for relational queries.
- **Architecture enforcement**: By requiring all SQL in `.sql` files under `internal/modules/{mod}/internal/repository/queries/`, we prevent raw SQL from leaking into handlers or services.

## Trade-offs

- **Generation step**: Must run `sqlc generate` after changing queries. The Dockerfile includes this step.
- **Migrations separate**: sqlc doesn't handle migrations. Schema changes require a separate migration tool.
- **Learning curve**: Developers must write SQL (many Go devs are ORM-dependent).

## Consequences

- Each module has `internal/repository/queries/*.sql` for query definitions
- Each module has `internal/repository/sqlc/` for generated Go code
- `scripts/generate_sqlc.sh` runs sqlc for all modules
- Handlers call Service -> Repository -> sqlc-generated Queries. Never direct SQL.
- The Dockerfile runs sqlc generation during the build step
