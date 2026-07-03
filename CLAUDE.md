# CLAUDE.md - rtylr Frontend

## 1. PROJECT IDENTITY

- **Product:** RTYLR - an all-in-one commerce operating system ("Buy rtylr, buy a whole business powered by AI.")
- **Repo role:** Frontend monorepo. Contains every customer-facing and operator-facing web app, plus all shared packages.
- **Built by:** Voxire - Abed El-Fattah Amouneh & Mohammad Homsi
- **Business email:** info@voxire.com
- **WhatsApp:** +961 3 940 708

---

### Definition of Done

A task is only complete when ALL of the following are true:

- `pnpm typecheck` passes with zero errors
- `pnpm lint` passes with zero errors
- The code follows every pattern documented in this file
- New user-facing strings are in locale files (not hardcoded)
- Mobile-first compliance checked (min-width only, 44px touch targets, 16px font on inputs)
- No `console.log` or commented-out code in the diff
- If a new architectural pattern was introduced, it is documented in `docs/` and in this CLAUDE.md

### PR Self-Review Checklist

Before opening a pull request, run through this list personally:

- [ ] `pnpm typecheck` and `pnpm lint` pass with zero errors
- [ ] No `.env`, `.env.local`, or secret values in the diff
- [ ] No accidental `console.log`, `debugger`, or commented-out code
- [ ] `pnpm-lock.yaml` changes are intentional (new dep added, not corrupted)
- [ ] Branch name follows the `feat/`, `fix/`, `chore/` prefix convention
- [ ] All new UI strings have i18n keys in all three locale files
- [ ] New components follow the `@voxire/ui` → local decision rule
- [ ] Diff reviewed line-by-line in the browser before assigning reviewers

### When You Introduce a New Pattern

If you create a new convention that other developers will need to follow (a new hook shape, a new component structure, a new state management approach, etc.), you **must** document it in this CLAUDE.md and in `rtylr/docs/` before the PR is mergeable. Don't let undocumented patterns accumulate.

### Updating Documentation

**Always update `rtylr/docs/` when you make architectural or pattern-level changes.** This includes adding a new shared package, changing the API layer, changing the auth flow, adding a new app, or changing how routing works. The `docs/` folder is the authoritative reference for the team - keep it current.

---

### Adding Dependencies

**Before adding any new npm package**, check whether an existing workspace package or already-installed dependency already covers the need. Check:

1. `packages/ui`, `packages/utils`, `packages/components` - shared primitives
2. The app's own `package.json` devDependencies and dependencies
3. The root `package.json` shared devDependencies

If a package already exists that does the job (even partially), use it. Don't install a second library for the same purpose.

### Bundle Size Discipline

