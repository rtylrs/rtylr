# CLAUDE.md — onion Backend

## 1. PROJECT IDENTITY

- **Product:** RTYLR — commerce operating system
- **Repo role:** The single Go API server (`onion`) that powers every frontend app. One process, one port, all modules.
- **Go module:** `github.com/voxire/onion`
- **Production API base:** `https://api.voxire.dev/`
- **Server port:** `9999`
- **Built by:** Voxire — Abed El-Fattah Amouneh & Mohammad Homsi
- **Business email:** info@voxire.com
- **WhatsApp:** +961 3 940 708

---

## 2. GIT WORKFLOW — RULES

- Feature branches are the **default**. Prefix: `feat/`, `fix/`, `chore/`, `refactor/`, `style/`
- **Direct push to `main`** is acceptable ONLY for: single-file hotfixes or trivial config changes (e.g., a typo in an error message). Every multi-file or logic-bearing change goes through a branch.
- Commit author is **Abed El-Fattah Amouneh** only. No `Co-Authored-By` trailers — ever.
- **No `.md` files** in commits except `CLAUDE.md` itself.
- **Atomic commits** — one logical unit per commit. Don't bundle unrelated changes.
- Commit message format: short imperative phrase, no conventional prefix. `add team invite handler` not `feat: add team invite handler`.

### Definition of Done

A task is only complete when ALL of the following are true:

- `go build ./...` compiles with zero errors
- `go vet ./...` passes with zero warnings
- All errors are handled explicitly — no `_` discards
- The code follows every pattern documented in this file (correct layer placement, named structs, etc.)
- No `fmt.Println` or `log.Println` in the diff
- No commented-out code in the diff
- Every tenant-scoped query filters by `owner_id`
- If a new module was added, it is registered in `main.go`
- If a new architectural pattern was introduced, it is documented in `docs/` and in this CLAUDE.md

### PR Self-Review Checklist

Before opening a pull request, run through this list personally:

- [ ] `go build ./...` and `go vet ./...` both pass with zero output
- [ ] No `.env` changes or secrets in the diff
- [ ] No `fmt.Println`, `log.Println`, or `fmt.Printf` debug statements
- [ ] No commented-out code
- [ ] Every new query filters by `owner_id`
- [ ] Every error is explicitly handled — no `_` discards
- [ ] `go.sum` changes are intentional (new dep added, not corrupted)
- [ ] Branch name follows the `feat/`, `fix/`, `chore/` prefix convention
- [ ] If a new module was added, it is in `apiModules` and `initializeModules()` in `main.go`
- [ ] If a new endpoint was added or changed, `rtylr/docs/` has been updated

### When You Introduce a New Pattern

