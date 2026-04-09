# Frontend Configuration

All frontend apps live in `rtylr/rtylr/apps/`. They are React + Vite SPAs (except `landing/` which is Next.js 15), deployed to Cloudflare Pages on individual subdomains.

---

## App Overview

| App       | Package Name     | Subdomain              | Local Port | Framework         |
|-----------|-----------------|------------------------|------------|-------------------|
| Hub       | `hub`           | `hub.rtylr.com`        | 5175       | React + Vite      |
| Auth      | `@rtylr/auth`   | `auth.rtylr.com`       | 5176       | React + Vite      |
| Menu      | `@rtylr/menu`   | `menu.rtylr.com`       | 5177       | React + Vite      |
| ERP       | `@rtylr/erp`    | `erp.rtylr.com`        | 5179       | React + Vite      |
| HR        | `@rtylr/hr`     | `hr.rtylr.com`         | 5180       | React + Vite      |
| CRM       | `@rtylr/crm`    | `crm.rtylr.com`        | 5181       | React + Vite      |
| Finance   | `@rtylr/finance`| `finance.rtylr.com`    | 5182       | React + Vite      |
| Flow      | `@rtylr/flow`   | `flow.rtylr.com`       | 5183       | React + Vite      |
| Landing   | `@rtylr/landing`| `rtylr.com`            | 4200       | Next.js 15        |
| vbitez    | `vbitez`        | (desktop app, not web) | 5178       | Tauri + React     |

> **vbitez is the POS.** It's a Tauri-based desktop app, not deployed to Cloudflare Pages. There is no `@rtylr/pos` web app — all POS functionality lives in vbitez.

---

## Cloudflare Pages Configuration (per app)

Each app is its own Cloudflare Pages project. **The git repo root is already the monorepo root** (it contains `pnpm-workspace.yaml`, `apps/`, `packages/`), so the Cloudflare Pages "Root directory" setting must be **empty** (or `/`). Do NOT set Root directory to `rtylr` — there is no `rtylr/` subdirectory inside the repo.

### Common Settings (all projects)

| Setting              | Value               |
|----------------------|---------------------|
| Framework preset     | None (custom)       |
| Root directory       | *(leave empty)*     |
| Node version         | `20`                |
| Package manager      | pnpm                |

> **Why Root directory must be empty:** The git repository's top-level directory (`voxire/rtylr` on GitHub) is already where `pnpm-workspace.yaml` lives, alongside `apps/` and `packages/`. Setting Root directory to anything else makes Cloudflare look in a subdirectory that doesn't exist, and the build is skipped because no files match the watch paths.

### Build Command and Output Directory (per app)

Paths are relative to the repo root (= the monorepo root).

| Pages Project   | Build Command                                                              | Build Output Directory |
|----------------|----------------------------------------------------------------------------|------------------------|
| `rtylr-hub`     | `pnpm install --frozen-lockfile && pnpm --filter hub run build`            | `apps/hub/dist`        |
| `rtylr-auth`    | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/auth run build`    | `apps/auth/dist`       |
| `rtylr-menu`    | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/menu run build`    | `apps/menu/dist`       |
| `rtylr-erp`     | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/erp run build`     | `apps/erp/dist`        |
| `rtylr-hr`      | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/hr run build`      | `apps/hr/dist`         |
| `rtylr-crm`     | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/crm run build`     | `apps/crm/dist`        |
| `rtylr-finance` | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/finance run build` | `apps/finance/dist`    |
| `rtylr-flow`    | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/flow run build`    | `apps/flow/dist`       |
| `rtylr-landing` | `pnpm install --frozen-lockfile && pnpm --filter @rtylr/landing run build` | `apps/landing/.next`   |

> **Note on hub:** the hub app's package name is just `hub` (not `@rtylr/hub`). Everything else (except vbitez) uses the `@rtylr/*` scope.

> **vbitez is not listed** because it's a desktop Tauri app, not deployed to Cloudflare Pages. vbitez is built and distributed as a native desktop binary via `pnpm --filter vbitez run tauri:build`.

> **Note on build output paths:** These are relative to the Root directory. Since Root directory is `rtylr`, `apps/flow/dist` resolves to the absolute path `rtylr/apps/flow/dist` on the build machine. Do NOT prefix these with `rtylr/` — that would produce `rtylr/rtylr/apps/flow/dist` which does not exist.

---

## Watch Paths (IMPORTANT — avoid unnecessary builds)

By default, Cloudflare Pages rebuilds an app on **every commit**, even if nothing in that app changed. This wastes build minutes. Fix it with **Include watch path** in each Pages project's Build configuration.

Go to: **Pages project → Settings → Build → Include paths**

Watch paths are relative to the Root directory (`rtylr`), so do **NOT** include the `rtylr/` prefix in these patterns.

Set these include paths per project:

| Pages Project   | Include paths                         |
|----------------|---------------------------------------|
| `rtylr-hub`     | `apps/hub/** packages/**`             |
| `rtylr-auth`    | `apps/auth/** packages/**`            |
| `rtylr-menu`    | `apps/menu/** packages/**`            |
| `rtylr-erp`     | `apps/erp/** packages/**`             |
| `rtylr-hr`      | `apps/hr/** packages/**`              |
| `rtylr-crm`     | `apps/crm/** packages/**`             |
| `rtylr-finance` | `apps/finance/** packages/**`         |
| `rtylr-flow`    | `apps/flow/** packages/**`            |
| `rtylr-landing` | `apps/landing/** packages/**`         |

**Why include `packages/**`?** Changes to shared packages (`@voxire/config`, `@voxire/ui`, etc.) should trigger rebuilds because they affect every app.

**Exclude paths** (optional, to explicitly ignore):

| Path            | Why                                              |
|----------------|--------------------------------------------------|
| `docs/**`       | Documentation changes                            |
| `**/*.md`       | All markdown files                               |

Setting the exclude path to `docs/** **/*.md` prevents builds for docs-only commits.

---

## Environment Variables (all apps)

Set these in each Cloudflare Pages project's Settings → Environment Variables:

| Variable       | Value                        | Notes                              |
|---------------|------------------------------|------------------------------------|
| `VITE_API_URL` | `https://api.voxire.dev/`   | All Vite apps — API base URL       |

