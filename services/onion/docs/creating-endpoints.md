# Creating Endpoints in Onion

This guide documents the full flow to create an endpoint, from local request and response structs to wiring it into the server.

## 1. Create the Module Structure

**Shared module** — `internal/modules/<name>/`:

```
internal/modules/<name>/
├── module.go
├── sql/
│   ├── sqlc.yaml
│   └── queries/<name>.sql
└── internal/
    ├── domain/
    │   └── <entity>.go
    ├── application/
    │   └── service.go
    ├── infrastructure/
    │   ├── repository.go
    │   └── persistence/sqlc/
    └── routes/
        ├── routes.go
        └── handlers/
            └── <handler>.go
```

For a platform-specific module, use `internal/platforms/<platform>/<name>/` with the same structure.

## 2. SQL Schema + Queries

### Schema

Add the table definition in `internal/platform/database/schemas/<db>/`:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Queries

Add SQLC queries in `<module>/sql/queries/<name>.sql`:

```sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
```

### SQLC Config

Create `<module>/sql/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries"
    schema:
      - "../../../../../internal/platform/database/schemas/<db>"
    gen:
      go:
        package: "sqlc"
        out: "../internal/infrastructure/persistence/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
```

Run `make generate` or let `make build` handle it.

## 3. Domain

`internal/<path>/internal/domain/<entity>.go`

Plain struct with a `FromRows` method that maps from SQLC-generated types:

```go
type User struct {
    ID    int32
    Email string
}

func (u *User) FromRows(row sqlc.User) {
    u.ID = row.ID
    u.Email = row.Email
}
```

## 4. Infrastructure

`internal/<path>/internal/infrastructure/repository.go`

Database access layer using SQLC queries and the generic read and write executors:

```go
type Repository struct {
    queries *sqlc.Queries
    pool    *pgxpool.Pool
}

func NewRepository(_ context.Context) *Repository {
    pool := dbclient.Pool()

    return &Repository{
        queries: sqlc.New(pool),
        pool:    pool,
    }
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
    user, err := db.NewReadExecutor[domain.User, sqlc.User](db.MainDatabase).
        Fetch(ctx, r.queries.GetUserByEmail, email)
    if err != nil {
        return nil, err
    }

    return &user, nil
}
```

## 5. Application

`internal/<path>/internal/application/service.go`

Business logic layer. Calls the repository, handles validation, and orchestrates operations:

```go
type Service struct {
    repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
    return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (int64, error) {
    user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(email))
    if err != nil {
        return 0, ErrInvalidCredentials
    }

    if !passwordutils.CheckPasswordHash(password, user.PasswordHash) {
        return 0, ErrInvalidCredentials
    }

    return int64(user.ID), nil
}
```

## 6. Handler

`internal/<path>/internal/routes/handlers/<handler>.go`

HTTP handlers use local request and response structs with JSON tags:

```go
func LoginRoutes() []string {
    return []string{"/login", "/auth/login"}
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token   string `json:"token"`
    Message string `json:"message"`
    UserID  int64  `json:"userId"`
    Email   string `json:"email"`
    Name    string `json:"name"`
}

type LoginHandler struct {
    UserService *application.Service
}

func (h *LoginHandler) Serve(ctx context.Context, req *LoginRequest) (LoginResponse, *errorModels.ServerError) {
    userID, err := h.UserService.Authenticate(ctx, req.Email, req.Password)
    if err != nil {
        return LoginResponse{}, &errorModels.ServerError{
            Code:    http.StatusUnauthorized,
            Message: "Invalid email or password",
        }
    }

    token, _ := h.UserService.GenerateToken(ctx, userID)

    return LoginResponse{
        Token:   token,
        Message: "Login successful",
        UserID:  userID,
    }, nil
}
```

Optional validation can stay beside the handler:

```go
func (h *RegisterHandler) Validate(_ context.Context, req *RegisterRequest) *errorModels.ServerError {
    if req.Email == "" || req.Password == "" {
        return &errorModels.ServerError{
            Code:    http.StatusBadRequest,
            Message: "Email and password are required",
        }
    }

    return nil
}
```

Use the `WithValidation` registration helpers when a handler exposes `Validate`.

## 7. Route Registration

`internal/<path>/internal/routes/routes.go`

Register handlers with the Chi router using the generic HTTP helpers and the local request and response types:

```go
func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
    http.RegisterPOSTWithValidation[handlers.LoginRequest, handlers.LoginResponse](
        ctx,
        router,
        handlers.LoginRoutes(),
        &handlers.LoginHandler{UserService: service},
    )
}
```

For `GET` endpoints, use `RegisterGET`. For handlers without validation, use the non-`WithValidation` helper.

## 8. Module Entry Point

`internal/<path>/module.go`

Expose the module through a small struct that initializes the service and registers routes:

```go
type Module struct {
    service *application.Service
}

func New(ctx context.Context) *Module {
    return &Module{service: application.NewService(ctx)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux) {
    routes.RegisterRoutes(ctx, router, m.service)
}
```

## 9. Verification

```bash
make generate
make build
```

That covers the full endpoint flow in Onion.
