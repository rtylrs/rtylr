# rtylr

All-in-one business operating system for SMBs. 15 apps, 10 shared packages.

## Architecture

Frontend monorepo with 15 apps and 10 shared packages. pnpm workspaces + Turborepo.

**Stack:** React 19, Vite 7, TypeScript 5.9, Tailwind CSS, TanStack Query v5, Zustand, neverthrow, motion

## Apps

| App | Port | Package | Description |
|-----|------|---------|-------------|
| Hub | 5175 | `hub` | Owner command center, dashboard, team management, settings |
| Auth | 5176 | `@rtylr/auth` | Login, register, password reset, invite accept |
| Menu | 5177 | `@rtylr/menu` | Customer-facing digital menu with QR codes |
| POS | 5178 | `vbitez` | Point of sale terminal (Tauri desktop) |
| ERP | 5179 | `@rtylr/erp` | Inventory, suppliers, purchasing, stock management |
| HR | 5180 | `@rtylr/hr` | Staff, shifts, payroll, time off, reviews |
| CRM | 5181 | `@rtylr/crm` | Customers, loyalty, campaigns, segments |
| Finance | 5182 | `@rtylr/finance` | Invoices, expenses, P&L, tax, reconciliation |
| Flow | 5183 | `@rtylr/flow` | Projects, tasks, roadmaps, cycles, team workload |
| Store | 5184 | `@rtylr/store` | Online storefront, products, orders, checkout |
| Book | 5185 | `@rtylr/book` | Appointments, scheduling, resources, waitlist |
| Support | 5186 | `@rtylr/support` | Help desk, tickets, shared inbox, knowledge base |
| Insights | 5187 | `@rtylr/insights` | Cross-app analytics, revenue, operations, reports |
| Docs | 5188 | `@rtylr/docs` | Public documentation (no auth required) |
| Landing | 4200 | `@rtylr/landing` | Marketing site (Next.js 15) |

## Shared Packages

| Package | Purpose |
|---------|---------|
| `@voxire/api` | Axios client factory, neverthrow `safeFetch` wrapper, `unwrap()` for React Query |
| `@voxire/config` | `APP_URLS`, `APP_CONFIG`, theme injection, brand colors |
| `@voxire/types` | `UserRole`, `AppId`, `User` interface |
| `@voxire/ui` | Radix-based primitives (Dialog, Select, Switch, etc.) |
| `@voxire/components` | HelmetContainer, Button, Modal, Spinner, FullScreenLoader |
| `@voxire/auth` | Cookie helpers, auth guards, user store |
| `@voxire/utils` | `useDebounce`, `useMediaQuery`, `useLocalStorage`, formatting, validation |
| `@voxire/assets` | Shared SCSS variables and static assets |
| `@voxire/menu` | Customer-facing menu page components |
| `@voxire/tailwind-config` | Shared Tailwind CSS preset |

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
  status: number;                    // HTTP status (0 for network errors)
  code: ApiErrorCode;                // NETWORK_ERROR | UNAUTHORIZED | VALIDATION | ...
  message: string;                   // User-friendly message
}
```

Each app's `lib/api.ts`:
```typescript
import { createApiClient, safeFetch } from '@voxire/api';
const api = createApiClient('store');    // raw axios (backwards compat)
export const safe = safeFetch(api);      // neverthrow-wrapped
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

## Addon System

Core apps (POS, ERP, HR, CRM, Finance, Flow) are always enabled. Store, Book, Support, and Insights are addon-gated and can be toggled per business from the Hub Addons page.

## Development

```bash
# Install dependencies
pnpm install

# Run all apps in dev mode
pnpm dev

# Run a single app
pnpm --filter @rtylr/store dev

# Build all
pnpm build

# Typecheck all
pnpm turbo typecheck

# Build + typecheck a single app
pnpm --filter @rtylr/crm build
```

