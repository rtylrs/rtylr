# Architecture

rtylr is an all-in-one business operating system. "Buy rtylr, buy a whole business powered by AI." It is structured as a pnpm monorepo managed with Turborepo on the frontend, and a single Go backend called **onion** that serves all modules through one API.

## Repository Layout

```
rtylr/
├── onion/                     # Go backend (all modules in one process)
│   ├── cmd/api/main.go        # Entry point, server wires all modules
│   ├── internal/
│   │   ├── modules/           # Domain modules - one module per bounded context (products, inventory, staff, customers, invoices, etc.)
│   │   └── platform/          # Cross-cutting concerns (auth, db, cache, http, logger)
│   └── scripts/               # migrate.sh, setup.sh, lint.sh, etc.
└── rtylr/                     # Turborepo monorepo (frontend)
    ├── apps/
    │   ├── hub/               # Port 5175 - Command center
    │   ├── auth/              # Port 5176 - SSO login/register
    │   ├── menu/              # Port 5177 - Customer-facing digital menu
    │   ├── pos/               # Port 5178 - Point of sale
    │   ├── erp/               # Port 5179 - Inventory & purchasing
    │   ├── hr/                # Port 5180 - Staff & payroll
    │   ├── crm/               # Port 5181 - Customers & loyalty
    │   ├── finance/           # Port 5182 - Invoices & expenses
    │   ├── flow/              # Port 5183 - Projects & tasks
    │   ├── landing/           # Port 4200 - Marketing site (Next.js)
    │   └── vbitez/            # Alias app (legacy)
    └── packages/
        ├── api/               # @voxire/api    - Axios client
        ├── auth/              # @voxire/auth   - Cookie helpers, guards, user store
        ├── types/             # @voxire/types  - Shared TypeScript types
        ├── config/            # @voxire/config - APP_URLS, APP_CONFIG
        ├── ui/                # @voxire/ui     - Shared UI components
        ├── components/        # @voxire/components - Additional components
        ├── assets/            # @voxire/assets
        ├── menu/              # @voxire/menu   - Shared menu components
        ├── tailwind-config/   # Shared Tailwind config
        └── utils/             # @voxire/utils
```

## App Port Assignments

| App     | Local Port | Production Domain | Framework    |
| ------- | ---------- | ----------------- | ------------ |
| Hub     | 5175       | hub.rtylr.com     | React + Vite |
| Auth    | 5176       | auth.rtylr.com    | React + Vite |
| Menu    | 5177       | menu.rtylr.com    | React + Vite |
| POS     | 5178       | pos.rtylr.com     | React + Vite |
| ERP     | 5179       | erp.rtylr.com     | React + Vite |
| HR      | 5180       | hr.rtylr.com      | React + Vite |
| CRM     | 5181       | crm.rtylr.com     | React + Vite |
| Finance | 5182       | finance.rtylr.com | React + Vite |
| Flow    | 5183       | flow.rtylr.com    | React + Vite |
| Landing | 4200       | rtylr.com         | Next.js 15   |
| API     | 9999       | api.voxire.dev    | Go + Chi     |

## Frontend Stack

- **React 19** with `react-router-dom` v6
- **Vite 7** for bundling (all apps except Landing)
- **Next.js 15** for Landing
- **Turborepo 2** for monorepo task orchestration
- **pnpm 10** as package manager
- **Zustand 5** for state management (`@voxire/auth` user store)
- **TanStack Query 5** for server state / data fetching
- **TypeScript 5.9**

## Backend Stack

- **Go 1.25** with `github.com/go-chi/chi/v5` router
- **MySQL** (database at `108.167.180.218`, db `voxirete_rtylr_db`)
- **Redis** via `github.com/redis/go-redis/v9`
- **JWT** via `github.com/golang-jwt/jwt/v5`
- **bigcache** for in-memory caching
- **squirrel** for query building

## Auth Flow (ASCII)

```
User visits pos.rtylr.com
        │
        ▼
  Has `token` cookie?
        │
   No ──┴─► Redirect to auth.rtylr.com?redirect=pos
                    │
                    ▼
              Login / Register
                    │
                    ▼
         POST /login → API returns JWT
                    │
                    ▼
         setCookie("token", jwt, { domain: ".rtylr.com" })
                    │
                    ▼
         Redirect back to pos.rtylr.com
                    │
                    ▼
        ProtectedRoute reads cookie ✓
                    │
                    ▼
          App renders for user

Every API request:
  cookie token → Authorization: Bearer <token> header
  401 response → clear cookie → redirect to auth
```

## Shared Packages Dependency Graph

```
apps/*
  └─ @voxire/auth      (guards, user store, cookie helpers)
  └─ @voxire/api       (axios client - reads token from cookie)
  └─ @voxire/types     (TypeScript types for all domains)
  └─ @voxire/config    (APP_URLS, APP_CONFIG keyed by AppId)
  └─ @voxire/ui        (shared components)
  └─ @voxire/components
  └─ @voxire/assets
  └─ @voxire/utils
```

## Backend Module Architecture

The backend follows **domain-level modules** - each module is a single bounded domain context, NOT an app-level wrapper. A domain module owns its routes, service, repository, and domain types.

```
internal/modules/
├── user/          - auth, registration, profile, /user/me
├── config/        - business config and themes
├── careers/       - job openings (public)
├── echo/          - AI SSE streaming (depends on memory)
├── memory/        - AI conversation memory and context
├── products/      - product catalog CRUD
├── inventory/     - stock levels, adjustments, low-stock alerts
├── suppliers/     - supplier management
├── purchase_orders/ - purchasing from suppliers
├── staff/         - employee profiles (HR)
├── departments/   - org chart (HR)
├── shifts/        - scheduling + clock-in/out (HR)
├── payroll/       - payroll runs and records (HR)
├── time_off/      - leave requests and approvals (HR)
├── customers/     - customer records, visits, dashboard (CRM)
├── loyalty/       - points, rewards, config (CRM)
├── campaigns/     - marketing campaigns (CRM)
├── segments/      - customer segments (CRM)
├── invoices/      - invoice CRUD, payments, status (Finance)
├── expenses/      - expense tracking and categories (Finance)
├── tax/           - tax config + P&L report (Finance)
├── workspaces/    - project workspaces (Flow)
├── projects/      - projects within workspaces (Flow)
├── tasks/         - tasks + labels (Flow)
├── pos_orders/    - sales orders (POS)
├── pos_tables/    - table management (POS)
├── pos_sessions/  - cash register sessions (POS)
├── pos_locations/ - store locations (POS)
└── team/          - staff users, invites, app permissions (Hub)
```

## Backend Documentation

All backend-specific documentation lives in `onion/docs/`:

- [`onion/docs/backend.md`](../../onion/docs/backend.md) - Architecture, handler pattern, middleware, module structure
- [`onion/docs/database.md`](../../onion/docs/database.md) - Schema overview, migrations, multi-tenancy pattern
