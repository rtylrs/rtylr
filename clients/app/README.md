# rtylr

The business workspace for RTYLR. Portal brings every business app into one browser workspace.

## Architecture

One pnpm package with business modules, public entrypoints and shared libraries.

**Stack:** React 19, Vite 8, TypeScript 6 (strict), Tailwind CSS, TanStack Query v5, Zustand, neverthrow, motion

## Entry points

One package, `@rtylr/portal`, owns the browser applications:

| Entry point    | Dev command        | Port | Build output   |
| -------------- | ------------------ | ---- | -------------- |
| Portal         | `pnpm dev`         | 5175 | `dist/portal`  |
| SSO            | `pnpm dev:auth`    | 5176 | `dist/auth`    |
| Public Menu    | `pnpm dev:menu`    | 5174 | `dist/menu`    |
| Public Recruit | `pnpm dev:recruit` | 5184 | `dist/recruit` |

Portal contains Home, Insights, Flow, CRM, ERP, Finance, HR, Recruit administration, Caret and Automations. Each business app uses a route prefix and shares authentication, business selection, theme, launcher and `/settings`. Menu management remains in Home. SSO and public Menu/careers retain their public hosts and separate builds. POS web and native remain in `clients/pos`; MCP is in `clients/mcp`.

Automations uses `/automations` and `/automations/new?id=…`, with legacy CRM links redirected. It retains the CRM automation API and `crm:automations:manage` permission. New workflows configure order-triggered emails and start paused; other saved definitions remain editable with their configuration preserved.

`pnpm build` checks all entrypoints and shared dependencies before building each output. `pnpm test` runs the app and shared-package suites and verifies every collected file ran. Browser tests use isolated API stubs.

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

## Source structure

- `src/shell/`: Portal navigation, profile, business selection and launcher.
- `src/features/`: Home and shared account/business administration.
- `src/modules/`: business app features, routes, navigation and locale resources.
- `src/shared/`: cross-module stores, hooks and service helpers.
- `entrypoints/`: SSO and public Menu/Recruit bootstrap and build configuration.
- `e2e/`, `test/`, `scripts/`: browser coverage, shared fixtures and repository gates.

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

Use generated clients from `@voxire/api/onekit` through the module's existing API helpers. Contract changes belong in `voxie/api/`; regenerate with `make -C voxie generate`. Do not add raw transport calls or edit generated output.

Helpers return typed results. Hooks use `unwrap()` to bridge them with React Query:

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
pnpm --filter @rtylr/portal dev

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

For a same-origin local preview against the hosted API, run `VITE_API_URL=/api VITE_API_PROXY_TARGET=https://api.voxire.com pnpm dev`.
