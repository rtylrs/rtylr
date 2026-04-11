# Development Guide

## Prerequisites

| Tool    | Version | Install                       |
| ------- | ------- | ----------------------------- |
| Node.js | 20+     | https://nodejs.org            |
| pnpm    | 10.33.0 | `npm install -g pnpm@10.33.0` |
| Go      | 1.25+   | https://go.dev/dl/            |
| MySQL   | 8.0+    | Local or remote               |
| Redis   | 7+      | Optional — can be disabled    |

---

## Repository Layout

```
rtylr/
├── onion/        # Go backend
└── rtylr/        # pnpm + Turborepo frontend monorepo
```

Both are independent projects. The frontend (`rtylr/`) talks to the backend (`onion/`) over HTTP.

---

## Starting the Frontend Apps

From `rtylr/rtylr/`:

```bash
# Install dependencies (first time or after package changes)
pnpm install

# Start all apps in parallel (Turborepo)
pnpm dev
```

Turborepo starts every app in `apps/` concurrently. Ports:

| App     | Port |
| ------- | ---- |
| Hub     | 5175 |
| Auth    | 5176 |
| Menu    | 5177 |
| POS     | 5178 |
| ERP     | 5179 |
| HR      | 5180 |
| CRM     | 5181 |
| Finance | 5182 |
| Flow    | 5183 |
| Landing | 4200 |

To run only one app:

```bash
pnpm --filter hub dev         # run Hub only
pnpm --filter pos dev         # run POS only
pnpm --filter @voxire/auth dev  # run a package in watch mode
```

To build all apps:

```bash
pnpm build
```

Turborepo caches build outputs in `node_modules/.cache/turbo`. It builds packages before apps automatically (`dependsOn: ["^build"]`).

---

## Starting the Backend

From `rtylr/onion/`:

```bash
# Install Go dependencies
go mod download

# Copy and fill in the env file
cp .env.example .env  # if one exists, otherwise create it

# Run the server (port 9999)
go run ./cmd/api/
```

Or build a binary:

```bash
go build -o bin/api ./cmd/api/
./bin/api
```

The server restarts automatically on panic (5-second delay with retry loop).

---

## Backend Environment Variables

Create `onion/.env`:

```env
APP_ENV=development

# MySQL
DB_HOST=localhost
DB_PORT=3306
DB_NAME=voxirete_rtylr_db
DB_USER=your_db_user
DB_PASSWORD=your_db_password

# JWT
JWT_SECRET=your_jwt_secret

# Redis (optional)
REDIS_ENABLED=false

# In-memory cache
INMEMORY_CACHE_MAX_SIZE_MB=64

# AI integrations (optional)
OPENAI_API_KEY=
GEMINI_API_KEY=
ELEVENLABS_API_KEY=
```

---

## Frontend Environment Variables

Each Vite app reads its own `.env` file. The only env var typically needed:

```env
# apps/<appname>/.env.local
VITE_API_URL=https://api.voxire.dev/
```

If you are running the backend locally on port 9999:

```env
VITE_API_URL=http://localhost:9999/
```

All other configuration (app URLs, port assignments) is handled automatically by `@voxire/config` based on `window.location.hostname`.

---

## Running Database Migrations

From `rtylr/onion/`:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=your_user
export DB_PASSWORD=your_password
export DB_NAME=voxirete_rtylr_db