If you create a new convention (a new middleware, a new cross-cutting utility, a new way to handle a class of errors, etc.), document it in this CLAUDE.md and in `rtylr/docs/` (the frontend docs folder is the team's shared reference) before the PR is mergeable.

### Updating Documentation

**Always update `rtylr/docs/` when you make changes that affect the API contract, authentication flow, endpoint structure, or deployment.** The frontend team reads those docs — keep them accurate.

---

## 3. TECH STACK

| Component | Technology | Version | Notes |
|---|---|---|---|
| Language | Go | 1.26 | `go 1.26` in `go.mod` |
| HTTP router | go-chi/chi | v5.2.2 | Middleware-based, context-driven |
| Database | MySQL | 8.0 | Production: `108.167.180.218`, db `voxirete_rtylr_db` |
| DB driver | go-sql-driver/mysql | v1.9.3 | |
| Query builder | Masterminds/squirrel | v1.5.4 | Used in some modules |
| SQL codegen | sqlc (internal) | — | Generated code lives in `internal/modules/{mod}/internal/infrastructure/persistence/sqlc/` |
| Cache (in-process) | allegro/bigcache | v3.1.0 | In-memory LRU |
| Cache (distributed) | redis/go-redis | v9.17.0 | Optional; can be disabled |
| Auth | golang-jwt/jwt | v5.2.2 | HS256 JWT |
| Password hashing | golang.org/x/crypto | — | bcrypt |
| ID generation | google/uuid | v1.6.0 | |
| Email | resend/resend-go | v2.28.0 | |
| AI (OpenAI) | openai/openai-go | v2.0.2 | Echo module |
| AI (Gemini) | google/generative-ai-go | v0.20.1 | Echo module |
| XML | beevik/etree | v1.5.1 | |
| Env loading | joho/godotenv | v1.5.1 | Loads `.env` at startup |
| Concurrency cache | puzpuzpuz/xsync | v3.5.1 | Lock-free concurrent map |
| Proto contracts | github.com/voxire/voxie | submodule at `./voxie` | Generated Go types |
| Scheduler | codnect.io/chrono | v1.1.3 | Background task scheduling |

### Adding Dependencies

**Before adding any new Go module**, check whether an existing dependency already covers the need. Run `grep -r "your-need" go.mod` or check the packages already imported across `internal/`. Don't add a second library for a purpose already covered (e.g., don't add a new JWT library if `golang-jwt/jwt` already exists, don't add a new UUID library if `google/uuid` is present).

If a new dependency is genuinely needed, add it via `go get` and ensure `go.sum` is updated in the same commit.

---

## 4. ARCHITECTURE & PATTERNS

### Repository Layout

```
onion/
├── cmd/
│   └── api/
│       └── main.go           Entry point — initializes all modules, wires routes, starts server
├── internal/
│   ├── modules/              Domain modules — one directory per bounded context
│   │   └── <module>/
│   │       ├── module.go     Module struct, NewModule(), RegisterRoutes(), GetAnonymousRoutes()
│   │       ├── sql/          Raw SQL migration files for this module
│   │       └── internal/
│   │           ├── domain/         Pure business types, sentinel errors, repository interfaces
│   │           ├── application/    Service layer — business logic only, calls repository
│   │           ├── infrastructure/ SQL repository implementation
│   │           │   ├── repository.go
│   │           │   └── persistence/
│   │           │       └── sqlc/   sqlc-generated query code (auto-generated, do not edit)
│   │           └── routes/
│   │               ├── routes.go             Route registration (calls RegisterGET/POST/etc.)
│   │               └── handlers/
│   │                   └── *_handlers.go     One operation per file (ListX, CreateX, UpdateX, DeleteX)
│   └── platform/             Cross-cutting infrastructure (no business logic)
│       ├── auth/             JWT validation, AuthUser context, device context
│       ├── cache/            Redis + bigcache abstraction, cache invalidation
│       ├── config/           os.Getenv wrappers for env var access
│       ├── database/
│       │   ├── dbclient/     MySQL connection pool (singleton)
│       │   └── schemas/      All SQL schema files organised by domain
│       ├── email/            Email sending (Resend)
│       ├── errors/           ServerError type, error constructors
│       ├── logger/           Structured logger (wraps zap or similar)
│       ├── models/           Shared value types (none currently)
│       ├── server/
│       │   ├── http/         Chi wrappers: RegisterGET/POST/PUT/DELETE, middleware, auth, CORS
│       │   └── grpc/         gRPC server (secondary, not primary API)
│       ├── snips/            Code snippet utilities
│       ├── status/           /status health check endpoint
│       └── utils/            Cross-cutting utilities (jwtutils, passwordutils, etc.)
├── scripts/                  Shell scripts: migrate.sh, reset-db.sh, build.sh, lint.sh, test.sh
├── docs/                     backend.md, database.md — authoritative architecture references
├── voxie/                    Git submodule — proto contracts + generated Go types
├── app.yaml                  App Engine / deployment config (see note below)
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

> **Note on `app.yaml`:** This file contains deployment configuration for Google Cloud App Engine (or a compatible platform). It defines the runtime, scaling settings, and environment variable references used in production. Do not edit it without understanding the production deployment implications. It is not used for local development.

### The Onion Architecture — STRICT

Every module follows four layers with inward dependency only. **This is non-negotiable.**

```
HTTP Request
    ↓
Chi Router → wrapHandler()                    [platform/server/http]
    ↓
HandlerFunc[Req, Res].Serve(ctx, *Req)        [routes/handlers/]
    ↓
application.Service.Method(ctx, ...)          [application/]
    ↓
domain.Repository interface                   [domain/]  (interface only)
    ↓
infrastructure.Repository (SQL)               [infrastructure/]
    ↓
MySQL
```

- **Handlers**: parse HTTP request, call service, format response. No SQL. No business rules.
- **Service (application)**: business rules, validation, orchestration. No SQL. No HTTP details.
- **Repository (infrastructure)**: SQL queries only. No business logic.
- **Domain**: pure Go structs, sentinel errors (`var ErrNotFound = errors.New(...)`), repository interfaces. Zero external dependencies.

### Forbidden Patterns — Never Do These

These are the most common ways the architecture gets broken. Treat violations as bugs:

```go
// ❌ NEVER call dbclient.Pool() in a handler or service
func (h *MyHandler) Serve(ctx context.Context, req *MyRequest) (MyResponse, *errors.ServerError) {
    db := dbclient.Pool() // WRONG — SQL only in infrastructure layer
}

// ❌ NEVER write SQL strings in a handler or service
func (s *Service) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    row := db.QueryRow("SELECT * FROM users WHERE id = ?", id) // WRONG
}

// ❌ NEVER put business logic in a repository
func (r *Repository) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    user := r.queries.GetUser(ctx, id)
    if user.Role == "admin" { // WRONG — business logic belongs in service
        user.Permissions = allPermissions
    }
}

// ❌ NEVER reference http.Status* codes in the service layer
func (s *Service) InviteUser(ctx context.Context, email string) error {
    return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, ...} // WRONG — service should return domain errors
}
```

### Handler Pattern — EVERY handler follows this exactly

```go
// Routes function — lists all URL paths for this operation
func ListUsersRoutes() []string { return []string{"/team/users", "/team/members"} }

