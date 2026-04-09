# Frontend Configuration

All frontend apps live in `rtylr/rtylr/apps/`. They are React + Vite SPAs (except `landing/` which is Next.js 15), deployed to Cloudflare Pages on individual subdomains.

---

## App Overview

| App       | Package Name     | Subdomain              | Local Port | Framework    |
|-----------|-----------------|------------------------|------------|--------------|
| Hub       | `hub`           | `hub.rtylr.com`        | 5175       | React + Vite |
| Auth      | `@rtylr/auth`   | `auth.rtylr.com`       | 5176       | React + Vite |
| Menu      | `@rtylr/menu`   | `menu.rtylr.com`       | 5177       | React + Vite |
| POS       | `@rtylr/pos`    | `pos.rtylr.com`        | 5178       | React + Vite |
| ERP       | `@rtylr/erp`    | `erp.rtylr.com`        | 5179       | React + Vite |
| HR        | `@rtylr/hr`     | `hr.rtylr.com`         | 5180       | React + Vite |
| CRM       | `@rtylr/crm`    | `crm.rtylr.com`        | 5181       | React + Vite |
| Finance   | `@rtylr/finance`| `finance.rtylr.com`    | 5182       | React + Vite |
| Flow      | `@rtylr/flow`   | `flow.rtylr.com`       | 5183       | React + Vite |
| Landing   | `@rtylr/landing`| `rtylr.com`            | 4200       | Next.js 15   |

---

## Cloudflare Pages Configuration (per app)

Each app is its own Cloudflare Pages project.

### Common Settings (all Vite apps)

| Setting              | Value               |
|----------------------|---------------------|
| Framework preset     | None (custom)       |
| Build command        | `pnpm build`        |
| Build output dir     | `dist`              |
| Root directory       | `rtylr/apps/{app}`  |
| Node version         | `20`                |

### Landing (Next.js)

| Setting              | Value                        |
|----------------------|------------------------------|
| Framework preset     | Next.js                      |
| Build command        | `pnpm build`                 |
| Build output dir     | `.next`                      |
| Root directory       | `rtylr/apps/landing`         |
| Node version         | `20`                         |

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
| POS     | `rtylr-pos`         |
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
| POS     | `pos.rtylr.com`        |
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
| POS     | `src/services/posService.ts`        | tables, orders, sessions, locations             |
