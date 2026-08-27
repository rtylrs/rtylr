# Backend (onion)

The backend is a single Go service called **onion**, located at `rtylr/onion/`. It serves all business modules through one HTTP process using the Chi router, listening on port `9999`.

- **Go module:** `github.com/voxire/onion`
- **Go version:** 1.25
- **API base:** `https://api.voxire.dev/`
- **Router:** `github.com/go-chi/chi/v5`
- **Database:** MySQL via `github.com/go-sql-driver/mysql`
- **Auth:** JWT via `github.com/golang-jwt/jwt/v5`
- **Cache:** Redis + bigcache

---

## Onion Architecture

Each business module follows a strict four-layer structure:

```
internal/modules/<module>/
├── module.go              # Module struct, NewModule(), RegisterRoutes(), GetAnonymousRoutes()
└── internal/
    ├── domain/            # Pure business types and repository interfaces
    ├── application/       # Service layer — business logic, calls repository
    ├── infrastructure/    # Repository implementation (SQL queries)
    └── routes/
        └── handlers/      # One file per operation, implements HandlerFunc[Req, Res]
```

**Data flow:**
```
HTTP Request
    ↓
Chi Router → wrapHandler()
    ↓
HandlerFunc[Req, Res].Serve(ctx, *Req) → (Res, *errors.ServerError)
    ↓
application.Service → domain.Repository interface
    ↓
infrastructure/ (SQL via squirrel)
    ↓
MySQL
```

### Intentional SQL-construction exception: AI query compiler

`internal/modules/aiquery/internal/domain/spec.go` is an intentional exception
to the rule that SQL belongs in infrastructure. It does not execute SQL or
accept caller-provided SQL. Instead, it compiles a structured `QuerySpec` into
a read-only, parameterized statement for the AI query feature. Every view,
column, aggregate, operator, grouping key, and ordering key is selected from an
explicit allowlist; `owner_id` is always injected as a bound tenant filter; and
the result limit is capped at `MaxRows`. The infrastructure executor remains
responsible for database access and runs the compiled query on the read-only
pool.

Keep this exception narrow. Changes to the AI query surface must preserve the
allowlist and tenant-binding tests in `internal/modules/aiquery/internal/domain`.
Do not generalize this pattern to ordinary module repositories; new SQL-backed
features still belong in `internal/infrastructure`.

---

## Handler Pattern

Every operation is its own struct implementing `HandlerFunc[Req, Res]`:

```go
type HandlerFunc[Req any, Res any] interface {
    Serve(ctx context.Context, req *Req) (Res, *errors.ServerError)
}

// Optional — implement to get automatic request validation before Serve()
type ValidatorFunc[Req any] interface {
    Validate(ctx context.Context, req *Req) *errors.ServerError
}
```

Route registration helpers in `internal/platform/server/http`:

```go
// GET — query params parsed automatically into Req struct
http.RegisterGET[Req, Res](ctx, router, []string{"/erp/products"}, &handler)

// POST/PUT with validation (calls Validate before Serve)
http.RegisterPOSTWithValidation[Req, Res](ctx, router, []string{"/erp/products"}, &handler)
http.RegisterPUTWithValidation[Req, Res](ctx, router, []string{"/erp/products"}, &handler)
http.RegisterDELETEWithValidation[Req, Res](ctx, router, []string{"/erp/products"}, &handler)
```

For GET and DELETE, query parameters are automatically parsed into the `Req` struct. Numbers become `int64`/`float64`, booleans are parsed from `"true"`/`"false"` strings.

For POST/PUT, the request body is JSON-decoded into the `Req` struct.

### Example handler

```go
type ListProductsRequest struct {
    OwnerID int64 `json:"ownerId"`
}

type ListProductsResponse struct {
    Products []domain.Product `json:"products"`
}

type ListProductsHandler struct {
    ERPService *application.Service
}

func ListProductsRoutes() []string { return []string{"/erp/products"} }

func (h *ListProductsHandler) Serve(ctx context.Context, req *ListProductsRequest) (ListProductsResponse, *errors.ServerError) {
    products, err := h.ERPService.ListProducts(ctx, req.OwnerID)
    if err != nil {
        return ListProductsResponse{}, errors.Internal(err)
    }
    return ListProductsResponse{Products: products}, nil
}
```

---

## Auth in Handlers

The auth middleware validates the `Authorization: Bearer <token>` header on all non-anonymous routes and injects an `AuthUser` into the request context.

```go
// Read the authenticated user from context
authUser, ok := authPkg.GetAuthUser(ctx)
if !ok {
    return Res{}, errors.Unauthorized()
}
userID := authUser.UserID
```