// Request struct — GET: query params auto-parsed; POST/PUT: JSON body auto-decoded
type ListUsersRequest struct {
    OwnerID int64 `json:"owner_id"`
}

// Response struct — always a named struct, even for single-field responses
type ListUsersResponse struct {
    Users []*hubv1.StaffUserPayload `json:"users"`
}

// Handler struct — holds service dependency
type ListUsersHandler struct{ Service *application.Service }

// Serve — implements HandlerFunc[Req, Res]
func (h *ListUsersHandler) Serve(ctx context.Context, _ *ListUsersRequest) (ListUsersResponse, *errors.ServerError) {
    device := authPkg.DeviceCtx(ctx)
    if device.User == nil {
        return ListUsersResponse{}, errorModels.PermissionDeniedServerError()
    }

    users, err := h.Service.ListStaffUsers(ctx, device.User.UserID)
    if err != nil {
        return ListUsersResponse{}, &errorModels.ServerError{
            HTTPCode: http.StatusInternalServerError,
            Code:     http.StatusInternalServerError,
            Message:  "Failed to list users",
        }
    }

    payloads := make([]*hubv1.StaffUserPayload, 0, len(users))
    for _, u := range users {
        payloads = append(payloads, toProto(u))
    }

    return ListUsersResponse{Users: payloads}, nil
}
```

### Route Registration Pattern

```go
// routes/routes.go
func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
    serverHttp.RegisterGET[handlers.ListUsersRequest, handlers.ListUsersResponse](
        ctx, router, handlers.ListUsersRoutes(), &handlers.ListUsersHandler{Service: service},
    )
    serverHttp.RegisterPOSTWithValidation[handlers.InviteUserRequest, handlers.InviteUserResponse](
        ctx, router, handlers.InviteUserRoutes(), &handlers.InviteUserHandler{Service: service},
    )
}
```

Registration helpers: `RegisterGET`, `RegisterGETWithValidation`, `RegisterPOST`, `RegisterPOSTWithValidation`, `RegisterPUT`, `RegisterPUTWithValidation`, `RegisterDELETE`, `RegisterDELETEWithValidation`.

**Use `WithValidation` variants for all mutating operations** (POST, PUT, DELETE). Implement `Validate(ctx, *Req) *errors.ServerError` on the handler to opt in.

### Module Entry Point Pattern

```go
type Module struct {
    service *application.Service
}