./scripts/migrate.sh
```

Migrations apply `IF NOT EXISTS` so they are safe to re-run. Order: `shared → rtylr → voxire → erp → hr → crm → finance → flow → pos → hub`.

To reset and reapply from scratch (destructive):

```bash
./scripts/reset-db.sh   # drops and recreates the database
./scripts/migrate.sh    # then re-apply schemas
```

---

## Adding a New Frontend App

1. Create the app directory:

   ```bash
   mkdir -p rtylr/apps/myapp/src
   ```

2. Create `apps/myapp/package.json`:

   ```json
   {
     "name": "myapp",
     "version": "0.0.0",
     "private": true,
     "scripts": {
       "dev": "vite --port 5184",
       "build": "tsc && vite build",
       "preview": "vite preview"
     },
     "dependencies": {
       "@voxire/auth": "workspace:*",
       "@voxire/api": "workspace:*",
       "@voxire/types": "workspace:*",
       "@voxire/config": "workspace:*"
     }
   }
   ```

3. Add the new `AppId` to `@voxire/types`:

   ```ts
   // packages/types/src/user.ts
   export type AppId = 'hub' | 'auth' | ... | 'myapp';
   ```

4. Add the app to `@voxire/config`:

   ```ts
   // packages/config/src/apps.ts
   export const APP_URLS: Record<AppId, string> = {
     // ...existing entries
     myapp: IS_LOCAL ? 'http://localhost:5184' : 'https://myapp.rtylr.com',
   };

   export const APP_CONFIG: Record<
     AppId,
     { name: string; devPort: number; allowedRoles: UserRole[] }
   > = {
     // ...existing entries
     myapp: { name: 'My App', devPort: 5184, allowedRoles: ['owner', 'admin'] },
   };
   ```

5. Add the app to `pnpm-workspace.yaml` if it is not already covered by the glob.

6. Wire up auth in your app entry point:
   ```tsx
   import { ProtectedRoute } from '@voxire/auth';
   <ProtectedRoute appId="myapp">
     <App />
   </ProtectedRoute>;
   ```

---

## Adding a New Backend Module

1. Create the directory tree:

   ```
   onion/internal/modules/mymodule/
   ├── module.go
   └── internal/
       ├── domain/
       │   ├── model.go       # structs, error sentinels
       │   └── repository.go  # interface
       ├── application/
       │   └── service.go     # business logic
       ├── infrastructure/
       │   └── repository.go  # SQL implementation
       └── routes/
           └── handlers/
               ├── list_handler.go
               └── create_handler.go
   ```

2. Write `module.go`:

   ```go
   package mymodule

   import (
       "context"
       "github.com/go-chi/chi/v5"
       "github.com/voxire/onion/internal/modules/mymodule/internal/application"
       "github.com/voxire/onion/internal/modules/mymodule/internal/routes"
   )

   type Module struct {
       Service *application.Service
   }

   func NewModule(ctx context.Context) *Module {
       return &Module{Service: application.NewService(ctx)}
   }

   func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
       routes.RegisterRoutes(ctx, router, m.Service)
   }

   func (m *Module) GetAnonymousRoutes() []string {
       return []string{}
   }
   ```

3. Write each handler implementing `HandlerFunc[Req, Res]`:

   ```go
   type ListThingsHandler struct{ Service *application.Service }
   func ListThingsRoutes() []string { return []string{"/mymodule/things"} }
   func (h *ListThingsHandler) Serve(ctx context.Context, req *ListThingsRequest) (ListThingsResponse, *errors.ServerError) {
       // ...
   }
   ```

4. Register in `routes/routes.go`:

   ```go
   func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
       http.RegisterGET[handlers.ListThingsRequest, handlers.ListThingsResponse](
           ctx, router, handlers.ListThingsRoutes(), &handlers.ListThingsHandler{Service: svc},
       )
       // ...
   }
   ```

5. Add to `cmd/api/main.go`:

   ```go
   // in apiModules struct
   MyModule *mymodule.Module

   // in initializeModules()
   MyModule: mymodule.NewModule(ctx),
   ```

6. Write schema and run migrations:
   ```bash
   mkdir -p internal/platform/database/schemas/mymodule
   # write schema.sql
   ./scripts/migrate.sh
   ```

---

## Common Scripts

All scripts are in `onion/scripts/`:

| Script              | Purpose                             |
| ------------------- | ----------------------------------- |
| `migrate.sh`        | Apply all schema SQL files in order |
| `reset-db.sh`       | Drop and recreate the database      |
| `setup.sh`          | Full environment bootstrap          |
| `lint.sh`           | Run Go linters                      |
| `test.sh`           | Run Go tests                        |
| `build.sh`          | Build the Go binary                 |
| `generate_sqlc.sh`  | Run sqlc code generation            |
| `check_sql_case.sh` | Check SQL file naming conventions   |

Frontend scripts (from `rtylr/rtylr/`):

```bash
pnpm dev          # start all apps
pnpm build        # build all apps
pnpm typecheck    # type-check all packages and apps
pnpm test         # run all tests
pnpm lint         # lint all packages
```

---

## Tech Stack Reference

| Layer              | Technology                                          |
| ------------------ | --------------------------------------------------- |
| Frontend framework | React 19                                            |
| Bundler            | Vite 7 (Turborepo orchestration)                    |
| Landing            | Next.js 15                                          |
| Package manager    | pnpm 10                                             |
| State management   | Zustand 5 (global), TanStack Query 5 (server state) |
| HTTP client        | Axios (via `@voxire/api`)                           |
| Routing            | react-router-dom v6                                 |
| Backend language   | Go 1.25                                             |
| Backend router     | chi v5                                              |
| Database           | MySQL 8                                             |
| Cache              | Redis + bigcache                                    |
| Auth               | JWT (golang-jwt/jwt v5)                             |
| Query builder      | squirrel                                            |
| TypeScript         | 5.9                                                 |
