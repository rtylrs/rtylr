# @voxire/packages

Shared frontend packages for the voxire platform. Used as a git submodule in `voxire/rtylr` and `voxire/pos`.

## Packages

| Package                   | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `@voxire/api`             | API client, realtime WebSocket, result helpers |
| `@voxire/assets`          | Icons, global SCSS variables                   |
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
| `@voxire/tailwind-config` | Shared Tailwind configuration                  |
| `@voxire/types`           | Shared domain types (business, user, etc.)     |
| `@voxire/ui`              | Headless UI primitives (shadcn-based)          |
| `@voxire/utils`           | Formatting, validation, and common hooks       |

`voxie/` is a nested git submodule pointing at the generated contract source.
Edit and regenerate contracts in the Voxie repository; update this repository
with a submodule pointer only. The pointer may intentionally refer to an older
Voxie commit; consumers do not have to follow Voxie main immediately. Do not
copy contract or generated files into this repository.

## Usage

This repo is consumed as a git submodule, not published to npm.

### Adding to a repo

```bash
git submodule add https://github.com/voxire/packages.git packages
git submodule update --init
```

Your `pnpm-workspace.yaml` must include `packages/*` so pnpm resolves `@voxire/*` imports locally.

### Updating to latest

```bash
git submodule update --remote packages
git add packages && git commit -m "chore: bump voxire/packages"
```

To update the nested contract submodule from this repository:

```bash
git submodule update --remote voxie
git add voxie && git commit -m "chore: bump Voxie contract submodule"
```

### CI (GitHub Actions / Cloud Build)

Use a PAT with `repo` scope stored in your secrets manager. Before running `git submodule update --init`, rewrite HTTPS URLs:

```bash
git config --global \
  url."https://x-access-token:${GITHUB_PAT}@github.com/".insteadOf \
  "https://github.com/"
git submodule update --init --recursive
```