func NewModule(ctx context.Context) *Module {
    return &Module{service: application.NewService(ctx)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
    routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
    return []string{}
}
```

### Adding a New Module — Step-by-Step

Follow this checklist exactly when creating a new module:

1. Create `internal/modules/<module_name>/` directory.
2. Create `module.go` with `Module` struct, `NewModule()`, `RegisterRoutes()`, `GetAnonymousRoutes()`.
3. Create `internal/domain/` with: domain structs, sentinel errors (`var ErrXNotFound = errors.New(...)`), repository interface.
4. Create `internal/application/service.go` with: `Service` struct, `NewService()`, all business logic methods.
5. Create `internal/infrastructure/repository.go` with: `Repository` struct holding `*sqlc.Queries`, implementing the domain interface.
6. Create `internal/infrastructure/persistence/sqlc/queries/` with SQL query files.
7. Run `./scripts/generate_sqlc.sh` to generate the query code.
8. Create `internal/routes/routes.go` with the route registration function.
9. Create `internal/routes/handlers/` with one `*_handlers.go` file per operation.
10. Add the module's SQL schema to `internal/platform/database/schemas/`.
11. Run `./scripts/migrate.sh` to apply the schema.
12. **Add the module to `apiModules` struct in `cmd/api/main.go`** and instantiate it in `initializeModules()`.
13. Confirm `go build ./...` passes.

### Auth in Handlers: DeviceCtx vs GetAuthUser

**Always prefer `authPkg.DeviceCtx(ctx)`** for reading auth state in handlers. It bundles both user identity and platform/device context in one call.

```go
// ✅ Preferred
device := authPkg.DeviceCtx(ctx)
if device.User == nil {
    return Res{}, errorModels.PermissionDeniedServerError()
}
ownerID := device.User.UserID
businessID := device.User.BusinessID
platform := device.Platform   // e.g., "hub", "pos", "erp"
```

Use `authPkg.GetAuthUser(ctx)` only in lower-level platform utilities that don't have access to the device layer.

### Error Handling Pattern

- All errors are wrapped in `*errors.ServerError` before being returned from handlers.
- **Never discard an error with `_`**. This applies to helper functions too — `strconv.ParseFloat`, `strconv.ParseInt`, `time.Parse`, and similar conversions inside repository scan helpers must handle errors explicitly. The standard pattern for parse helpers that can't propagate an error is to log at Warn level and return the zero value:

```go
// ✅ correct — never discard parse errors silently
func parseDecimal(s string) float64 {
    if s == "" {
        return 0
    }
    f, err := strconv.ParseFloat(s, 64)
    if err != nil {
        logger.Global().Warn("failed to parse decimal value from database",
            zap.String("value", s),
            zap.Error(err),
        )
        return 0
    }
    return f
}

// ❌ wrong — silently returns 0 on corrupt data
func parseDecimal(s string) float64 {
    f, _ := strconv.ParseFloat(s, 64)
    return f
}
```

```go
if err != nil {
    if errors.Is(err, domain.ErrNotFound) {
        return Res{}, &errorModels.ServerError{
            HTTPCode: http.StatusNotFound,
            Code:     http.StatusNotFound,
            Message:  "Resource not found",
        }
    }
    return Res{}, &errorModels.ServerError{
        HTTPCode: http.StatusInternalServerError,
        Code:     http.StatusInternalServerError,
        Message:  "Internal server error",
    }
}
```

Use `errorModels.PermissionDeniedServerError()` for 403s. Use `errorModels.RedirectServerError(location)` for redirect responses.

### Logging Pattern

Use `logger.Ctx(ctx)` everywhere. Never use `fmt.Println` or the standard `log` package.

```go
logger.Ctx(ctx).Info("user invited",
    zap.Int64("owner_id", ownerID),
    zap.String("email", email),
)

logger.Ctx(ctx).Error("failed to list users",
    zap.Int64("owner_id", ownerID),
    zap.Error(err),
)
```

Always attach `owner_id`, resource ID, and operation context to error logs.

### Pagination Pattern

**All list endpoints must be paginated.** Never return an unbounded result set. The standard pagination envelope is:

```go
// Request — always include limit + offset
type ListOrdersRequest struct {
    Limit  int32 `json:"limit"`
    Offset int32 `json:"offset"`
}

// Response — always include total so the frontend knows how many pages exist
type ListOrdersResponse struct {
    Orders []*posv1.OrderPayload `json:"orders"`
    Total  int64                 `json:"total"`
    Limit  int32                 `json:"limit"`
    Offset int32                 `json:"offset"`
}
```

Default `Limit` to 20 in the service layer if the request sends 0. Cap `Limit` at 100 to prevent abuse. Pass `LIMIT` and `OFFSET` to the SQL query; also run a separate `COUNT(*)` query (or use sqlc's count variant) to populate `total`.

### Transaction Pattern

Use database transactions for any operation that writes to multiple tables and must be atomic. Transactions belong in the **infrastructure layer only** — never in the service layer.

```go
// infrastructure/repository.go
func (r *Repository) TransferStock(ctx context.Context, fromID, toID int64, qty int32) error {
    tx, err := dbclient.Pool().BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()

    qtx := r.queries.WithTx(tx)

    if err := qtx.DeductStock(ctx, sqlc.DeductStockParams{ID: fromID, Qty: qty}); err != nil {
        _ = tx.Rollback()
        return err
    }
    if err := qtx.AddStock(ctx, sqlc.AddStockParams{ID: toID, Qty: qty}); err != nil {
        _ = tx.Rollback()
        return err
    }

    return tx.Commit()
}
```

Rules:
- Always `defer` a rollback guard that fires on panic.
- Always check and return the error from every statement inside the transaction.
- Never let a transaction escape the repository — commit or rollback before returning.
- Transactions that span multiple repositories are a design smell — reconsider the domain model.

### Background Job Pattern

Background jobs use the `codnect.io/chrono` scheduler, initialised in `main.go`. Register new jobs in `initializeModules()` by calling the module's scheduler method after `NewModule()`.

```go
// module.go — expose a method that registers the job
func (m *Module) RegisterScheduledJobs(scheduler *chrono.TaskScheduler) {
    _, err := scheduler.ScheduleWithCron(func(ctx context.Context) {
        if err := m.service.ProcessPendingOrders(ctx); err != nil {
            logger.Ctx(ctx).Error("background job failed", zap.Error(err))
        }
    }, "0 * * * *") // every hour
    if err != nil {
        panic(fmt.Sprintf("failed to schedule job: %v", err))
    }
}
```

Rules:
- Every job must log errors — never swallow them silently.
- Jobs must respect context cancellation: pass `ctx` through to every DB/cache/service call.
- Jobs must be idempotent — running them twice must produce the same result as running once.
- Do not put business logic directly in the job function. Call a service method.

### Graceful Shutdown

The server handles `SIGINT`/`SIGTERM` with a shutdown deadline. Any goroutine you start must respect context cancellation — do not block indefinitely.

```go
// ✅ correct — respects context
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            // do work
        case <-ctx.Done():
            return
        }
    }
}()

// ❌ wrong — ignores cancellation, blocks shutdown
go func() {
    for {
        time.Sleep(5 * time.Minute)
        // do work
    }
}()
```

### Cache Usage Pattern

```go
// Cache key naming: domain:ownerID:resource
key := fmt.Sprintf("config:%d:settings", ownerID)
cached, err := cache.Get(ctx, key)
if err == nil {
    return cached, nil
}

config, err := r.queries.GetConfig(ctx, ownerID)
if err != nil {
    return nil, err
}

if setErr := cache.Set(ctx, key, config, 5*time.Minute); setErr != nil {
    logger.Ctx(ctx).Warn("failed to set cache", zap.String("key", key), zap.Error(setErr))
}

