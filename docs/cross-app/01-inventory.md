# Phase 1: Inventory

Machine-generated inventory of every contract, request, action and entity in
the RTYLR frontend, plus the backend modules that serve them. Generated on
2026-08-21 from branch `docs/cross-app-data-sharing-audit` (forked from
`origin/main` at `ab006407`), packages submodule `7596d813`, vendored voxie at
`packages/voxie` commit `0e48b902`, and a read-only onion checkout at
`~/Development/onion`.

## How to regenerate

Run from the repo root, in this order (later sweeps join against earlier CSVs):

```bash
node scripts/cross-app/sweep-contracts.mjs
node scripts/cross-app/sweep-endpoints.mjs
node scripts/cross-app/sweep-actions.mjs        # independent of the others
node scripts/cross-app/sweep-entities.mjs
node scripts/cross-app/sweep-backend-modules.mjs  # needs ONION_DIR or ~/Development/onion
```

Each script documents its own row semantics in its header comment. Anything a
script could not parse or match is printed to stderr, never silently dropped.

## Counts

| Artifact                        | Rows                                           | Source                             |
| ------------------------------- | ---------------------------------------------- | ---------------------------------- |
| `inventory/contracts.csv`       | 7,388 (1,453 messages, 5,316 fields, 619 RPCs) | 299 `.onk` files, 0 unparsed lines |
| `inventory/endpoints.csv`       | 715 request call sites                         | 10 apps + `packages/*`             |
| `inventory/actions.csv`         | 1,241 user-action rows                         | 10 apps                            |
| `inventory/entities.csv`        | 178 persisted entities                         | derived from contracts.csv         |
| `inventory/backend_modules.csv` | 133 onion modules                              | `onion/internal/modules/*`         |

### Contract files per domain

booking 14, common 5, console 2, crm 16, erp 24, finance 28, flow 24, hr 54,
hub 48, insights 16, pos 58, support 10. Total 299: every capability is a
`models.onk`/`service.onk` pair except `common/pagination/v1`, which has a
`models.onk` and no service (150 model files, 149 service files).

### Requests per app, by client idiom

All three idioms were swept: `safeFetch` (481 call sites), OneKit generated
clients (223, every one resolved to a contract RPC), raw axios (33, apps/menu
plus the shared clients in `packages/`).

| App         | Requests | Notes                                                                                         |
| ----------- | -------- | --------------------------------------------------------------------------------------------- |
| hr          | 148      | safeFetch + onekit                                                                            |
| crm         | 132      | safeFetch dominant                                                                            |
| finance     | 112      | safeFetch + onekit; several legacy paths, see below                                           |
| flow        | 97       | safeFetch; includes 11 calls into `hr.*` and 8 into `hub.*`                                   |
| dash        | 90       | onekit-heavy; reads `pos`, `insights`, `erp`, `hr` domains                                    |
| erp         | 58       | mostly onekit (this is why a naive `safe.` grep reports 5)                                    |
| insights    | 30       | almost entirely onekit (naive grep reports 2)                                                 |
| menu        | 28       | raw axios, public app                                                                         |
| recruit     | 23       | safeFetch against legacy unprefixed paths                                                     |
| auth        | 13       | onekit against `common`/`hub`                                                                 |
| packages/\* | 7        | `@voxire/api` (5), `@voxire/auth` (1: `/user/me`), `@voxire/components` (1: `/license/state`) |

### Requests per backend domain (by path prefix)

hr 156, crm 125, hub 104, finance 101, flow 78, insights 35, erp 51, pos 12,
common 11, support 2, booking 0, console 0, plus 62 rows whose path carries no
domain prefix (legacy routes and dynamic paths).

The app-to-domain crossings observed here (an app calling a domain that is not
its own) are the seed of the Phase 3 sharing matrix; the largest are
flow→hr (11), dash→pos (9), crm→hub (9), dash→insights (6), finance→crm (3),
menu→pos (3), crm→support (2).

### Backend modules