Anonymous routes are registered by returning their paths from `GetAnonymousRoutes()` on the module:

```go
func (m *Module) GetAnonymousRoutes() []string {
    return []string{"/login", "/register"}
}
```

---

## Module Registration

In `cmd/api/main.go`, every module is initialized and registered:

```go
func initializeModules(ctx context.Context) *apiModules {
    return &apiModules{
        User:    user.NewModule(ctx),
        Config:  config.NewModule(ctx),
        ERP:     erp.NewModule(ctx),
        HR:      hr.NewModule(ctx),
        CRM:     crm.NewModule(ctx),
        Finance: finance.NewModule(ctx),
        Flow:    flow.NewModule(ctx),
        POS:     pos.NewModule(ctx),
        Hub:     hub.NewModule(ctx),
        // ...
    }
}
```

`http.AutoRegisterRoutes()` calls `RegisterRoutes()` on each module automatically.

---

## Available Endpoints

All endpoints are relative to `https://api.voxire.dev`.

### user module

| Method | Path | Description |
|--------|------|-------------|
| POST | `/register` | Register new user (anonymous) |
| POST | `/login` | Login, returns JWT (anonymous) |
| GET | `/account` | Get current account |
| PUT | `/account/profile` | Update profile |
| PUT | `/account/password` | Change password |
| DELETE | `/account/delete` | Delete account |

### config module

| Method | Path | Description |
|--------|------|-------------|
| POST | `/config/create` | Create business config |
| GET | `/get-configs` | Get config by owner |

### hub module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/hub/users` | List team users |
| POST | `/hub/users/invite` | Invite a user |
| PUT | `/hub/users/permissions` | Update user app permissions / role |
| GET | `/hub/me` | Current user hub profile |

### ERP module (OneKit — `/erp/*`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/erp/product/v1/products` | List products |
| POST | `/erp/product/v1/products` | Create product |
| GET | `/erp/product/v1/products/{id}` | Get product |
| PUT | `/erp/product/v1/products/{id}` | Update product |
| DELETE | `/erp/product/v1/products/{id}` | Delete product |
| GET | `/erp/product/v1/dashboard` | ERP dashboard stats |
| GET | `/erp/supplier/v1/suppliers` | List suppliers |
| POST | `/erp/supplier/v1/suppliers` | Create supplier |
| GET | `/erp/purchase-order/v1/purchase-orders` | List purchase orders |
| PATCH | `/erp/purchase-order/v1/purchase-orders/{id}/status` | Update PO status |
| POST | `/erp/inventory/v1/adjust` | Adjust inventory |
| GET | `/erp/inventory/v1/report` | Inventory report |
| GET | `/erp/inventory/v1/low-stock` | Low-stock items |

Legacy chi `/erp/*` routes are removed — use OneKit paths above. Voxie protos under `proto/erp/` are the source of truth.

### HR module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/hr/dashboard` | HR dashboard stats |
| GET | `/hr/staff` | List staff |
| POST | `/hr/staff` | Create staff member |
| PUT | `/hr/staff` | Update staff member |
| DELETE | `/hr/staff` | Delete staff member |
| GET | `/hr/departments` | List departments |
| POST | `/hr/departments` | Create department |
| PUT | `/hr/departments` | Update department |
| GET | `/hr/shifts` | List shifts |
| POST | `/hr/shifts` | Create shift |
| PUT | `/hr/shifts` | Update shift |
| POST | `/hr/shifts/clock-in` | Clock in |
| POST | `/hr/shifts/clock-out` | Clock out |
| GET | `/hr/time-off` | List time-off requests |
| POST | `/hr/time-off` | Submit time-off request |
| PUT | `/hr/time-off/approve` | Approve/reject time-off |
| GET | `/hr/payroll` | List payroll records |
| POST | `/hr/payroll/run` | Run payroll for period |
| PUT | `/hr/payroll/mark-paid` | Mark payroll as paid |

### CRM module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/crm/dashboard` | CRM dashboard |
| GET | `/crm/customers` | List customers |
| POST | `/crm/customers` | Create customer |
| PUT | `/crm/customers` | Update customer |
| DELETE | `/crm/customers` | Delete customer |
| GET | `/crm/customers/detail` | Customer detail |
| POST | `/crm/customers/record-visit` | Record a customer visit |
| GET | `/crm/loyalty/config` | Get loyalty config |
| PUT | `/crm/loyalty/config` | Update loyalty config |
| POST | `/crm/loyalty/award` | Award loyalty points |
| POST | `/crm/loyalty/redeem` | Redeem loyalty points |
| GET | `/crm/loyalty/transactions` | Loyalty transaction history |
| GET | `/crm/segments` | List segments |
| POST | `/crm/segments` | Create segment |
| PUT | `/crm/segments` | Update segment |
| DELETE | `/crm/segments` | Delete segment |
| GET | `/crm/campaigns` | List campaigns |
| POST | `/crm/campaigns` | Create campaign |
| PUT | `/crm/campaigns` | Update campaign |
| PUT | `/crm/campaigns/status` | Update campaign status |