return config, nil
```

Cache writes are never fatal. **Invalidate cache** on every mutation:

```go
cache.Delete(ctx, fmt.Sprintf("config:%d:settings", ownerID))
```

### Database Pattern

- Connection: singleton MySQL pool via `dbclient.Pool()` — initialized in `main.go`.
- New queries: write SQL in `internal/infrastructure/persistence/sqlc/queries/*.sql` then run `./scripts/generate_sqlc.sh`.
- **Multi-tenancy**: every tenant-scoped query filters by `owner_id`.
- Primary keys: `BIGINT NOT NULL AUTO_INCREMENT` → `int64` in Go.
- Timestamps: `TIMESTAMP` columns → `time.Time` in Go.

**Cross-schema reads**: Reporting and payroll modules read directly from tables owned by other modules (e.g., `hr_payroll` reading from `hr_timesheets`). This is acceptable at the infrastructure layer only — never in the service layer. Every cross-schema query **must** be annotated with a comment on the line immediately above it:

```go
// cross-schema read from hr_time — approved timesheet hours for payroll computation
rows, err := db.QueryContext(ctx, "SELECT employee_id, SUM(duration_minutes) FROM hr_timesheets WHERE ...")

// cross-schema aggregation read from hr_core + hr_leave
rows, err := db.QueryContext(ctx, "SELECT e.department_id, COUNT(*) FROM hr_employees e JOIN hr_leave_requests ...")
```

Missing this comment is treated as a bug — it signals undocumented coupling between modules.

### JSON & API Contract

- All Go JSON struct tags use `snake_case`. **Never change to camelCase.**
- Response bodies always use named structs. Never return `map[string]interface{}` from a handler.
- Use proto-generated types from `voxie/gen/go/` as payload types when available.

### Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Module package name | lowercase | `package team` |
| Handler struct names | PascalCase + operation + Handler | `ListUsersHandler` |
| Routes functions | PascalCase + operation + Routes | `ListUsersRoutes()` |
| Request/Response types | PascalCase + operation + Request/Response | `InviteUserRequest` |
| Service methods | PascalCase verb + noun | `ListStaffUsers`, `InviteUser` |
| Repository methods | mirror service methods | `ListStaffUsers`, `GetUserByEmail` |
| Domain errors | Err + noun | `ErrUserNotFound` |
| Module directories | `snake_case` | `pos_orders/`, `time_off/` |
| Cache keys | `domain:ownerID:resource` | `config:123:settings` |

### Middleware Stack (applied globally)

1. `middleware.Logger` (Chi built-in)
2. `middleware.Recoverer` (Chi built-in)
3. `corsMiddleware()` — allows `*.rtylr.com`, `*.voxire.com`, `*.voxire.tech`, `localhost`
4. `CrashNotifierMiddleware()` — logs panics
5. `SecurityHeadersMiddleware()` — CSP, HSTS, X-Frame-Options, etc.
6. `SecureErrorMiddleware()` — sanitizes error responses
7. `RateLimitMiddleware(100, time.Minute)` — global rate limiter
8. `SecurityMiddleware()` — blocks suspicious request patterns
9. `RequestLoggerMiddleware()` — structured request logging
10. `GetAuthMiddlewareFunc()` — JWT validation
11. `authPkg.DeviceMiddleware()` — device/platform context

---

## 5. ROUTES / ENDPOINTS

All relative to `https://api.voxire.dev`. Auth required unless noted.

### Platform

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | None | Health check |
| GET | `/status` | None | System status (DB, cache) |
| GET | `/` | None | Liveness check |
| POST | `/invalidate-cache` | None | Cache invalidation |

### User & Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/register` | None | Register new account |
| POST | `/login` | None | Login, returns JWT |
| GET | `/account` | ✓ | Get current account |
| PUT | `/account/profile` | ✓ | Update profile |
| PUT | `/account/password` | ✓ | Change password |
| DELETE | `/account/delete` | ✓ | Delete account |

### Config

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/config/create` | ✓ | Create business config |
| GET | `/get-configs` | ✓ | Get config for current business |
| PUT | `/config/update` | ✓ | Update config |
| GET | `/config/faqs` | ✓ | Get FAQs |

### Team (Hub)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/team/users` | ✓ | List staff users |
| GET | `/team/members` | ✓ | List team members (alias) |
| GET | `/team/members/{id}` | ✓ | Get member by ID |
| POST | `/auth/invite` | ✓ | Invite a user |
| PUT | `/team/permissions` | ✓ | Update user permissions/role |
| DELETE | `/team/users` | ✓ | Deactivate user |
| GET | `/team/me` | ✓ | Current user profile |
| GET | `/auth/invite/verify` | None | Verify invite token |
| POST | `/auth/invite/accept` | None | Accept invite |
| GET | `/team/onboard` | ✓ | Mark user as onboarded |
| POST | `/team/request-access` | ✓ | Request app access |

### ERP

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/erp/products` | ✓ | List products |
| POST | `/erp/products` | ✓ | Create product |
| PUT | `/erp/products` | ✓ | Update product |
| DELETE | `/erp/products` | ✓ | Delete product |
| GET | `/erp/suppliers` | ✓ | List suppliers |
| POST | `/erp/suppliers` | ✓ | Create supplier |
| PUT | `/erp/suppliers` | ✓ | Update supplier |
| GET | `/erp/purchase-orders` | ✓ | List purchase orders |
| POST | `/erp/purchase-orders` | ✓ | Create purchase order |
| PUT | `/erp/purchase-orders/status` | ✓ | Update PO status |
| POST | `/erp/inventory/adjust` | ✓ | Adjust inventory |
| GET | `/erp/inventory/report` | ✓ | Inventory report |

> Full endpoint list for HR, CRM, Finance, Flow, POS, and all other modules follows the same pattern. See each module's `routes/handlers/` directory for the definitive list.

---

## 6. DATABASE

- **Engine:** MySQL (InnoDB) — production at `108.167.180.218`
- **Database name:** `voxirete_rtylr_db`
- **Multi-tenancy model:** `owner_id BIGINT NOT NULL` on every tenant-scoped table. Every query must filter by `owner_id`.
- **Primary keys:** `BIGINT NOT NULL AUTO_INCREMENT` throughout.
- **Schema files:** `internal/platform/database/schemas/` — apply in order: `shared → rtylr → voxire → erp → hr → crm → finance → flow → pos → hub`.

### Migration commands

```bash
export DB_HOST=108.167.180.218
export DB_PORT=3306
export DB_USER=voxirete_onion
export DB_PASSWORD=<password>
export DB_NAME=voxirete_rtylr_db

./scripts/migrate.sh       # apply all schemas (idempotent)
./scripts/reset-db.sh      # DROP + recreate (DESTRUCTIVE — dev only)
```

### sqlc codegen

```bash
./scripts/generate_sqlc.sh
```

---

## 7. CODE QUALITY RULES — NON-NEGOTIABLE

1. **Handle every error explicitly.** Never discard with `_`.
2. **No naked returns.** Always name what you're returning.
3. **No global mutable state** outside of explicitly managed singletons (dbclient, cache, logger).
4. **No SQL in handlers or service layer.** SQL belongs in `infrastructure/` only.
5. **No business logic in handlers.** Handlers parse inputs, call services, format outputs.
6. **No HTTP concepts in service layer.** Services must not know about status codes or request/response types.
7. **Every function must have an explicit return type.**
8. **No `fmt.Println` or `log.Println` in committed code.**
9. **All JSON struct tags are snake_case.** No exceptions.
10. **Validate all external inputs in service layer** — not just in handlers.
11. **All env vars must be read via `platform/config` package.** No `os.Getenv()` in business modules.
12. **New modules must be added to `apiModules` and `initializeModules()` in `main.go`.**

---

## 8. CONFIGURABILITY RULES

RTYLR is a multi-tenant SaaS product sold to restaurants, agencies, law firms, retail groups, and every other type of business. **This shapes every design decision.** A restaurant and a law firm have fundamentally different workflows, policies, and rules — but both subscribe to the same software. The only way to serve them both well is to make every business rule configurable by the company, never hardcoded by us.

### The developer checklist

Before hardcoding any behavior, ask both of these questions. If either answer is yes, it must be a configurable value stored in the company's config:

1. **"Would a restaurant and a law firm want this to behave differently?"** — If two different types of businesses would need different behavior, it is a configuration, not a product rule.
2. **"Is this a business rule or a product rule?"** — Product rules (pagination, auth flow, JSON structure) are fixed. Business rules (leave entitlements, approval chains, overtime thresholds, role names, notification channels) belong in config tables, not in code constants.

If you can answer "no" to both questions with confidence, hardcoding is acceptable — but add a comment explaining the reasoning so the next developer understands the decision was intentional.

### Things that must never be hardcoded — ever

These are the most common violations. Treat them as bugs identical in severity to a missing `owner_id` filter:

```go
// ❌ NEVER hardcode an approval chain
if user.Role == "manager" {
    // auto-approve
}
// ✅ Read the company's configured approval chain for this request type
chain, err := s.approvalChainRepo.GetChain(ctx, ownerID, domain.RequestTypeLeave)

// ❌ NEVER hardcode role names as permission gates
if user.Role == "hr_admin" {
    // allow action
}
// ✅ Check the permission key against the company's role permission map
if !s.hasPermission(ctx, user, domain.PermissionManageEmployees) {
    return domain.ErrPermissionDenied
}

// ❌ NEVER hardcode a business policy value
const annualLeaveDays = 14
// ✅ Read it from the company's leave type configuration
leaveType, err := s.leaveRepo.GetLeaveType(ctx, ownerID, leaveTypeID)
entitlement := leaveType.EntitledDays

// ❌ NEVER hardcode a notification channel
email.Send(ctx, user.Email, subject, body)
// ✅ Dispatch through the notification config layer which reads enabled channels
s.notificationService.Dispatch(ctx, ownerID, domain.NotificationEvent{...})
```

### Config storage pattern

Two layers, both required:

- **Module-level config tables**: domain-specific business rules live in the module that owns them. Leave policies in `hr_leave`. Overtime multipliers in `hr_overtime`. Approval chains in `hr_core`. Payroll methodology in `hr_payroll`. Each module reads its own config and is responsible for applying defaults if a company hasn't configured it yet.
- **Platform-level config** (`platform/config`): cross-cutting settings that affect infrastructure (env vars, feature flags, external API keys). Read via `platform/config` functions, never `os.Getenv()` directly in modules.

### Fallback behavior when config is missing

A company may not have completed their setup. The system must handle unconfigured state gracefully:

- **Non-critical features** (dashboard widgets, optional analytics, display preferences): fall back to a sensible default silently. Never crash or return an error because a config row doesn't exist yet.
- **Critical features** (payroll run, timesheet approval, leave balance deduction): if the required config is absent, return a clear domain error that the frontend can translate into a "Configure X first" prompt. Do not allow the operation to proceed with silent defaults when money or compliance is involved.

```go
// Service layer — distinguish critical vs non-critical missing config
func (s *Service) GetPayrollConfig(ctx context.Context, ownerID int64) (*domain.PayrollConfig, error) {
    cfg, err := s.repo.GetPayrollConfig(ctx, ownerID)
    if errors.Is(err, domain.ErrNotFound) {
        // Critical — payroll cannot run without config
        return nil, domain.ErrPayrollNotConfigured
    }
    if err != nil {
        return nil, err
    }
    return cfg, nil
}

func (s *Service) GetDashboardConfig(ctx context.Context, ownerID int64) (*domain.DashboardConfig, error) {
    cfg, err := s.repo.GetDashboardConfig(ctx, ownerID)
    if errors.Is(err, domain.ErrNotFound) {
        // Non-critical — return a safe default, log at debug level
        logger.Ctx(ctx).Debug("no dashboard config, using defaults", zap.Int64("owner_id", ownerID))
        return domain.DefaultDashboardConfig(), nil
    }
    if err != nil {
        return nil, err
    }
    return cfg, nil
}
```

### Scope and the boy scout rule

- **All new code** must follow these rules without exception.
- **Existing code**: when you open a file to make any change, fix any hardcoded business rule you encounter in that file before committing. Leave the code more configurable than you found it. You are not required to audit the entire codebase — only the files you touch.
- This is a **strong guideline**, not a PR blocker for every instance. If hardcoding is genuinely the right call (e.g. a product-level constant that will never differ between customers), document the reason in a comment. Do not use this exception for business rules.

---

## 9. SECURITY RULES

- **No secrets in code.** All credentials (DB password, JWT secret, API keys) live in `.env` and are loaded via `platform/config`. Never hardcode.
- **SQL injection prevention.** Always use parameterised queries via sqlc or squirrel. Never build SQL strings with `fmt.Sprintf` and user-supplied input.
- **All tenant data must be owner-scoped.** Every query against a tenant table must include `WHERE owner_id = ?`. Missing this filter is a data-leak vulnerability.
- **JWT validation is non-negotiable.** Never bypass `GetAuthMiddlewareFunc()` on a route that returns tenant data. Only declare a route anonymous via `GetAnonymousRoutes()` if it genuinely needs no auth.
- **Rate limiting is global.** The `RateLimitMiddleware` applies to every route. If a specific endpoint needs tighter limits (e.g., login, password reset), add a secondary per-route limiter.
- **Sensitive data must not appear in logs.** Never log passwords, tokens, or full PII (email is acceptable in structured fields for tracing, but never log the password even hashed).
- **Dependency audits:** Run `go list -m -u all` periodically and address high-severity CVEs before major releases.

---

## 9. PERFORMANCE RULES

- **All list endpoints must be paginated.** See the Pagination Pattern in section 4. Never return unbounded result sets.
- **N+1 query prevention.** If you need related data for a list, use a JOIN or a batch query — never query inside a loop. Review the SQL in every new list endpoint before merging.
- **New queries touching large tables must have an index.** Before adding a query that filters or sorts by a column, confirm there is a MySQL index on that column. Add the index in the schema migration if it is missing.
- **Cache expensive reads.** Any endpoint that aggregates data across multiple tables or performs heavy computation should cache its result. Default TTL: 5 minutes for user-facing config, 1 minute for transactional data.
- **Avoid SELECT *.** Always name the columns you need in queries. This prevents pulling large columns (e.g., image blobs) when you only need IDs and names.

---

## 10. KNOWN ISSUES

| Issue | File / Location | Priority |
|---|---|---|
| `go.mod` declares `go 1.26` but Go 1.26 is not yet released — likely `go 1.25` intended | `go.mod` | Low |
| Some modules may use `squirrel` directly instead of sqlc — inconsistent query approach | Various `infrastructure/` files | Medium |
| No `.env.example` in repo — environment setup relies on `docs/development.md` | Root | Medium |
| `echo` module depends on `memory` module directly (passed via `NewModule`) — only cross-module dependency | `cmd/api/main.go` | Low |
| Production DB host hardcoded in `docs/database.md` — should be env var only | `docs/database.md` | Low |
| gRPC server exists (`platform/server/grpc`) but no modules currently register gRPC services | `internal/platform/server/grpc/` | Info |
| No automated testing infrastructure visible (no `_test.go` files found beyond `platform/`) | Various modules | High |

---

## 11. SESSION STARTUP CHECKLIST

1. `git pull origin main` (or the relevant branch)
2. `git submodule update --init --recursive`
3. `git checkout -b feat/<task-name>`
4. Re-read `docs/backend.md` and `docs/database.md` for any updates
5. Re-read this CLAUDE.md
6. Confirm server starts: `go run ./cmd/api/`

---

## 12. INCIDENT RESPONSE

### API is returning 5xx / service is down

1. **Check the `/status` endpoint first:** `curl https://api.voxire.dev/status` — reports DB and cache health without needing log access.
2. **Check Render dashboard** (current backend host): log into [render.com](https://render.com), navigate to the `onion` service, and check the Logs tab for panics or startup errors.
3. If the DB is unreachable: verify MySQL at `108.167.180.218` is accessible. Check that DB credentials are set correctly in the Render service environment variables.
4. If it is a code panic: the `CrashNotifierMiddleware` logs the stack trace. Search Render logs for `panic` or `recovered`.

### Rolling back a deployment

In the Render dashboard: navigate to the `onion` service → Deploys tab → select the last known-good deploy → click **Rollback to this deploy**.

Alternatively, force-push or revert the commit on `main` and trigger a new deploy.

### Rolling back a database migration

Migrations use `IF NOT EXISTS` and are additive-only — there is no automatic rollback. To undo a schema change:

1. Write a new migration that reverses the change (DROP COLUMN, DROP TABLE, etc.).
2. Apply it manually via `./scripts/migrate.sh` against the production DB.
3. **Never run `./scripts/reset-db.sh` against production** — it is destructive.

### Monitoring

- **Current state:** No monitoring tooling is configured. Axiom (structured logging) and PostHog (product analytics) are planned.
- **Until monitoring is live:** rely on the Render logs dashboard and `curl https://api.voxire.dev/status` for health checks.
- **Escalation:** Contact Abed El-Fattah Amouneh — info@voxire.com / +961 3 940 708.

---

## 13. ENVIRONMENT SETUP

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.25+ | https://go.dev/dl/ |
| MySQL | 8.0+ | Local or remote |
| Redis | 7+ | Optional — set `REDIS_ENABLED=false` to skip |
| sqlc | v1.29.0 | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0` |
| air | latest | `go install github.com/air-verse/air@latest` (live reload) |

### First-time local setup (one-time only)

The `voxie` git submodule uses a relative URL (`../voxie`) that requires GitHub credentials to clone automatically. Since you already have the `voxie` repo locally, copy it instead:

```bash
cd ~/Development/onion
rm -rf voxie
cp -rp ~/Development/voxie voxie
```

The sqlc-generated code is gitignored and must be generated locally. Install sqlc and generate:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
export PATH="$PATH:$(go env GOPATH)/bin"
./scripts/generate_sqlc.sh   # generates all 98 modules (~10 seconds)
```

> **Note:** If `generate_sqlc.sh` fails on a module that includes a `migration_vN.sql` file in its schema directory, change that schema entry in the module's `sqlc.yaml` to reference `schema.sql` directly (e.g. `schemas/flow/schema.sql` instead of `schemas/flow/`). The `migration_*.sql` files use MySQL procedural syntax that sqlc cannot parse.

### Start the server

```bash
cd onion/
go mod download
cp .env.example .env  # or create manually
make dev              # runs air with live reload on port 9999
```

### Environment Variables

Create `onion/.env`:

```env
APP_ENV=development

# MySQL (required)
DB_HOST=localhost
DB_PORT=3306
DB_NAME=voxirete_rtylr_db
DB_USER=your_db_user
DB_PASSWORD=your_db_password

# JWT (required)
JWT_SECRET=your_jwt_secret_min_32_chars

# Redis (optional)
REDIS_ENABLED=false
REDIS_HOSTS=localhost:6379
REDIS_USERNAME=
REDIS_PASSWORD=

# In-memory cache
INMEMORY_CACHE_MAX_SIZE_MB=64
CACHED_MARSHALLING_ENABLED=true

# AI integrations (optional — echo module)
OPENAI_API_KEY=
OPENAI_ORG_ID=
OPENAI_BASE_URL=
GEMINI_API_KEY=

# Email (optional — invite emails)
RESEND_API_KEY=

# Cloudflare R2 (optional — recruit resume uploads)
# All four must be set together, or uploads are disabled and the
# frontend falls back to the "paste a link" input.
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
```

### Useful make targets

```bash
make build     # build binary
make test      # run tests
make lint      # run linters
```
