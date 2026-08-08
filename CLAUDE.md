# CLAUDE.md - rtylr frontend

RTYLR is a multi-tenant commerce operating system. This repo is the frontend monorepo: 10 Vite + React 19 apps and 16 shared `@voxire/*` packages, on top of the `onion` Go API.

Built by Voxire - Abed El-Fattah Amouneh & Mohammad Homsi. info@voxire.com

---

## BEFORE YOU CLAIM DONE

```bash
pnpm typecheck && pnpm lint && pnpm test
```

Filtered while iterating: `pnpm --filter <app> typecheck`. All three must pass with zero errors. Don't report work as complete on unrun gates.

---

## NEVER

- ~~Edit `apps/hr/**`. Owned by another developer, read-only.~~ **Lifted 2026-08-06 (Abed's call)** - `apps/hr` is in scope for editing. Mohammad still reviews at PR time like any other change, same as before. `.claude/settings.json`/the guard hook are being updated to match this - if you still hit a hard block there, that's a config lag, not a live restriction; ask Abed to update it rather than working around it.
- **Edit generated output:** `packages/voxie/gen/**`.
- **Edit `.env`, `.env.local`,** or put a real secret behind a `VITE_` prefix - those are inlined into the bundle.
- **Hand-edit `pnpm-lock.yaml`.**
- **Push to `main`,** commit, or open a PR unless asked. Branch → PR is the workflow (§Workflow), but the human decides when.
- **Change a submodule pointer** unless that is the point of the change.

## STOP AND ASK

Don't decide these alone: changing a `voxie` `.onk` contract · adding a dependency · flipping a feature flag · anything touching auth, tenancy/`owner_id`, money, or stock · a schema change · deleting a route or public export.

## DON'T TRUST THESE

They are known stale and will lead you to build the wrong thing:

- `README.md` in this repo - claims Vite 7 / TS 5.9 / Next 15 / 13 apps. Wrong.
- `vault/rtylr/apps/*` - predates the TanStack Router and onekit migrations. Describes `react-router-dom`, `routes/AppRoutes.tsx`, `@voxire/contracts`.
- `vault/voxie/*` - describes the retired protobuf/`buf` stack.
- `routes/AppRoutes.tsx` in flow, hr, insights - dead files. `src/router.tsx` is live.

**When docs and code disagree, the code wins** - then fix the doc.

---

## THE QUALITY BAR

Consistency beats cleverness. Before introducing a pattern, find the existing one and follow it.

1. **One way to do a thing.** If you find two, you are inside a migration: use the newer one and say so in the PR. Don't add a third.
2. **Never ship a control that doesn't work.** A button that cannot perform its action is disabled with a reason, or absent. No handler whose whole body is `toast.success(...)`. This repo has ~20 of those already - don't add the 21st.
3. **Make invariants impossible to break, not merely documented.** If correctness depends on a reviewer noticing, add the lint rule or the test instead.
4. **Types at the boundary.** Zod at the wire, generated contract types inside. No `any`, no `@ts-ignore`, no index signatures on domain types.
5. **Tests where the risk is.** Stock, payments, payroll, tenancy, auth. Coverage percentage is not the goal; a test on `owner_id` scoping is.
6. **Delete on the way past.** Dead route, dead file, stale comment, commented-out block - remove it in the PR that touched it.
7. **Prefer a smaller surface.** Declining to build something is a valid, often correct, engineering answer.
8. **Explain why in the PR.** The diff shows what. If you can't articulate why, the change isn't ready.

---

## ARCHITECTURE - THE INVARIANTS

**Apps** (all Vite + React 19, one Cloudflare Pages project each): `dash` 5175 · `auth` 5176 · `menu` 5177 · `erp` 5179 · `hr` 5180 · `crm` 5181 · `finance` 5182 · `flow` 5183 · `recruit` 5184 · `insights` 5187.
Not in this repo: `landing` (own repo), `vbitez`/POS (the `pos` repo), `docs`.

**`packages/` is a git submodule** (`tryrtylr/packages`) containing `packages/voxie` directly (vendored file copy, not a nested submodule - no `.gitmodules` entry, `git ls-files -s voxie` returns plain `100644` blobs), so there is no inner pointer to bump and nothing to `git pull` inside it: new contracts land as a direct file sync. Always `git submodule update --init --recursive` after a pull, then `pnpm install`. A stale pointer builds the wrong code with no error.

**Routing.** `@tanstack/react-router`. `src/router.tsx` is the live route tree in every app - never add routes to `AppRoutes.tsx`. Path params use `$id`, not `:id`. **Every route component is lazy-loaded** behind `<Suspense fallback={<FullScreenLoader />}>`; layouts are the exception.

**Guards, outermost first.** `ErrorBoundary` (from `@voxire/ui`) is always outermost and is present in every app's root route today - keep it that way. Then `AppFlagGuard` (surface flag), then `AppGuard` (cookie + `/user/me` + `allowedRoles`). dash uses `BusinessRoute` instead; hr adds `HRConfigRoute`. **Dashboard widgets get one `ErrorBoundary` each** - a shared one means one broken widget kills the page.

**Boot order in `main.tsx`** - getting this wrong causes theme flash or a self-host pointing at the wrong API:
`initTheme()` → `injectTheme()` → `bootObservability({app})` → render `<FullScreenLoader />` → `loadRuntimeConfig().finally(() => render tree)`.

**Runtime config / self-hosting.** Self-host ships prebuilt images, so apps learn their URLs from `GET /runtime-config` at boot. **Read URLs through the resolvers** - `resolveApiBase()`, `resolveAuthUrl()`, `resolveAppUrl(id)`, `resolveCdnBaseUrl()`, `resolveCookieDomain()` - never `APP_URLS.*` directly in new code. `isSelfHosted()` / `allFeaturesEnabled()` are the posture helpers. Hosted is a strict no-op. Defaults: API `https://api.voxire.com/`, CDN `https://cdn.voxire.com`.

**Feature flags.** `FLAG_KEYS` in `packages/config/src/flags.ts`. Resolution: PostHog → `VITE_APP_*_ENABLED` → `FLAG_PROD_DEFAULTS` → `allFeaturesEnabled()` forces true on self-host. **All app surfaces now default ON in prod.** Only `app.pos.payments`, `app.recruit.fileUpload`, `app.flow.previewRoutes` are off. Before flipping anything, read `vault/rtylr/audit/per-app-audit/<app>.md`.

**State.** Server state → TanStack Query, in `shared/hooks/queries/`. UI state → Zustand, in `shared/stores/`, persisted via `persist` + `partialize`. **Never mirror API data into Zustand** (exception: app-wide config/theme loaded once). **Never persist tokens or PII.**

**Auth.** `token` cookie scoped `.rtylr.com`; its existence is the synchronous logged-in signal. Never read the token from a store or prop. **Logout must clear both cookies** or auth bounces the user straight back in:

```ts
removeCookie('token', { domain: getDefaultCookieDomain() });
removeCookie('token');
window.location.href = `${resolveAuthUrl()}?redirect=<appId>`;
```

Never write `document.cookie = 'token=; …'` by hand. `UserRole` is `owner | admin | operator | staff | viewer`.

### API layer - two live patterns

**Prefer the generated onekit client** when a voxie contract exists:

```ts
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { LoyaltyServiceClient } from '@voxire/voxie/crm/loyalty/v1/client';
const loyaltyClient = createOneKitClient(LoyaltyServiceClient, 'crm');
// errors: getOneKitErrorMessage(e, t('errors.generic')) - never read e.message
```

Generated TS fields are **camelCase**; `decode*`/`encode*` handle the snake_case wire. **Add no key conversion on this path.**
Reference: `apps/crm/src/shared/hooks/queries/useLoyalty.ts`

**Otherwise `safeFetch`:** `unwrap(safe.get<T>('/path'))` from `shared/lib/api.ts`. On this path `@voxire/api` interceptors convert keys automatically - **never convert manually**.
Reference: `apps/dash/src/shared/hooks/queries/useTeam.ts`

**Never raw `axios.get(...)` in a component or hook.** Both paths handle 401 by clearing the cookie and redirecting.

**Contracts** come from `@voxire/voxie/{app}/{capability}/v1/{types|client}`. `@voxire/contracts` no longer exists. Never copy-paste a contract shape - import it.

---

## HARD RULES

1. **TypeScript strict per app.** No `any`, `@ts-ignore`, unsafe assertions. If unavoidable, comment why.
2. **Explicit return type on every function.** `const f = (): string => {}`.
3. **Zod at every external boundary** - API responses, form input, `localStorage`, env vars.
4. **All user-facing strings via i18next `t()`** - including toasts, empty states, placeholders, ErrorBoundary fallbacks. **All three locales properly translated before merge** - `ar`/`fr` left as the English string is not done. An `ar` change also means checking the RTL layout with `rtl:` variants, not just adding the key. Locales are `locales/{en,ar,fr}/<ns>.json` (hr is flat; **recruit has none - fix when you touch it**).
5. **No hardcoded hex in component code.** Reference a CSS custom property. Opacity via `color-mix(in srgb, var(--x) 30%, transparent)`.
6. **Mobile-first only** - `min-width` breakpoints, ≥16px font on inputs, ≥44×44px touch targets.
7. **a11y**: `aria-label` on every icon-only button, real `<label>` on inputs, Radix `Dialog` for focus trapping, `aria-live="polite"` on updating regions.
8. **Paginate every list query** - pass `limit`/`offset`, respect `{items,total,limit,offset}`. Virtualise anything over ~100 rows with `@tanstack/react-virtual`.
9. **No `console.log`** (repo is currently clean), no commented-out code, no direct DOM manipulation outside hooks, no inline styles except dynamic CSS variables.
10. **Import from `@voxire/*`**, never a relative path across a package boundary.
11. **Never hardcode a business rule.** See Configurability below.
12. **No dead user-visible controls.** If a user can reach it and click it, it works end-to-end - or it sits
    behind an off flag. **Never a success toast for something that mutated nothing.** A disabled control with a
    reason is honest; a fake confirmation is not. Wired-but-unreachable code is fine. Boy-scout rule: open a file
    containing a dead control and you fix that control before you leave.
13. **A frontend change carries its backend.** If the screen needs an endpoint that does not exist, building it
    is part of this task, not a follow-up. Order is fixed: voxie contract merges, then the onion endpoint, then
    this repo's consumer with the pointer bump. Never build the screen first and hope.

### Configurability - the product depends on this

RTYLR sells the same binary to restaurants, law firms, and retail groups. Before hardcoding anything, ask: _would a restaurant and a law firm want this to behave differently?_ If yes, it is configuration, not code.

Never hardcode: role names in permission checks (`user.role === 'manager'`), feature visibility, approval-chain labels or step counts, leave-type names, notification channels. Read them from company config - `configStore` (dash), `hrConfigStore` (hr).

Missing config: **non-critical** → sensible default or empty state, never an error. **Critical** (payroll, timesheets, anything touching money or compliance) → a clear prompt linking to Settings, never assumed values.

**Boy scout rule:** when you open a file, fix any hardcoded business rule in it.

### Component placement

`@voxire/ui` primitive exists? Use it. Higher-level shared? `@voxire/components`. AI assistant? `@voxire/caret` (mounted in dash, crm, erp, finance, flow, hr, insights). Feature-specific? `features/<x>/components/`. Reused in one app? `shared/components/`. Likely needed by other apps? Propose adding it to `@voxire/ui`.

---

## WORKFLOW

1. `git checkout main && git pull origin main` - branch from **latest** main, never a stale local main or another branch.
2. `git submodule update --init --recursive`
3. `git checkout -b <type>/<slug>` - `feat|fix|chore|refactor|style|docs|test|perf`, kebab-case. **No `claude/` or `codex/` prefixes.**
4. Work. **One logical unit per commit** - if the subject line needs an "and", it is two commits. A component plus the locale keys it introduces is one unit; two unrelated fixes are two. Every commit must build on its own, so group files when splitting them would break `typecheck`. Short imperative subject, no conventional prefix. No `.md` files except `CLAUDE.md` in any directory - it is agent configuration, not documentation. All prose documentation belongs in the vault.
5. Gates (top of this file).
6. Push, open a PR. `.github/pull_request_template.md` is the shape: **Why / What changed / Risk / Gates**. The body explains **why**; the diff shows what.
7. Green CI + one approving review from `@1homsi` (Abed when `@1homsi` authors) → squash merge → delete branch.

**Commit identity - non-negotiable.** Author every commit as `amounehabed@gmail.com`. Set it per repo so it
can't drift: `git config user.email amounehabed@gmail.com` (history already carries `abed@voxire.com` and
`info@voxire.com` for the same person - don't add a fourth). **No AI attribution anywhere** - no `Co-Authored-By`
trailers, no "Generated with", no tool, model, or agent names in commit messages, branch names, PR titles, or PR
bodies. The human who reviewed the change is its author and is accountable for it; that is the only attribution
the history carries.

**Direct pushes to `main` are not allowed. No exceptions** - not one-line fixes, not config tweaks. If prod is down, still branch and ask for an immediate review.

Rebase if main moved: `git fetch origin && git rebase origin/main`. Keep PRs under ~400 lines of real change.

**Cross-repo pointer chain:** `voxie → packages → rtylr` (and `pos`); `voxie → onion`. Merge upstream first; the pointer bump belongs **in** the consuming PR, never referencing an unmerged branch. **A `packages/` change redeploys every app** - say so in the PR body.

---

## GOTCHAS THAT WASTE A SESSION

- Forgetting `git submodule update --init --recursive` - builds the wrong code, no error.
- Copying `.env.example` verbatim - its flags are all `0` and hide six apps that are ON in prod.
- Cross-app SSO doesn't work across `localhost:5175` / `:5176` - no shared cookie domain. Map a hosts entry onto `*.rtylr.com`.
- `pnpm test && pnpm turbo build` runs on pre-push (husky) and is slow. Run filtered while iterating.
- `useState(() => readCookie())` causes a React 19 hydration mismatch that silently eats click handlers. Read in `useEffect`.

---

## WHERE THE DOCS LIVE

**`~/Development/vault`** is the written source of truth. There is **no `rtylr/docs/`**.

| Need                                                                | Read                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------ |
| Frontend architecture, state, theming, i18n, testing, observability | `vault/rtylr/<topic>.md`                                     |
| Why a choice was made                                               | `vault/rtylr/decisions.md`                                   |
| One app's inventory                                                 | `vault/rtylr/apps/<app>/code.md` (stale on router/contracts) |
| Shared packages                                                     | `vault/rtylr/packages.md`, and `packages/CLAUDE.md`          |
| Pre-launch gaps, security, a11y, per-app readiness                  | `vault/rtylr/audit/*` (reviewed 2026-04-27)                  |
| Backend                                                             | `vault/onion/*`, `~/Development/onion/CLAUDE.md`             |
| Contracts                                                           | `~/Development/voxie/CLAUDE.md`                              |
| Product, pricing, wedge                                             | `vault/rtylr/business/*`, `vault/rtylr/research/*`           |
| ERP + unified stock                                                 | `vault/rtylr/erp-readiness.md`                               |

Per-app detail lives in `apps/<app>/CLAUDE.md`, loaded when you read files there.

**Update `vault/rtylr/` on any architectural or pattern-level change**, and this file if you add a rule.

---

## KNOWN ISSUES

| Issue                                                                                                         | Where                                             | Priority |
| ------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- | -------- |
| `.env.example` flags are `0` while `FLAG_PROD_DEFAULTS` are `true` - copying it hides live apps               | `.env.example`, `packages/config/src/flags.ts`    | High     |
| `tsconfig.base.json` lacks `strict: true`; a new app can silently ship non-strict                             | `tsconfig.base.json`                              | High     |
| `apps/recruit` has no `locales/` directory                                                                    | `apps/recruit/src/`                               | Medium   |
| `StoredUser` has `[key: string]: unknown`; `@voxire/auth` `User` has `[key: string]: any`                     | dash `userStore.ts`, `packages/auth/src/types.ts` | Medium   |
| ~20 toast-only buttons that mutate nothing (ERP PO send/approve, Finance close-year, dash Email/Domain/Theme) | `vault/rtylr/audit/per-app-audit/*`               | Medium   |
| 20 `test.describe.skip()` blocks across 8 e2e specs; no coverage thresholds in CI                             | `e2e/`                                            | Medium   |
| hr uses flat locale files                                                                                     | `apps/hr/src/locales/`                            | Low      |
| Dead `routes/AppRoutes.tsx` in flow, hr, insights; `CreateItem.jsx` mixed extension in dash                   | those dirs                                        | Low      |
| Query-key conventions inconsistent (`['team']` vs `['crm-customers', …]`)                                     | `shared/hooks/queries/`                           | Low      |