### Finance module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/finance/dashboard` | Finance dashboard |
| GET | `/finance/invoices` | List invoices |
| POST | `/finance/invoices` | Create invoice |
| PUT | `/finance/invoices` | Update invoice |
| DELETE | `/finance/invoices` | Delete invoice |
| GET | `/finance/invoices/detail` | Get invoice detail |
| PUT | `/finance/invoices/status` | Update invoice status |
| POST | `/finance/invoices/payment` | Record payment |
| GET | `/finance/expenses` | List expenses |
| POST | `/finance/expenses` | Create expense |
| PUT | `/finance/expenses` | Update expense |
| DELETE | `/finance/expenses` | Delete expense |
| GET | `/finance/expense-categories` | List expense categories |
| POST | `/finance/expense-categories` | Create expense category |
| GET | `/finance/pl` | Profit & loss report |
| GET | `/finance/tax/config` | Get tax config |
| PUT | `/finance/tax/config` | Update tax config |
| GET | `/finance/tax/report` | Tax report |

### POS module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/pos/dashboard` | POS dashboard |
| GET | `/pos/locations` | List locations |
| GET | `/pos/tables` | List tables |
| POST | `/pos/tables` | Create table |
| PUT | `/pos/tables` | Update table |
| PUT | `/pos/tables/status` | Update table status |
| DELETE | `/pos/tables` | Delete table |
| GET | `/pos/sessions/current` | Get current open session |
| POST | `/pos/sessions/open` | Open session |
| POST | `/pos/sessions/close` | Close session |
| GET | `/pos/sessions` | List sessions |
| GET | `/pos/orders` | List orders |
| GET | `/pos/orders/detail` | Get order detail |
| PUT | `/pos/orders/status` | Update order status |
| PUT | `/pos/orders/item-status` | Update order item status |
| POST | `/pos/orders/payment` | Process payment |
| POST | `/pos/orders/refund` | Process refund |

### Flow module

| Method | Path | Description |
|--------|------|-------------|
| GET | `/flow/dashboard` | Flow dashboard |
| GET | `/flow/workspaces` | List workspaces |
| POST | `/flow/workspaces` | Create workspace |
| PUT | `/flow/workspaces` | Update workspace |
| DELETE | `/flow/workspaces` | Delete workspace |
| POST | `/flow/workspaces/member` | Add workspace member |
| DELETE | `/flow/workspaces/member` | Remove workspace member |
| GET | `/flow/projects` | List projects |
| POST | `/flow/projects` | Create project |
| PUT | `/flow/projects` | Update project |
| DELETE | `/flow/projects` | Delete project |
| GET | `/flow/tasks` | List tasks |
| POST | `/flow/tasks` | Create task |
| GET | `/flow/tasks/detail` | Get task detail |
| PUT | `/flow/tasks` | Update task |
| DELETE | `/flow/tasks` | Delete task |
| PUT | `/flow/tasks/status` | Update task status |
| GET | `/flow/tasks/my` | Get my tasks |
| POST | `/flow/tasks/comment` | Add comment |
| DELETE | `/flow/tasks/comment` | Delete comment |
| GET | `/flow/labels` | List labels |
| POST | `/flow/labels` | Create label |

---

## Adding a New Module

1. Create the directory structure:
   ```
   internal/modules/mymodule/
   ├── module.go
   └── internal/
       ├── domain/       model.go, repository.go
       ├── application/  service.go
       ├── infrastructure/ repository.go (SQL impl)
       └── routes/handlers/ list_handler.go, create_handler.go, ...
   ```

2. Implement `module.go`:
   ```go
   package mymodule

   func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
       routes.RegisterRoutes(ctx, router, m.Service)
   }

   func (m *Module) GetAnonymousRoutes() []string {
       return []string{} // paths that skip auth
   }
   ```

3. Add the module to `cmd/api/main.go` in `initializeModules()` and add a field to `apiModules`.

4. Write SQL schema in `internal/platform/database/schemas/mymodule/schema.sql` and run `scripts/migrate.sh`.