> **Note:** `NEXT_PUBLIC_API_URL` for the landing app if you expose API calls from Next.js pages.

---

## Cloudflare Pages Projects

Create one Pages project per app. Recommended project names:

| App     | Pages Project Name   |
|---------|---------------------|
| Hub     | `rtylr-hub`         |
| Auth    | `rtylr-auth`        |
| Menu    | `rtylr-menu`        |
| ERP     | `rtylr-erp`         |
| HR      | `rtylr-hr`          |
| CRM     | `rtylr-crm`         |
| Finance | `rtylr-finance`     |
| Flow    | `rtylr-flow`        |
| Landing | `rtylr-landing`     |

---

## Custom Domains (Cloudflare Pages → Custom domains)

After creating each Pages project, add the custom domain in Pages → Custom domains:

| App     | Custom Domain          |
|---------|------------------------|
| Hub     | `hub.rtylr.com`        |
| Auth    | `auth.rtylr.com`       |
| Menu    | `menu.rtylr.com`       |
| ERP     | `erp.rtylr.com`        |
| HR      | `hr.rtylr.com`         |
| CRM     | `crm.rtylr.com`        |
| Finance | `finance.rtylr.com`    |
| Flow    | `flow.rtylr.com`       |
| Landing | `rtylr.com`            |

All `*.rtylr.com` subdomains should be CNAME'd to the Cloudflare Pages URL (auto-managed by Cloudflare if the domain is on the same account).

---

## Cookie Auth (Cross-subdomain)

Auth is cookie-based. The token is set on `.rtylr.com` so all subdomains share it:

```
setCookie("token", jwt, { domain: ".rtylr.com", path: "/", secure: true, sameSite: "Lax" })
```

The Cloudflare Pages apps don't need any special headers for this — it works out of the box since all apps are on `*.rtylr.com`.

---

## Build from Monorepo Root

If building from the monorepo root with Turborepo:

```bash
# Build all apps
pnpm turbo build

# Build a specific app
pnpm turbo build --filter=@rtylr/erp

# Dev mode
pnpm turbo dev --filter=@rtylr/crm
```

Set the **root directory** in Cloudflare Pages to `rtylr` (the monorepo root) and use build commands like:

```bash
cd apps/erp && pnpm build
```

Or rely on Turborepo caching by setting build command to:

```bash
pnpm turbo build --filter=@rtylr/erp
```

---

## API Client (all apps)

Each app has `src/lib/api.ts` — an axios client that reads the `token` cookie and attaches it as a Bearer token:

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'https://api.voxire.dev/',
});

api.interceptors.request.use((config) => {
  const token = document.cookie.split(';')
    .map(c => c.trim())
    .find(c => c.startsWith('token='))
    ?.split('=')[1];
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
```

---

## New Apps Added (this session)

The following apps were newly wired up to the API:

| App     | Services file                       | Key endpoints                                  |
|---------|-------------------------------------|------------------------------------------------|
| ERP     | `src/services/erpService.ts`        | products, inventory, suppliers, purchase orders |
| HR      | `src/services/hrService.ts`         | staff, departments, shifts, payroll, time off   |
| CRM     | `src/services/crmService.ts`        | customers, loyalty, campaigns, segments         |
| Finance | `src/services/financeService.ts`    | invoices, expenses, tax, P&L                    |
| Flow    | `src/services/flowService.ts`       | workspaces, projects, tasks, labels             |
| vbitez  | `src/services/posService.ts`        | tables, orders, sessions, locations (POS)       |
