# rtylr

The all-in-one business operating system for SMBs. 11 apps, 11 shared packages, 1,250+ tests.

## Architecture

Frontend monorepo with 11 apps and 11 shared packages. pnpm workspaces + Turborepo.

**Stack:** React 19, Vite 7, TypeScript 5.9 (strict), Tailwind CSS, TanStack Query v5, Zustand, neverthrow, motion

## Apps

| App      | Port | Package           | Description                                                |
| -------- | ---- | ----------------- | ---------------------------------------------------------- |
| Hub      | 5175 | `hub`             | Owner command center, dashboard, team management, settings |
| Auth     | 5176 | `@rtylr/auth`     | Login, register, password reset, invite accept             |
| Menu     | 5177 | `@rtylr/menu`     | Customer-facing digital menu with QR codes                 |
| POS      | 5178 | `vbitez`          | Point of sale terminal (Tauri v2 desktop)                  |
| ERP      | 5179 | `@rtylr/erp`      | Inventory, suppliers, purchasing, stock management         |
| HR       | 5180 | `@rtylr/hr`       | Staff, shifts, payroll, time off, reviews                  |
| CRM      | 5181 | `@rtylr/crm`      | Customers, loyalty, campaigns, segments                    |
| Finance  | 5182 | `@rtylr/finance`  | Invoices, expenses, P&L, tax, reconciliation               |
| Flow     | 5183 | `@rtylr/flow`     | Projects, tasks, roadmaps, cycles, team workload           |
| Insights | 5187 | `@rtylr/insights` | Cross-app analytics, revenue, operations, reports          |
| Landing  | 4200 | `@rtylr/landing`  | Marketing site (Next.js 15)                                |

## Shared Packages

| Package                   | Purpose                                                                          |
| ------------------------- | -------------------------------------------------------------------------------- |
| `@voxire/api`             | Axios client factory, neverthrow `safeFetch` wrapper, `unwrap()` for React Query |
| `@voxire/permissions`     | Typed RBAC with platform roles, POS roles, and 100+ granular capabilities        |
| `@voxire/config`          | `APP_URLS`, `APP_CONFIG`, theme injection, brand colors                          |
| `@voxire/types`           | `UserRole`, `AppId`, `User` interface                                            |
| `@voxire/ui`              | Radix-based primitives (Dialog, Select, Switch, etc.)                            |
| `@voxire/components`      | HelmetContainer, Button, Modal, Spinner, FullScreenLoader                        |
| `@voxire/auth`            | Cookie helpers, auth guards, user store                                          |
| `@voxire/utils`           | `useDebounce`, `useMediaQuery`, `useTheme`, formatting, validation               |
| `@voxire/assets`          | Shared SCSS variables and static assets                                          |
| `@voxire/menu`            | Customer-facing menu page components                                             |
| `@voxire/tailwind-config` | Shared Tailwind CSS preset                                                       |

## Project Structure (per app)

Every app follows the feature-based structure:

```
apps/{name}/src/
  features/
    dashboard/
      Dashboard.tsx        # screen component
      useDashboard.ts      # co-located hook
    products/
      Products.tsx
      useProducts.ts
    ...
  shared/
    components/            # cross-feature components
    lib/                   # api.ts, format.ts, utilities
    services/              # API service files (safeFetch)
    hooks/                 # cross-feature hooks
    stores/                # Zustand stores
  layouts/
    DefaultLayout.tsx      # single file, no wrapper folder
  routes/
    AppRoutes.tsx
  App.tsx
  main.tsx
  globals.css
```

## API Contracts (voxie)

API request/response types are defined as Protocol Buffers in the [voxie](https://github.com/voxire/voxie) submodule, shared between frontend and backend as the single source of truth.

```
voxie/proto/
  common/v1/             # Pagination, errors, timestamps
  auth/v1/               # Login, register, password reset
  hub/v1/                # Team, business, apps, dashboard, billing, developer
```

### Setup

```bash
# Clone with submodules
git clone --recurse-submodules git@github.com:voxire/rtylr.git

# Or if already cloned
git submodule update --init --recursive

# Generate TypeScript types from proto files
cd voxie && make generate
```

### Updating contracts

```bash
cd voxie
# Edit proto files in proto/{module}/v1/
make generate    # Regenerate Go + TypeScript types
git add -A && git commit -m "feat: add X contract"
git push

# Then update the submodule ref in frontend + backend
cd .. && git add voxie && git commit -m "chore: update voxie submodule"
```

## API Layer

All API calls use neverthrow for type-safe error handling:

```typescript
// packages/api/src/result.ts
interface SafeApi {
  get<T>(url: string, config?): ResultAsync<T, ApiError>;
  post<T>(url: string, data?, config?): ResultAsync<T, ApiError>;
  put<T>(url: string, data?, config?): ResultAsync<T, ApiError>;
  delete<T>(url: string, config?): ResultAsync<T, ApiError>;
}

interface ApiError {
  status: number; // HTTP status (0 for network errors)
  code: ApiErrorCode; // NETWORK_ERROR | UNAUTHORIZED | VALIDATION | ...
  message: string; // User-friendly message
}
```

Each app's `lib/api.ts`:

```typescript
import { createApiClient, safeFetch } from '@voxire/api';
const api = createApiClient('crm'); // raw axios (backwards compat)
export const safe = safeFetch(api); // neverthrow-wrapped
export default api;
```

Hooks use `unwrap()` to bridge with React Query:

```typescript
import { unwrap } from '@voxire/api';

export function useProducts() {
  return useQuery({
    queryKey: ['products'],
    queryFn: () => unwrap(storeApi.listProducts()),
  });
}
```

## Development

```bash
# Install dependencies
pnpm install

# Run all apps in dev mode
pnpm dev

# Run a single app
pnpm --filter @rtylr/crm dev

# Build all
pnpm build

# Typecheck all (strict mode)
pnpm turbo typecheck

# Run tests (1,250+ unit tests)
pnpm test

# Run tests with coverage
pnpm test:coverage

# Format
pnpm format
```

## Testing

- **Unit tests:** Vitest + React Testing Library (1,250+ tests)
- **E2E tests:** Playwright (auth, hub, landing flows)
- **Coverage:** stores ~95%, services ~85%, hooks ~80%, utils ~90%
- **CI pipeline:** format check -> typecheck (strict) -> unit tests -> lint -> build
