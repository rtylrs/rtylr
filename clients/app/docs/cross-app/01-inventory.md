# 01. Inventory

What was swept, how, what was found, and what the sweep could not reach.
Generated on 2026-08-21 from branch `docs/cross-app-data-sharing-audit`
(base `ab006407`). Regenerate everything with:

```bash
node scripts/cross-app/sweep-contracts.mjs
node scripts/cross-app/sweep-endpoints.mjs        # requires contracts.csv
node scripts/cross-app/sweep-entities.mjs         # requires contracts.csv
node scripts/cross-app/sweep-backend-modules.mjs  # requires both, and ~/Development/onion
node scripts/cross-app/sweep-actions.mjs
```

## Headline counts

| Artifact                        | Records | What one record is                                           |
| ------------------------------- | ------- | ------------------------------------------------------------ |
| `inventory/contracts.csv`       | 7388    | one message, message field, or RPC from the 299 `.onk` files |
| `inventory/endpoints.csv`       | 743     | one browser-leaving request call site                        |
| `inventory/actions.csv`         | 1241    | one JSX action handler (superset; filtered during analysis)  |
| `inventory/entities.csv`        | 178     | one persisted noun (id-bearing contract message)             |
| `inventory/backend_modules.csv` | 133     | one onion module with its domains and consumers              |

Contract totals: 1579 messages, 619 RPCs, across exactly the 12 domains
`booking, common, console, crm, erp, finance, flow, hr, hub, insights, pos,
support` (`packages/voxie/api/*`).

## Requests per app, by client idiom

The three idioms from the brief were all found, plus a fourth (raw `fetch`)
the brief did not list:

| App                 | safeFetch | onekit | axios | fetch |
| ------------------- | --------- | ------ | ----- | ----- |
| auth                | 1         | 12     | 0     | 0     |
| crm                 | 111       | 21     | 0     | 0     |
| dash                | 70        | 20     | 0     | 0     |
| erp                 | 5         | 53     | 0     | 0     |
| finance             | 72        | 40     | 0     | 0     |
| flow                | 74        | 23     | 0     | 0     |
| hr                  | 122       | 26     | 0     | 0     |
| insights            | 2         | 28     | 0     | 0     |
| menu                | 0         | 0      | 28    | 0     |
| recruit             | 23        | 0      | 0     | 1     |
| packages/api        | 0         | 0      | 5     | 1     |
| packages/caret      | 0         | 0      | 0     | 2     |
| packages/components | 1         | 0      | 0     | 1     |
| packages/config     | 0         | 0      | 0     | 1     |

The raw `fetch` sites are Caret's streaming endpoints
(`packages/caret/src/hooks/useCaretStream.ts:79`,
`packages/caret/src/components/CaretPanel.tsx:272`) and recruit's presigned
resume upload (`src/modules/recruit/shared/hooks/queries/usePublic.ts:132`),
which uploads to object storage, not to the API.

`packages/api/src/realtime.ts` additionally opens a WebSocket channel; it is
not an HTTP request and is not in `endpoints.csv`, but it is cross-app
plumbing and is covered in `07-plumbing.md`.

## Which app already reads which backend domain

This table is the empirical baseline for the whole audit: every cell where an
app's requests resolve to a domain other than its own is an existing
cross-domain read ("IS shared" by definition).

| App      | Requests by resolved domain                               |
| -------- | --------------------------------------------------------- |
| auth     | common:11, hub:2                                          |
| crm      | crm:121, hub:9, **support:2**                             |
| dash     | hub:73, **pos:9**, **insights:6**, **erp:1**, **hr:1**    |
| erp      | erp:50, hub:4, unmapped:4                                 |
| finance  | finance:101, **crm:3**, hub:3, **insights:1**, unmapped:4 |
| flow     | flow:78, **hr:11**, hub:8                                 |
| hr       | hr:144, hub:1, unmapped:3                                 |
| insights | insights:28, hub:1, unmapped:1                            |
| menu     | **pos:3**, hub:2, **crm:1**, unmapped:22                  |
| recruit  | hub:1, unmapped:23                                        |

`hub` reads are not bolded: every app reads hub for auth, team and config by
design. "Unmapped" rows are paths whose first segment is none of the 12
domains; see the coverage notes below.

## Backend modules

133 modules in `~/Development/onion/internal/modules`. 59 have **no matched
frontend consumer in this repo**. That is expected, not alarming, for three
groups, and a finding for the rest:

- `pos_*` (22 modules) and `store_*` / `storefronts` (8) serve the separate
  POS terminal repo and the storefront, not these 10 apps.
- `booking_*` (7) and `support_*` (5) serve app surfaces that do not exist in
  `apps/` at all: there is no booking app and no support app in this repo,
  yet both domains have full contracts and backend modules. The CRM app
  reads `support` read-only (2 calls). This is the clearest case of backend
  capability with no operator surface.
