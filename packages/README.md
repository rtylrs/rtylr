# @voxire/packages

Shared frontend libraries for the rtylr platform. Workspace members of the
`rtylr` monorepo, consumed by the apps through `workspace:*`; not published to
npm.

## Packages

| Package                   | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `@voxire/api`             | API client, realtime WebSocket, result helpers |
| `@voxire/auth`            | Auth store, route guards, cookie helpers       |
| `@voxire/caret`           | Caret AI chat components and hooks             |
| `@voxire/components`      | Core UI component library                      |
| `@voxire/config`          | App config, feature flags, app guards          |
| `@voxire/consent`         | Cookie consent banner and `useConsent` hook    |
| `@voxire/voxie`           | Shared TypeScript types for all API contracts  |
| `@voxire/i18n`            | i18next setup and language definitions         |
| `@voxire/menu`            | Public digital menu components and layouts     |
| `@voxire/observability`   | Sentry + PostHog boot helpers                  |
| `@voxire/permissions`     | Role/capability matrix and guards              |
| `@voxire/tailwind-config` | Shared Tailwind v4 theme tokens                  |
| `@voxire/types`           | Shared domain types (business, user, etc.)     |
| `@voxire/ui`              | Headless UI primitives (shadcn-based)          |
| `@voxire/utils`           | Formatting, validation, and common hooks       |

## Contracts

`@voxire/voxie` exposes generated contract types from the in-tree `voxie/`
module. Generated output is not committed; run `make -C voxie generate` (or any
rtylr `typecheck`/`build`, which generates it when missing) before typechecking
a package that imports it.

## Usage

Add a dependency with `workspace:*` and run `pnpm install` from the repository
root. A single lockfile keeps one copy of each dependency and type identity
across packages — see the root `README.md`.