Before adding any new library, check its size on [bundlephobia.com](https://bundlephobia.com). Prefer tree-shakeable packages. Never import entire namespaces:

```ts
// ❌ imports the entire lodash bundle (~72 kB gzipped)
import _ from 'lodash';

// ✅ tree-shakeable - only debounce is bundled
import debounce from 'lodash/debounce';
```

Run `pnpm --filter <app> build` and inspect the Vite output chunk sizes after adding a heavy dependency. If any chunk exceeds 500 kB gzipped, investigate and split it before merging.

---

## 4. ARCHITECTURE & PATTERNS

### Monorepo Layout

```
rtylr/
├── apps/
│   ├── dash/         Port 5175 - Command center (menu, orders, analytics, billing, settings, team)
│   ├── auth/         Port 5176 - Cross-app SSO login/register/oauth consent
│   ├── menu/         Port 5177 - Customer-facing digital menu (public, no auth)
│   ├── erp/          Port 5179 - Inventory & purchasing
│   ├── hr/           Port 5180 - Staff, shifts, payroll
│   ├── crm/          Port 5181 - Customers, loyalty, campaigns, sales
│   ├── finance/      Port 5182 - Invoices, expenses, P&L
│   ├── flow/         Port 5183 - Projects & tasks
│   ├── recruit/      Port 5184 - Careers portal (public apply pages + admin)
│   ├── insights/     Port 5187 - Analytics deep-dive
│   └── docs/         Internal docs viewer app
├── packages/         GIT SUBMODULE (voxire/packages) - all shared workspace packages
│   ├── api/          @voxire/api      - Axios client, safeFetch, ResultAsync, realtime
│   ├── auth/         @voxire/auth     - AppGuard, ProtectedRoute, RoleGuard, useUserStore, cookies, signOut
│   ├── caret/        @voxire/caret    - AI assistant (Caret) shared features
│   ├── components/   @voxire/components - Higher-level shared components (FullScreenLoader, etc.)
│   ├── config/       @voxire/config   - APP_URLS, APP_CONFIG, AppFlagGuard, industries, themes
│   ├── consent/      @voxire/consent  - Cookie/consent management
│   ├── contracts/    @voxire/contracts - Re-exports gen/ts from packages/contracts/voxie (nested submodule)
│   ├── i18n/         @voxire/i18n     - Shared i18next setup
│   ├── observability/ @voxire/observability - Client-side error/log reporting
│   ├── permissions/  @voxire/permissions - RBAC helpers
│   ├── types/        @voxire/types    - Shared TypeScript interfaces (no runtime)
│   ├── ui/           @voxire/ui       - Shared low-level components (incl. ErrorBoundary)
│   ├── assets/, menu/, tailwind-config/, utils/   - as named
├── docs/             Architecture docs + ADRs (authoritative)
├── e2e/              Playwright e2e tests (one spec per app area)
├── test/             Shared test utilities
├── scripts/          audit-request-contracts.mjs - audits request shapes against voxie contracts
├── eslint.config.mjs Flat ESLint config (root)
├── pnpm-workspace.yaml
├── turbo.json
├── tsconfig.base.json
└── package.json      Root - shared devDependencies and scripts
```

> **Removed apps:** `landing` (marketing site) now lives in its own repo. `hub` was renamed to `dash`. `vbitez` (POS desktop) is not part of this monorepo. Empty `apps/hub`, `apps/vbitez`, or `apps/landing` folders on disk are untracked leftovers - safe to delete.
>
> **Submodules:** `packages/` (tryrtylr/packages) is the only top-level submodule here. It carries its own nested `voxie` submodule at `packages/contracts/voxie` (proto contracts + generated `gen/ts`, which `@voxire/contracts` re-exports directly) — there is no separate top-level `voxie/` anymore. Run `git submodule update --init --recursive` after every pull, then `pnpm install`. `pos/` (the separate POS terminal repo) shares this same `packages` submodule and gets contracts the same way.

### Tech Stack

| Component    | Technology                          | Version                                         |
| ------------ | ----------------------------------- | ----------------------------------------------- |
| Runtime / PM | Node 20+ / pnpm                     | 10.33.0 (`packageManager` in root package.json) |
| Framework    | React                               | 19                                              |
| Routing      | @tanstack/react-router              | v1 (react-router has been fully removed)        |
| Server state | @tanstack/react-query               | v5                                              |
| UI state     | Zustand                             | v5                                              |
| Build        | Vite + Turborepo                    | Vite 8                                          |
| Language     | TypeScript                          | 6.x                                             |
| Lint         | ESLint flat config                  | `eslint.config.mjs` at root                     |
| Tests        | Vitest (unit) + Playwright (`e2e/`) | —                                               |

### App Structure (Vite apps)

Each Vite app follows this internal structure:

```
apps/<name>/src/
├── features/         Feature modules - the most important directory
│   └── <feature>/
│       ├── <Feature>.tsx         Main screen component
│       ├── <SubScreen>.tsx       Additional screens
│       ├── use<Feature>.ts       Feature-specific hooks
│       └── components/           Local-only components (not shared)
├── shared/
│   ├── components/   Shared components within this app only
│   ├── hooks/
│   │   └── queries/  TanStack Query hooks (useX.ts per concern)
│   ├── stores/       Zustand stores
│   ├── schemas/      Zod validation schemas
│   ├── services/     (legacy) direct API call wrappers
│   ├── lib/
│   │   └── api.ts    App-specific `safeFetch` instance
│   └── utils/        App-local utilities
├── router.tsx        TanStack Router route tree (createRootRoute/createRoute/createRouter)
├── layouts/          Layout wrappers (DefaultLayout, etc.)
├── locales/          Namespaced i18n JSON: locales/{en,ar,fr}/<namespace>.json (hr still uses flat en.json/ar.json/fr.json)
├── App.tsx           Mostly legacy - routing moved to router.tsx
└── main.tsx          Entry point - React root, QueryClient, RouterProvider
```

> Some apps still carry a legacy `routes/AppRoutes.tsx` file - `src/router.tsx` is the live route tree everywhere. Don't add routes to `AppRoutes.tsx`.

### Naming Conventions

| Item            | Convention                  | Example                                      |
| --------------- | --------------------------- | -------------------------------------------- |
| Component files | PascalCase                  | `TeamMemberCard.tsx`                         |
| Hook files      | camelCase, `use` prefix     | `useTeam.ts`                                 |
| Store files     | camelCase, `Store` suffix   | `userStore.ts`                               |
| Service files   | camelCase, `Service` suffix | `teamService.ts`                             |
| Schema files    | camelCase, `schema` suffix  | `inviteSchema.ts`                            |
| Type-only files | camelCase                   | `types.ts`                                   |
| Route tree      | `router.tsx` (per app)      | `apps/erp/src/router.tsx`                    |
| Feature dirs    | camelCase                   | `features/qr-codes/`                         |
| Query hooks     | `use` + noun + verb or noun | `useGetConfig`, `useTeam`, `useUpdateConfig` |

### Component Pattern

Components are functional, always explicitly typed with `React.JSX.Element` or `React.ReactNode` return types. Props interfaces are colocated. Example:

```tsx
interface TeamCardProps {
  member: StaffUserPayload;
  onEdit: (id: number) => void;
}

const TeamCard = ({ member, onEdit }: TeamCardProps): React.JSX.Element => {
  // ...
};

export default TeamCard;
```

### Component Placement: @voxire/ui vs Local

Use this decision rule when creating a new component:

- **Use `@voxire/ui`** if a suitable primitive already exists (Button, Modal, Card, Input, Badge, etc.)
- **Use `@voxire/components`** if a higher-level shared component already covers the use case
- **Create a local `features/<name>/components/` component** if the UI is genuinely specific to one feature and will not be reused across apps
- **Create a local `shared/components/` component** if the UI is reused within one app but not across apps
- **Propose adding to `@voxire/ui` or `@voxire/components`** if you're building something that other apps will likely need

Don't duplicate a component that already exists in a shared package. Read the shared packages before building new UI.

### Lazy Loading - REQUIRED

**All route-level components must be lazy-loaded.** This keeps the initial bundle small and only loads code when the user navigates to that route.

```tsx
// ✅ correct - lazy load every route component (in router.tsx)
const Team = React.lazy(() => import('./features/team/Team'));
const TeamMember = React.lazy(() => import('./features/team/TeamMember'));

// wrap() puts each lazy component behind a Suspense fallback
function wrap(component: React.ReactNode) {
  return () => <Suspense fallback={<FullScreenLoader />}>{component}</Suspense>;
}

const teamRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/team',
  component: wrap(<Team />),
});

// ❌ wrong - eager import bloats the initial bundle
import Team from './features/team/Team';
```

Never eagerly import a page-level component. Layout components (`DefaultLayout`, etc.) are the exception - they are always needed and can be eagerly imported.

### Routing Pattern - TanStack Router

Every app defines its route tree in `src/router.tsx` using `@tanstack/react-router` (`createRootRoute`, `createRoute`, `createRouter`) and mounts it with `<RouterProvider router={router} />` in `main.tsx`. The root route composes the guards:

```tsx
const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="erp"
        fetchUser={(): Promise<Record<string, unknown>> =>
          unwrap(safe.get<Record<string, unknown>>('/user/me'))
        }
      >
        <Outlet />
      </AppGuard>
    </ErrorBoundary>
  ),
});
```

`AppGuard` (from `@voxire/auth`) replaces the old per-app `ProtectedRoute` wiring: it checks the cookie, loads the user, and enforces `allowedRoles` from `APP_CONFIG`. Feature-flagged apps additionally wrap with `AppFlagGuard` from `@voxire/config`. Unknown paths render a `CatchAllRedirect` that sends the user back to `/`.

### Error Boundary Placement

Every top-level route group must be wrapped in an `<ErrorBoundary>` (exported from `@voxire/ui`). This prevents a crash in one feature from taking down the entire app.

```tsx
// Wrap the route tree - one boundary per app at minimum (in the root route component)
<ErrorBoundary>
  <AppGuard appId="hr" fetchUser={...}>
    <Outlet />
  </AppGuard>
</ErrorBoundary>
```

Add a second, tighter boundary around any feature that loads third-party integrations or performs complex data transformations (e.g., the analytics chart section, the rich-text editor).

**Dashboard widgets must each have their own `<ErrorBoundary>`.** A shared boundary around all widgets means one broken widget crashes the entire dashboard. Wrap every widget individually:

```tsx
// ✅ correct - fault-isolated widgets
<ErrorBoundary fallback={<WidgetError label={t('dashboard.widgets.errorFallback')} />}>
  <HeadcountWidget />
</ErrorBoundary>
<ErrorBoundary fallback={<WidgetError label={t('dashboard.widgets.errorFallback')} />}>
  <PayrollCostWidget />
</ErrorBoundary>

// ❌ wrong - one crash kills everything
<ErrorBoundary fallback={<DashboardError />}>
  <HeadcountWidget />
  <PayrollCostWidget />
</ErrorBoundary>
```

`ErrorBoundary` is exported from `@voxire/ui` - import it from there. Do not create per-app copies.

### State Management Pattern

- **Zustand** stores hold UI state, user session, and preferences. Each store has a clear `State` interface and an `Actions` interface. Stores are persisted via `zustand/middleware persist` when state must survive page refresh.
- **TanStack Query** handles all server state. Query hooks live in `shared/hooks/queries/`. Mutations use `useMutation` with `onSuccess`/`onError` toast feedback. Data from queries is NOT duplicated into Zustand.
- **Do NOT** put API response data into Zustand manually. Exception: config/theme data that is loaded once and needed app-wide (e.g., `configStore`, `themeStore`).

### Query Key Naming Conventions

Consistent query keys are required for correct cache invalidation. Follow these rules:

```ts
// List (no filters): single-element array
queryKey: ['team'];
queryKey: ['products'];

// Single item by ID: two-element array
queryKey: ['team', userId];
queryKey: ['product', productId];

// Filtered or parameterized list: three-element array
queryKey: ['team', 'filter', { role, status }];
queryKey: ['orders', 'date', { from, to }];

// Mutation-triggered refetch: use the list key
queryClient.invalidateQueries({ queryKey: ['team'] });
```

Never use inline objects as the first element. Never use strings with embedded IDs (e.g., `'team-123'`).

### Cache Invalidation Pattern

On every successful mutation, invalidate the relevant query keys immediately:

```ts
const mutation = useMutation({
  mutationFn: (data: InviteUserRequest) => unwrap(safe.post('/auth/invite', data)),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['team'] });
    toast.success(t('team.inviteSent'));
  },
  onError: () => {
    toast.error(t('errors.generic'));
  },
});
```

Never rely on stale data after a mutation. Always invalidate.

### API Layer Pattern

The canonical API call chain is:

```
TanStack Query hook (useX.ts)
  → unwrap(safe.get/post/put/delete(...))   ← safe = safeFetch(api)
  → ResultAsync<T, ApiError>
  → unwrap() converts to Promise<T> for React Query
```

- `safe` is the `SafeApi` instance created via `safeFetch(createApiClient('appId'))` in `src/shared/lib/api.ts`.
- Keys are automatically converted: request body camelCase → snake_case (outgoing), response snake_case → camelCase (incoming). This is handled by interceptors in `@voxire/api`. **Never manually convert keys.**
- On 401: cookie is cleared, user is redirected to `auth.rtylr.com?redirect=<appId>`.

```ts
// src/shared/lib/api.ts
import { createApiClient, safeFetch } from '@voxire/api';
export const api = createApiClient('dash'); // swap appId per app
export const safe = safeFetch(api);
```

```ts
// src/shared/hooks/queries/useTeam.ts
import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';

export const useListTeam = () =>
  useQuery({
    queryKey: ['team'],
    queryFn: () => unwrap(safe.get<ListUsersResponse>('/team/users')),
  });
```

### Form Handling & Validation

- Zod schemas live in `shared/schemas/`.
- Forms use controlled inputs or react-hook-form where complexity warrants.
- All external API response data must be parsed through a Zod schema before use. Example:

```ts
const configSchema = z.object({ maintanance: z.boolean(), ... });
const parsed = configSchema.parse(response.data.config);
```

### Internationalisation (i18n) - NON-NEGOTIABLE

**All user-facing strings must go through i18next.** No hardcoded English text in JSX or component logic.

```tsx
// ✅ correct
const { t } = useTranslation();
return <button>{t('team.inviteUser')}</button>;

// ❌ wrong - hardcoded string
return <button>Invite User</button>;
```

Locale files are namespaced per app: `locales/{en,ar,fr}/<namespace>.json` (e.g. `locales/en/erp.json`, `locales/en/common.json`). The HR app still uses the legacy flat `locales/en.json` format. When adding a new string:

1. Add the key to the relevant `locales/en/<namespace>.json` first.
2. Add the same key to the matching `ar/` and `fr/` namespace files.
3. Use the key in JSX via `t('...')`.

Never ship a new UI element without a translation key, even in early development.

### Authentication Pattern

- `AppGuard` from `@voxire/auth` wraps each app's root route (see Routing Pattern above): checks the cookie, fetches `/user/me`, and enforces the app's `allowedRoles` from `APP_CONFIG`.
- `ProtectedRoute` and `RoleGuard` from `@voxire/auth` remain available for finer-grained gating of routes and UI elements.
- `useUserStore` from `@voxire/auth` is the single source of truth for user identity; `signOut` from the same package handles logout.
- **Never read the token from a store or prop.** Always read it via the cookie mechanism in `@voxire/api` or `@voxire/auth`.

### BusinessRoute (Dash)

Dash wraps its route tree in a `<BusinessRoute>` component (`apps/dash/src/shared/utils/BusinessRoute.tsx`, used in `router.tsx`). This component:

- Ensures the business config (`/get-configs`) is loaded and stored in `configStore` before rendering any dash route.
- Applies the business's theme CSS variables to the document root.
- Redirects to the onboarding flow if the business has not completed setup.

**Do not remove or bypass `BusinessRoute`.** If you add new dash routes, they must be inside it.

### Contracts / Types Pattern

- Prefer types from `@voxire/contracts` over hand-written interfaces when a proto definition exists.
- Import paths mirror the voxie proto layout (`{app}/{capability}/v1/{operation}` plus a `types` module per capability):
  `import type { StaffUserPayload } from '@voxire/contracts/hub/team/members/v1/types'`
  `import type { ClockInRequest } from '@voxire/contracts/hr/attendance/v1/attendance_clock_in'`
- `@voxire/types` contains non-proto types (legacy and shared domain types).
- Never copy-paste type shapes from contracts. Import them.
- Run `pnpm contracts:audit-requests` to verify request shapes used in apps still match the contracts.

---

## 5. ROUTES / ENDPOINTS

> The marketing site (landing) lives in its own repository - it is no longer part of this monorepo.

### Dash (`apps/dash/`, port 5175)

| Route                    | Screen                |
| ------------------------ | --------------------- |
| `/`                      | Dashboard             |
| `/menus`                 | Menu list             |
| `/menu`                  | Menu editor           |
| `/menu/:id`              | Specific menu editor  |
| `/create-item`           | Create menu item      |
| `/create-banner`         | Create menu banner    |
| `/orders`                | Orders                |
| `/qr-codes`              | QR code list          |
| `/qr-codes/create`       | QR code generator     |
| `/design`                | Theme/branding editor |
| `/analytics`             | Analytics             |
| `/chat`                  | Chat                  |
| `/team`                  | Team list             |
| `/team/invite`           | Invite user           |
| `/team/:id`              | Member detail         |
| `/permissions`           | Permissions           |
| `/roles`                 | Roles                 |
| `/apps`                  | App access            |
| `/activity`              | Activity log          |
| `/audit`                 | Audit trail           |
| `/notifications`         | Notifications         |
| `/api-keys`              | Developer API keys    |
| `/webhooks`              | Webhooks              |
| `/security`              | Security settings     |
| `/theme`                 | Theme customizer      |
| `/help`                  | Help center           |
| `/settings/account`      | Account profile       |
| `/settings/general`      | Business settings     |
| `/settings/industry`     | Business type         |
| `/settings/domain`       | Custom domain         |
| `/settings/email`        | Email settings        |
| `/settings/billing`      | Billing               |
| `/settings/addons`       | Add-ons               |
| `/settings/integrations` | Integrations          |

### Other Apps

| App             | Key Routes                                                                                                                                                                                     |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Auth (5176)     | `/` (login), `/register`, `/reset-password`, `/email-verification`, `/oauth` (consent), `/callback`, `/invite`                                                                                 |
| Menu (5177)     | public digital menu + cart                                                                                                                                                                     |
| ERP (5179)      | `/`, `/inventory`, `/purchasing`, `/suppliers`, `/products`, `/warehouses`, `/stock-transfers`, `/stock-takes`, `/price-lists`, `/variants`, `/barcodes`, `/expiry`, `/valuation`, `/settings` |
| HR (5180)       | `/`, `/staff`, `/employees`, `/shifts`, `/payroll`, `/departments`, `/leave`, `/attendance`, `/timesheets`, `/performance`, `/onboarding`, `/setup-wizard`                                     |
| CRM (5181)      | `/`, `/customers`, `/loyalty`, `/campaigns`, `/segments`, `/sales` (leads, contracts), `/templates`, `/automations`                                                                            |
| Finance (5182)  | `/`, `/invoices`, `/expenses`, `/bills`, `/quotes`, `/journal`, `/reconcile`, `/profit-loss`, `/balance-sheet`, `/cash-flow`, `/tax`                                                           |
| Flow (5183)     | `/`, `/w/:workspaceId`, `/w/:workspaceId/p/:projectId`, `/inbox`, `/triage`, `/roadmap`, `/settings`                                                                                           |
| Recruit (5184)  | public apply pages + `/admin` (careers, submissions, form fields)                                                                                                                              |
| Insights (5187) | `/`, `/revenue`, `/customers`, `/products`, `/financial`, `/operations`, `/reports`, `/settings`                                                                                               |

Each app's `src/router.tsx` is the definitive route list.

---

## 6. DESIGN SYSTEM & UI CONVENTIONS

### Tailwind Configuration

All apps extend `@voxire/tailwind-config` (shared preset). Dash overrides `primary` and `admin` color tokens to use CSS custom properties:

```ts
primary: { DEFAULT: 'var(--hub-text)', foreground: 'var(--hub-text)' }
admin: { DEFAULT: '#262626', bg: 'var(--hub-bg)', border: 'var(--hub-border)' }
```

> The `--hub-*` CSS variable names are legacy (pre-rename); `globals.css` also defines newer `--dash-*` tokens. Don't introduce new `--hub-*` variables.

CSS variables are set at runtime based on the business's chosen theme (loaded from the `themes` table).

### Mobile-First - NON-NEGOTIABLE

- Write CSS mobile-first: breakpoints use `min-width` only. **No `max-width` queries on new code.**
- All `<input>`, `<select>`, `<textarea>` elements must have `font-size: 16px` minimum - prevents iOS auto-zoom.
- All interactive touch targets must be ≥ 44×44px.

### Accessibility (a11y) Requirements

Every interactive element must be keyboard-navigable and screen-reader accessible:

- All `<button>` and `<a>` elements must have a visible label or `aria-label`. Never use an icon-only button without `aria-label`.
- Modals must trap focus when open and restore focus to the trigger when closed. Use the Radix `Dialog` primitive - it handles this automatically.
- Color contrast must meet WCAG 2.1 AA: 4.5:1 for normal text, 3:1 for large text and UI components.
- Form inputs must have associated `<label>` elements (or `aria-label`). Never rely on `placeholder` as the only label - placeholders disappear on input.
- Images that convey information need descriptive `alt` text. Decorative images use `alt=""`.
- Loading states must announce themselves to screen readers: use `aria-live="polite"` on the container that updates.

Do not ship a new screen or component without verifying it is navigable by keyboard (Tab, Shift+Tab, Enter, Escape) and that the DOM order matches the visual order.

### Component Library

- `@voxire/ui` for low-level primitives (Button, Modal, Card, Input, etc.)
- `@radix-ui/react-dialog` for dialogs in Dash
- `lucide-react` and `phosphor-react` for icons
- Skeleton loaders via `react-loading-skeleton`

---

## 7. CODE QUALITY RULES - NON-NEGOTIABLE

1. **TypeScript strict mode is ON** (`"strict": true` in all tsconfigs). Never use `any`, `@ts-ignore`, or unsafe type assertions. If unavoidable, add a comment explaining why.
2. **No commented-out code** in commits. Delete dead code.
3. **No `console.log`** in committed code. Use structured error reporting or remove.
4. **Every function must have an explicit return type.** `const foo = (): string => { ... }` not `const foo = () => { ... }`.
5. **Zod validation** on all external data boundaries: API responses, form inputs, env vars, localStorage reads.
6. **Never bypass the `safeFetch` / `ResultAsync` pattern** for API calls. No raw `axios.get(...)` in components or hooks - always go through `safe.get/post/...`.
7. **No direct DOM manipulation** outside of custom hooks.
8. **No inline styles** except for dynamic CSS variable values that cannot be expressed in Tailwind.
9. **No hardcoded hex color literals anywhere in component code.** Every color must reference a CSS custom property (`var(--hr-status-approved)`, `var(--hr-error)`, etc.) defined in `globals.css`. Never write `#10b981`, `#ef4444`, `color: '#8b5cf6'`, or any hex string inline. For opacity variants where the Pill component uses string concatenation, use `color-mix(in srgb, var(--hr-xxx) 30%, transparent)` via the `style` prop. If a new semantic color is needed, add it to `:root` in `globals.css` first, then reference it.
10. **Imports must be from workspace packages** (`@voxire/*`) not via relative paths crossing package boundaries.
11. **Do not mix Zustand and TanStack Query responsibilities.** Server state → Query. UI state → Zustand.
12. **`StoredUser` dynamic key access is a known type hole.** Never add new user fields via the `[key: string]: unknown` index signature. Declare them as explicit typed properties on the interface.
13. **All user-facing strings must use i18next `t()`.** No hardcoded English text in JSX.

---

## 8. CONFIGURABILITY RULES

RTYLR is a multi-tenant SaaS product sold to restaurants, agencies, law firms, retail groups, and every other type of business. **This shapes every UI decision.** A restaurant and a law firm have different workflows, different terminology, different policies - but both subscribe to the same app. Every feature the frontend renders must be driven by the company's configuration, not by assumptions we bake into the code.

### The developer checklist

Before hardcoding any behavior, ask both of these questions. If either answer is yes, the value or behavior must come from company config, not from a constant or conditional in the component:

1. **"Would a restaurant and a law firm want this to behave differently?"** - If two types of businesses would expect different behavior, it is not a product rule. It is a configuration.
2. **"Is this a business rule or a product rule?"** - Product rules (routing, auth flow, pagination) are fixed. Business rules (role names, approval flows, visible features, leave types, notification preferences) are always configurable.

If you can answer "no" to both with confidence, hardcoding is acceptable - but add a comment explaining the decision so the next developer understands it was intentional.

### Things that must never be hardcoded - ever

```tsx
// ❌ NEVER hardcode a role name in permission logic
if (user.role === 'manager') {
  return <ApproveButton />;
}
// ✅ Check against the permission map loaded from company config
if (hasPermission(hrConfig, user, 'approve_timesheets')) {
  return <ApproveButton />;
}

// ❌ NEVER hardcode a feature's visibility
{user.role === 'hr_admin' && <PayrollTab />}
// ✅ Read from the company's enabled features / role config
{hasPermission(hrConfig, user, 'view_payroll') && <PayrollTab />}

// ❌ NEVER hardcode approval flow labels or step counts
<p>Waiting for manager approval</p>
// ✅ Read from the approval chain config the company set up
<p>{t('approvals.waitingFor', { approver: currentStep.approverLabel })}</p>

// ❌ NEVER hardcode a leave type name
<Badge>Sick Leave</Badge>
// ✅ Render from the company's configured leave types
<Badge>{leaveType.name}</Badge>

// ❌ NEVER hardcode notification channel assumptions
toast.success('Email sent to manager');
// ✅ The backend dispatches through configured channels; the frontend just confirms
toast.success(t('notifications.requestSent'));
```

### Company config drives the UI

Every RTYLR app has a config guard component (e.g. `BusinessRoute` in Dash, the HR config store/guard in the HR app) that:

1. Loads the company's configuration before rendering any feature screens.
2. Stores the config in a Zustand store with `persist` so it survives page refresh.
3. Redirects to the setup wizard if config is incomplete.

Feature components read from this store to decide what to show - they never make their own assumptions about what the company has configured.

### Fallback behavior when config is missing

A company may not have finished their setup. The UI must handle this gracefully:

- **Non-critical features** (dashboard widgets, display preferences, optional analytics): render with a sensible default or show an empty state. Do not block the user or show an error.
- **Critical features** (payroll run, timesheet submission, leave requests that require a configured policy): show a clear, actionable prompt - "Set up your payroll configuration first" with a link to Settings. Do not allow the operation to proceed silently with assumed values when money or compliance is involved.

```tsx
// Non-critical: empty state, not an error
const PayrollCostWidget = (): React.JSX.Element => {
  const { config } = useHRConfigStore();
  if (!config?.payrollEnabled) {
    return <WidgetEmptyState message={t('dashboard.payrollNotConfigured')} />;
  }
  // ... render chart
};

// Critical: block with guidance
const RunPayrollButton = (): React.JSX.Element => {
  const { config } = useHRConfigStore();
  if (!config?.payrollConfigured) {
    return (
      <ConfigurationRequired
        message={t('payroll.configRequired')}
        settingsPath="/settings/payroll-config"
      />
    );
  }
  return <Button onClick={runPayroll}>{t('payroll.runPayroll')}</Button>;
};
```

### Scope and the boy scout rule

- **All new code** must follow these rules without exception.
- **Existing code**: when you open a file to make any change, fix any hardcoded business rule you encounter in that file before committing. Leave the code more configurable than you found it. You are not required to audit the entire codebase - only the files you touch.
- This is a **strong guideline**. If hardcoding is genuinely right (a true product constant that will never differ between customers), document the reason in a comment. Never use this exception for business rules.

---

## 9. SECURITY RULES

Never commit secrets. Never expose sensitive data to the browser. Follow these rules without exception:

- **No secrets in code.** API keys, JWT secrets, DB passwords, and tokens must live in `.env` files and never be committed. Use `VITE_` prefix only for values that are intentionally public (they are inlined into the bundle).
- **Never trust `localStorage` for auth.** The JWT token is stored in an `httpOnly` cookie managed by the backend. Never store tokens in `localStorage`, `sessionStorage`, or Zustand.
- **Validate on the server, not just the client.** Frontend validation (Zod, form rules) is for UX only - the backend must independently validate all inputs. Never skip server-side validation because the frontend already checked.
- **No `dangerouslySetInnerHTML`** with user-supplied content. This is an XSS vector. If rich text rendering is required, use a sanitised rendering library (e.g., DOMPurify) and document the decision.
- **Never log PII** (email, phone, user IDs) to the browser console in committed code.
- **Dependency audits:** Run `pnpm audit` periodically and before major releases. Address high/critical vulnerabilities before shipping.
- **No open redirects.** Any redirect that includes a URL from query params (e.g., `?redirect=`) must validate the destination is on an allowed domain before redirecting.

---

## 9. PERFORMANCE RULES

- **All list queries must be paginated.** Never fetch an unbounded list from the API. Every `useQuery` that fetches a collection must pass `limit` and `offset` (or a cursor). Display loading states while fetching subsequent pages.
- **Avoid re-renders from unstable references.** Callbacks passed as props should be wrapped in `useCallback`. Derived values should use `useMemo` when the computation is expensive. Do not create new objects or arrays inline in JSX if they are passed as props to memoised children.
- **Skeleton loaders over spinners.** Use `react-loading-skeleton` for content-area loading states so the layout does not shift when data arrives.
- **Images must specify `width` and `height`** to prevent layout shift (CLS). For dynamic images, use an aspect-ratio container.
- **Virtualise long lists.** Any list that can exceed ~100 items must use a virtualisation library (e.g., `@tanstack/react-virtual`) rather than rendering all rows to the DOM.

---

## 10. KNOWN ISSUES

| Issue                                                                                                                                                          | File / Location                                                          | Priority |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | -------- |
| `tsconfig.base.json` lacks `strict: true` at root level - individual app tsconfigs must set it                                                                 | `tsconfig.base.json`                                                     | High     |
| `useUserStore` in Dash has `[key: string]: unknown` index signature on `StoredUser`; `@voxire/auth` `User` type has `[key: string]: any` - weakens type safety | `apps/dash/src/shared/stores/userStore.ts`, `packages/auth/src/types.ts` | Medium   |
| Legacy `routes/AppRoutes.tsx` files still exist in some apps alongside the live `router.tsx`; `MenuWithId.jsx` is a mixed-extension import in dash             | `apps/dash/src/routes/`, `apps/dash/src/features/menu/`                  | Low      |
| HR app still uses flat locale files (`locales/en.json`) instead of the namespaced `locales/en/*.json` layout                                                   | `apps/hr/src/locales/`                                                   | Low      |
| Empty untracked `apps/hub/`, `apps/vbitez/`, `apps/landing/` leftover folders may exist on local checkouts                                                     | `apps/`                                                                  | Low      |
| `@voxire/contracts/package.json` exports `book/v1/*`, `store/v1/*`, and `auth/v1/auth` paths with no proto sources in voxie                                    | `packages/contracts/package.json`                                        | Medium   |
| No `.env.example` visible in repo root - environment setup requires reading docs                                                                               | Root                                                                     | Medium   |

---

## 11. SESSION STARTUP CHECKLIST

Run these steps at the start of every new session before touching any file:

1. `git pull origin main` (or the relevant branch)
2. `git submodule update --init --recursive` (keeps `packages/` and its nested `voxie` contracts submodule in sync)
3. `git checkout -b feat/<task-name>` to create a feature branch
4. `pnpm install` if `pnpm-lock.yaml` changed since last session
5. Check `docs/` for any new files added since last session
6. Re-read the relevant sections of this CLAUDE.md
7. Confirm the dev server starts: `pnpm --filter dash dev` (or the relevant app)

---

## 12. INCIDENT RESPONSE

### Frontend is down / showing a white screen

1. Check Cloudflare Pages dashboard → [FILL IN: your Cloudflare Pages project URL] → Deployments tab. Look for a failed or rolled-back deployment.
2. Roll back to the last known good deployment: Cloudflare Pages → Deployments → find the last green deploy → "Retry deployment" or pin it as the active deployment.
3. Check the browser console for JavaScript errors. A white screen in production is almost always a runtime error in a component that lacks an `<ErrorBoundary>`.
4. Check `pnpm typecheck` and `pnpm build` locally on the `main` branch - if either fails, the deploy pipeline may have bypassed type errors.

### API errors affecting all users (frontend side)

1. Open the browser Network tab and check whether the failing requests are returning 5xx from `api.voxire.dev` or failing in the frontend itself.
2. If 5xx: the issue is in the backend - follow the `onion` incident guide.
3. If the request never leaves the browser (CORS error, network error): check that `VITE_API_URL` is set correctly in the Cloudflare Pages environment variables for this project.

### Monitoring

- **Error tracking:** `@voxire/observability` wires up Sentry (error tracking) and PostHog (product analytics) on the client — see `packages/observability/src/boot.ts`. Cloudflare Pages deployment logs and the browser console remain useful for debugging.
- **Escalation:** Contact Abed El-Fattah Amouneh directly - info@voxire.com / +961 3 940 708.

---

## 13. ENVIRONMENT SETUP

### Prerequisites

| Tool    | Version | Install                       |
| ------- | ------- | ----------------------------- |
| Node.js | 20+     | https://nodejs.org            |
| pnpm    | 10.33.0 | `npm install -g pnpm@10.33.0` |

### First-time setup

```bash
# Clone and enter the repo
cd rtylr

# Install all dependencies (resolves workspace packages automatically)
pnpm install

# Run all apps in parallel
pnpm dev
```

### Run a single app

```bash
pnpm --filter dash dev         # Dash on :5175
pnpm --filter @rtylr/auth dev  # Auth on :5176
pnpm --filter @rtylr/erp dev   # ERP on :5179
```

### Point to local backend

Create `apps/<name>/.env.local`:

```env
VITE_API_URL=http://localhost:9999/
```

Default without `.env.local` is `https://api.voxire.dev/` (production API).

### Build

```bash
pnpm build         # build all apps via Turborepo
pnpm typecheck     # type-check all packages + apps
pnpm test          # Vitest unit tests
pnpm test:e2e      # Playwright e2e tests
```

### Submodules (packages, and its nested voxie contracts)

`packages/` (tryrtylr/packages — all shared workspace packages) is a git submodule, and it carries its own nested `voxie` submodule at `packages/contracts/voxie` (proto contracts + committed `gen/ts`). After cloning or pulling:

```bash
git submodule update --init --recursive
```

`gen/ts` is generated and committed by `voxie` itself — there is nothing to run locally to sync contracts. To pick up a newer voxie commit and audit usage:

```bash
pnpm contracts:sync             # bumps the nested voxie submodule pointer to its latest commit
pnpm contracts:audit-requests   # audits app request shapes against the contracts
```

### Cloudflare Pages deployment

Each app is its own Cloudflare Pages project. Root directory: empty (monorepo root). Build command pattern:

```
pnpm install --frozen-lockfile && pnpm --filter dash run build
```

Output directory: `apps/dash/dist`.
