# @voxire/packages

Shared frontend packages for the voxire platform. This directory is part of the rtylr workspace; only `voxie/` is a git submodule.

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

`voxie/` is a nested git submodule pointing at the generated contract source.
Edit and regenerate contracts in the Voxie repository; update this repository
with a submodule pointer only. The pointer may intentionally refer to an older
Voxie commit; consumers do not have to follow Voxie main immediately. Do not
copy contract or generated files into this repository.

## Usage

The packages are resolved from the workspace by `pnpm`. Only the Voxie contract submodule needs initialization:

```bash
git submodule update --init --recursive
```

### Updating contracts

```bash
git submodule update --remote packages/voxie
git add packages/voxie && git commit -m "chore: bump Voxie contract submodule"
```