- `console` (1) serves the internal admin console.
- The remainder (`addons`, `aiquery`, `echo`, `emailtest`, `memory`,
  `voxhire`, `websitelead`, and others) are flagged in
  `backend_modules.csv` with `has_frontend_consumer=no` for case-by-case
  review; several look like scaffolding or serve non-browser callers.

## App-to-domain mismatch (a structural finding)

Frontend apps and backend domains do not map one to one:

- **recruit** (the app) has **no contract domain at all**. All 23 of its
  requests hit legacy `/recruit/*` paths served by the `careers` module,
  which registers plain route strings and imports no generated contract
  (`~/Development/onion/internal/modules/careers/module.go`). Recruit is the
  only app whose entire API surface is uncontracted.
- **menu** (the app) reads `pos`, `crm` and `hub`, and 22 of its 28 requests
  go to legacy `/menu/*` and `/config*` paths (plus six `API_ROUTES.*`
  constants in stores) with no `.onk` contract behind them.
- **booking** and **support** (the domains) have contracts and backend
  modules but no app in this repo.
- **dash** is a window onto `hub` plus slices of `pos`, `insights`, `erp`
  and `hr`.

## Coverage, stated honestly

What the sweep reached:

- All 299 `.onk` files parsed; 23 lines were non-doc `//` separators and are
  skipped by design; everything else parsed or would be printed to stderr as
  `UNPARSED` (currently zero).
- All 10 apps and every `packages/*/src` tree, all four idioms. Raw
  per-app `safe.*` call-site counts were reconciled exactly against an
  independent grep before test-file exclusion differences were accepted.
- All 133 onion module directories.

What it could not reach or resolve:

- **13 dynamic paths** are recorded as `<dynamic:IDENT>` rather than a
  literal path: 6 `API_ROUTES.*` constants in `apps/menu` stores, HR's
  `AVAILABILITY_PATH` constant, and pass-through wrappers in
  `packages/api`. They are countable but their targets were resolved by
  hand only where the analysis needed them.
- **Test files are excluded everywhere** (`*.test.ts*`): mocks are not
  requests.
- **Comment lines are blanked before scanning**, because JSDoc examples
  (e.g. `api.get('/user/me')` in `packages/auth/src/guards/AppGuard.tsx:93`)
  otherwise count as requests. Same-line trailing comments are not blanked;
  none were observed to contain call-shaped text.
- **`queryKey` attribution is a heuristic** (nearest `queryKey:` within a
  window). Empty cells mean "not found near the call", not "no key".
- **actions.csv is a mechanical superset.** `entities_touched` there is the
  set of backend domains the same file requests, which is file-level, not
  per-button. Per-button attribution is done by hand in `03-sharing-matrix.md`
  and the dossiers for the actions that matter.
- **67 request paths matched no onion module route** (printed to stderr by
  `sweep-backend-modules.mjs`). Spot checks show three distinct causes:
  platform-level routes outside `internal/modules` (`/license/state`,
  `~/Development/onion/internal/platform/server/http/license_routes.go:21`),
  legacy frontend calls to endpoints that no longer exist (see
  `10-doc-drift.md` and the dossiers; the confirmed example is Flow's
  `/hr/staff/detail` at `the /hr/staff/detail call site (file since rewritten; see 15-round-one-verification.md)`,
  served nowhere in onion), and unresolved `<dynamic:...>` rows.
- The onion sweep reads route strings and generated-import evidence; it does
  not execute the router. A module that composes paths at runtime would be
  under-attributed. None of the spot-checked modules do.

## Round two coverage

The sweep was regenerated against this branch (contracts `847a34c8` via
packages `5ece9b79`) and extended: pos swept as an eleventh consumer surface,
the apps' migrated transport idiom (`hrTransport` and friends delegating to
safe) recognized, OneKit client declarations resolved scope-wide after the
centralization into shared/lib, template-literal paths folded to route shapes,
and constant folding applied to previously dynamic paths.

| Artifact            | Round one | Round two | Change                                                                |
| ------------------- | --------- | --------- | --------------------------------------------------------------------- |
| Request call sites  | 743       | 682       | menu migrated to oneKit (-28 axios), dynamic paths folded, pos +63    |
| Idioms              | 4         | 3         | raw axios eliminated from apps (515 onekit / 160 safeFetch / 7 fetch) |
| Unresolved dynamics | 13        | 1         | recruit presigned upload remains (object storage, not the API)        |
| onekit NO-RPC       | several   | 0         | RPC metadata join fixed for prefixed client names                     |

Per-app request counts (round two): hr 150, flow 101, finance 93, dash 89,
pos 63, erp 58, crm 54, insights 30, recruit 24, auth 14, menu 5, plus
packages 10. Backend coverage figures live in `13-backend-inventory.md`.