133 modules under `onion/internal/modules/`. 59 matched no frontend request
from this repo. The bulk of those are expected: `booking_*` (7), `pos_*` (17
of 24), `store_*` (8), `support_*` (4 of 5) serve the separate POS terminal
repo, the not-yet-built booking/support surfaces, and the storefront. The
remainder are genuine findings, either infrastructure (`cacheops`, `echo`,
`emailtest`, `memory`, `general`, `console`) or modules whose frontend has not
been wired (`flow_insights`, `finance_reconciliation`, `one_on_ones`,
`timesheets`, `tax`, `training`, `reviews`, `documents`, `webhooks`,
`billing`, `addons`, `aiquery`, `crm_settings`, `websitelead`, `voxhire`).
See `docs/cross-app/10-doc-drift.md` for the ones that contradict shipped UI.

## Route-serving mismatches found by the sweep

63 frontend request paths matched no onion module route. These are listed by
the sweep on stderr and fall into three classes:

1. **Legacy paths the backend no longer registers.** Verified individually:
   `/finance/reconciliation` and `/finance/reconciliation/match`
   (`apps/finance/src/shared/services/financeApi.ts:320` and `:326`) have no
   registration anywhere in onion; the reconciliation module serves only the
   generated `/finance/reconciliation/v1/reconciliation`
   (`onion/internal/modules/finance_reconciliation/module.go:22`). The same
   shape applies to `/finance/journal` (served only as
   `/finance/journal/v1/entries`) and `/team/members` (served only as
   `/hub/team/members/v1/members`). Each is verified per-path in
   `10-doc-drift.md` before being called broken.
2. **Routes registered by a mechanism the sweep does not parse** (path built
   by concatenation, or registered outside `internal/modules`). `/user/me`
   is called by `packages/auth` and every app's `AppGuard` and demonstrably
   works in production, so its registration exists somewhere the sweep cannot
   see. Treat NO-MODULE as "unverified", not "broken", until checked by hand.
3. **Dynamic paths** the static sweep cannot resolve (8 rows marked
   `(dynamic)` or `/<dynamic:...>`, mostly `apps/menu`'s `api_routes` config
   object and one availability path in hr).

## Coverage, stated honestly

- **Apps: 10 of 10 swept.** Note the root `CLAUDE.md` lists an eleventh app,
  `apps/docs`; it does not exist on disk. Recorded in `10-doc-drift.md`.
- **Contract domains: 12 of 12 swept**, zero unparsed lines.
- **Onion modules: 133 of 133 swept** for imports and route strings. Route
  matching is best-effort; the 63 unmatched frontend paths and 59
  unconsumed modules above carry the uncertainty explicitly.
- **Client idioms: all three swept**, plus the legacy `shared/services/*.ts`
  wrapper layer and feature-local `use*.ts` hooks (the ERP pattern), which
  are caught because the sweep walks every `src/**/*.ts(x)` file rather than
  fixed directories.
- **Known blind spots.** (a) Requests whose path is assembled from variables:
  8 rows, marked `(dynamic)`. (b) WebSocket/realtime traffic
  (`packages/api/src/realtime.ts`) is not a request and is not inventoried;
  it is discussed in `07-plumbing.md`. (c) `actions.csv` attributes labels
  and mutations lexically (nearest declaration, nearest `t()` call); each row
  cites `file:line` so any single row can be re-verified in one click.
  (d) Test files are excluded everywhere.

## Reading the CSVs

- `contracts.csv`: one row per message declaration (empty `field_name`), per
  field, and per RPC (`name` is `Service.method`; `doc_comment` starts with
  the HTTP binding). The wire path of a OneKit RPC is
  `/<domain>/<capability>/<version><binding path>`, confirmed against the
  generated servers (e.g. `packages/voxie/gen/go/finance/reconciliation/v1/server.gen.go:255`).
- `endpoints.csv`: one row per browser-leaving call site. `backend_domain` is
  derived from the resolved path prefix.
- `actions.csv`: one row per JSX action attribute
  (`onClick`/`onSubmit`/`onConfirm`/`onSave`/`onDelete`/`onCreate`/
  `onExport`/`onImport`, with the balanced handler expression captured) and
  per cross-app navigation (`APP_URLS` links: 33 rows, `window.open`,
  `location.href`). `entities_touched` holds the backend domains the same
  file requests (a file-level approximation); `cross_domain_candidate` is
  `yes` when that set has 2+ domains.
- `entities.csv`: messages with an `id` field, excluding Request/Response
  envelopes. `foreign_keys` lists every `*_id`/`*_ids` field with type and
  nullability.
- `backend_modules.csv`: per module, the generated capabilities it imports,
  the legacy route strings it declares, and the apps whose requests matched
  either.
