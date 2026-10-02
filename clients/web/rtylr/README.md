# rtylr

The business workspace for RTYLR. Portal brings Home, Insights and Flow into one browser app.

## Architecture

pnpm workspace with shared libraries and standalone products migrating into Portal.

**Stack:** React 19, Vite 8, TypeScript 6 (strict), Tailwind CSS, TanStack Query v5, Zustand, neverthrow, motion

## Apps

| App     | Port | Package          | Description                                             |
| ------- | ---- | ---------------- | ------------------------------------------------------- |
| Portal  | 5175 | `@rtylr/portal`  | Home, team, groups, account settings, Insights and Flow |
| Auth    | 5176 | `@rtylr/auth`    | Login, register, password reset, invite accept          |
| Menu    | 5177 | `@rtylr/menu`    | Customer-facing digital menu with QR codes              |
| POS     | 5178 | `vbitez`         | Point of sale terminal (Tauri v2 desktop)               |
| ERP     | 5179 | `@rtylr/erp`     | Inventory, suppliers, purchasing, stock management      |
| HR      | 5180 | `@rtylr/hr`      | Staff, shifts, payroll, time off, reviews               |
| CRM     | 5181 | `@rtylr/crm`     | Customers, loyalty, campaigns, segments                 |
| Finance | 5182 | `@rtylr/finance` | Invoices, expenses, P&L, tax, reconciliation            |
| Recruit | 5184 | `@rtylr/recruit` | Candidate careers sites + HR submissions inbox          |
| Docs    | 5186 | `@rtylr/docs`    | Internal + public product documentation                 |
| Landing | 4200 | `@rtylr/landing` | Marketing site - rtylr.com (Next.js 15 static export)   |

Portal runs at `portal.rtylr.com`; Insights lives under `/insights` with its own
header inside the shared workspace. Flow lives under `/flow`; app settings share
`/settings/flow` and `/settings/insights`. Auth stays at `sso.rtylr.com`. The other
workspace apps currently open separately while their modules move into Portal.
Menu and Recruit are setup/public products, not entries in the workspace launcher.
See [Portal architecture](docs/portal/architecture.md) for module boundaries and cutover order.

## Shared Packages

| Package                   | Purpose                                                                          |
| ------------------------- | -------------------------------------------------------------------------------- |
| `@voxire/api`             | Axios client factory, neverthrow `safeFetch` wrapper, `unwrap()` for React Query |
| `@voxire/permissions`     | Typed RBAC with platform roles, POS roles, and 100+ granular capabilities        |
| `@voxire/config`          | `APP_URLS`, `APP_CONFIG`, theme injection, brand colors                          |
| `@voxire/types`           | `UserRole`, `AppId`, `User` interface                                            |
| `@voxire/voxie`           | Generated TypeScript types and clients from voxie protobuf contracts             |
| `@voxire/ui`              | Radix-based primitives (Dialog, Select, Switch, etc.)                            |
| `@voxire/components`      | HelmetContainer, Button, Modal, Spinner, FullScreenLoader                        |
| `@voxire/auth`            | Cookie helpers, `AppGuard` with server-resolved effective access, user store     |
| `@voxire/utils`           | `useDebounce`, `useMediaQuery`, `useTheme`, formatting, validation               |
| `@voxire/i18n`            | Shared i18next setup and translation helpers                                     |
| `@voxire/menu`            | Customer-facing menu page components                                             |
| `@voxire/tailwind-config` | Shared Tailwind v4 theme tokens                                                  |

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
  router.tsx              # route definitions
  main.tsx
  globals.css
```

## API Contracts (voxie)

API request/response types are defined as `.onk` contracts in `voxie/api/`, shared between frontend and backend as the single source of truth.

```
voxie/api/
  common/v1/             # Pagination, errors, timestamps
  hub/business/v1/       # Team, business, apps, dashboard, billing, developer
  ...
```

### Setup

```bash
pnpm install
make -C ../../voxie generate   # gen/ is generated, not committed
```

### Updating contracts

```bash
# Edit .onk files under voxie/api/, regenerate, and prove it end-to-end in the
# same change. Contracts, generated consumers and Go handlers land together:
make -C ../../voxie generate
pnpm typecheck
cd ../../services/onion && go build ./...
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
pnpm typecheck

# Run tests
pnpm test

# Run tests with coverage
pnpm test:coverage

# Format
pnpm format
```
