# 13. Backend inventory

The new half of round two: onion swept as a system. Every RPC handler and HTTP
route, every sqlc query, every table with its touchers, the import graph, every
platform package with a live or dead verdict, the async backbone, and the
schema corpus.

Generated on 2026-08-22 from `docs/cross-app-round2` at `cb7d6d5a`
(`origin/main` = `ab6ff7ff` plus the sweep scripts). Everything mechanical is a
committed script in `~/Development/onion/scripts/cross-app/` (branch
`docs/cross-app-round2`, checked out at
`~/Development/onion-worktrees/cross-app-round2` because another working
session shares the primary clone; see DECISIONS.md R2-6). The CSVs land beside
round one's in this directory.

```bash
cd ~/Development/onion-worktrees/cross-app-round2   # or any checkout of the branch
node scripts/cross-app/sweep-handlers.mjs
node scripts/cross-app/sweep-queries.mjs            # run before sweep-tables
node scripts/cross-app/sweep-tables.mjs             # reads backend_queries.csv
node scripts/cross-app/sweep-imports.mjs
node scripts/cross-app/sweep-platform.mjs
node scripts/cross-app/sweep-events.mjs
node scripts/cross-app/sweep-migrations.mjs
```

## Headline counts

| Artifact                | Records | One record is                                                                      |
| ----------------------- | ------- | ---------------------------------------------------------------------------------- |
| `backend_handlers.csv`  | 709     | one served route: 577 oneKit RPCs, 118 legacy chi routes, 15 platform-level routes |
| `backend_queries.csv`   | 1,393   | one sqlc query across 129 module-owned query sets                                  |
| `backend_tables.csv`    | 296     | one table with owner and every module that reads/writes it                         |
| `backend_imports.csv`   | 771     | one directed import edge (module-module, module-platform, platform-module)         |
| `platform_packages.csv` | 41      | one `internal/platform/*` package with a verdict                                   |
| `backend_events.csv`    | 26      | one wsbus publish, wsbus subscribe, registered job, or webhook-emission probe      |
| `migrations.csv`        | 157     | one schema file with tables created/altered                                        |

Handler counts by module: hr_core 148, crm_sales 127, careers 105, workspaces
81, hr_time 54, hr_shifts 49, insights 35, hr_payroll 29. Full distribution in
the CSV; 110 modules serve at least one route.

## The multi-module table list, ranked

76 of 296 tables are touched by more than one module. These are the undeclared
coupling rows the brief calls the most important output of this phase: nothing
in the type system records them, and each is a join waiting to be written
wrongly. Ranked by how many distinct modules touch them:

| Modules | Table                                                                       | Owner (schema dir)   | Writers beyond owner                                                       |
| ------- | --------------------------------------------------------------------------- | -------------------- | -------------------------------------------------------------------------- |
| 15      | `businesses`                                                                | shared               | business, console, user (+12 readers)                                      |
| 14      | `erp_products`                                                              | erp                  | pos_menu, store_products, platform/stock_core (+10 readers incl. insights) |
| 12      | `users`                                                                     | shared               | console, security, team, user                                              |
| 10      | `internal/modules/hr_employees`                                             | hr_core              | staff, team, user (+5 readers incl. workspaces)                            |
| 8       | `crm_customers`                                                             | crm                  | customers, platform/customer_sync, store_orders (+4 readers)               |
| 7       | `user_app_permissions`                                                      | shared               | team                                                                       |
| 6       | `pos_orders`                                                                | pos                  | pos_kds, pos_day_summary, pos_receipts write items/refs                    |
| 6       | `finance_invoices`                                                          | finance              | quotes, tax                                                                |
| 6       | `finance_expenses`                                                          | finance              | invoices                                                                   |
| 5       | `subscriptions` / `plans` / `internal/modules/hr_shifts` / `erp_categories` | shared/hr_shifts/erp | console, subscriptions, team / staff / categories                          |

Reading of the top rows:

- **`erp_products`: 14 modules.** POS menu items, storefront products, stock
  core, insights and kpis all read it; `store_products` and `pos_menu` also
  write it. This is the ERP->POS-01 coupling already flowing through storage,
  invisible to every contract.
- **`crm_customers`: 8 modules**, including writes from outside the customers
  module: `store_orders` writes it directly, and `internal/platform/customer_sync`
  upserts it for POS (`~/Development/onion/internal/platform/customer_sync/customer_sync.go:18`).
- **`team` and `user` write `internal/modules/hr_employees`.** Seat provisioning creates the HR
  employee row (`docs/sqlc-query-ownership.md` "Ensure employee for user"
  group). HR's person record is therefore already co-owned by hub-side
  modules, which raises the cost of `HR->HR-01` beyond an intra-HR migration.
- **`insights` reads five domains' tables directly**: finance*expenses,
  finance_invoices, pos_orders, hr*\* and erp_products all appear in its query
  files. Insights is a cross-domain reader by construction; that is its job,
  but it means margin-style aggregates need no new mechanism, only queries.
- **`stock_ledger` is written by both `internal/platform/stock_core` and
  `pos_stock_movements`** under the sanctioned stock_core seam, matching the
  ownership register's "Inventory transaction with reference" row
  (`docs/sqlc-query-ownership.md:34`).

## Platform package verdicts

All 41 packages inventoried in `platform_packages.csv`. The three states:

- **live** (36): imported by at least one module. The load-bearing ones:
  `onekiterrors` (130 importers), `logger` (143), `auth` (126), `cache` (127),
  `tenant` (101).
- **live-port** (3): contract-only packages whose implementation lives in a
  module and is wired in `cmd/api/main.go`. This is the shape the brief called
  "a seam declared and never filled" about labourcost; that verdict is now
  wrong:
  - `internal/platform/labourcost`: contract only, but `internal/modules/hr_payroll` implements
    `LabourCostProvider`
    (`internal/modules/hr_payroll/internal/application/service.go:53`) and
    workspaces consumes it via `SetLabourCostProvider`
    (`internal/modules/workspaces/internal/application/service.go:113`),
    wired at `cmd/api/main.go:466`. Landed 2026-08-22 in #743, one day after
    round one closed.
  - `internal/platform/financejournal`: implemented by the journal module
    (`internal/modules/journal/module.go:32`), consumed by hr_payroll via
    `SetFinancePoster` (`cmd/api/main.go:499`). Payroll export to the journal
    is a working server-side path: `POST /hr/payroll/runs/{id}/export`.
  - `internal/platform/financeaccounts`: implemented by accounts, consumed by journal
    (`cmd/api/main.go:498`).
- **dead-or-internal-only** (2):
  - `internal/platform/runmode`: zero importers anywhere. `docs/platform-patterns.md:5`
    recommends it for new runmodes; no code follows that advice. Doc drift,
    logged in 10-doc-drift.md.
  - `models`: one importer; effectively vestigial.

Two further observations from the same file:

- **`wsbus` has 4 importing modules** (workspaces, pos_orders, pos_kds, plus
  its own backplane wiring). The "existing asynchronous backbone" carries UI
  realtime events only: topics are `user:<id>` / `workspace:<id>` /
  `business:<id>` fanout to browsers (`internal/modules/workspaces/internal/api/project_service.go:59`,
  `internal/modules/pos_orders/internal/application/service.go:267`). No
  domain-integration event flows on it today. Every proposal in this audit
  that assumed "an event on wsbus" as existing infrastructure should instead
  read: the transport exists, the event vocabulary does not.
- **`jobs` has 6 registered jobs** (`backend_events.csv`), all scheduled
  maintenance: hr_leave monthly accrual, hr_overtime daily detection, hr_payroll
  period-open notifications, hr_time unsubmitted-timesheet reminder,
  recurring_invoices hourly generation, voxhire CV processing. No job exists to
  propagate a state change between domains.

## Webhooks: registration without emission

The webhooks module is CRUD over tenant webhook registrations
(`internal/modules/webhooks/internal/application/service.go:46-100`), exposed
to dash through `internal/platform/hubdeveloper`'s provider port
(`cmd/api/main.go:732`). A repo-wide search for code that POSTs to a stored
webhook URL finds exactly one outbound POST site: Paddle's client
(`internal/platform/paddle/client.go`), which calls Paddle's API, not tenant
webhooks. **Nothing delivers tenant webhook events.** A tenant can register a
webhook endpoint in dash and no event will ever reach it. This converts round
one's plumbing note ("check webhooks before inventing a fourth mechanism",
07-plumbing.md shape 3) into a concrete finding: the mechanism cannot be
checked because half of it does not exist.

## Tenancy coverage of the sqlc corpus

1,163 of 1,393 queries (83%) carry a `business_id` filter; 230 carry neither
`business_id` nor `owner_id`. The largest blocks: workspaces (53), console
(26), voxhire (20), pos_orders (19), memory (18).

Spot-checks show most of those are indirect tenancy: Flow cycles, labels,
milestones and notifications filter by `workspace_id`/`project_id` and rely on
membership checks upstream
(`internal/modules/workspaces/sql/queries/features.sql:2`). That is a design,
not a leak, but it has a consequence this audit must state: **any new
cross-module aggregate that joins these tables must scope through the owning
aggregate id chain, not assume a `business_id` column exists to filter on.**
The `memory` module's queries are the sharpest case: none of them reference a
tenant column at all (`internal/modules/memory/internal/api/memory_service.go`),
which makes Memory the one module whose tenancy posture could not be verified
from SQL alone. Flagged for the backend owners rather than resolved here.

## Migrations and the schema-to-contract mapping

157 schema files across 35 directories (`migrations.csv`). Schema directory
names do not match contract domain names one to one (`internal/modules/hr_core` schema serves
the `hr` contract; `book` schema serves `booking`; `shared` holds cross-domain
tables like `users` and `businesses`; `rtylr`, `voxire` and `recruit` have no
contract domain at all, which matches round one's finding that Recruit is
entirely uncontracted). The full per-file listing with created/altered tables
is in the CSV.

Per the brief's rules, `scripts/migrate.sh` was not executed and no database
was contacted. Round one's local-database tenancy measurements (flow-deep-dive
section 7.4) remain the runtime evidence of record and are cited, not re-measured.

## What this sweep could not reach

Stated plainly:

1. **Route literals registered by hand-built chi patterns other than
   `serverHttp.Register*`.** The legacy count (118) covers registrations
   through the typed helpers. Round one's looser string scan found 227
   path-shaped literals in module files; the difference is mostly dead
   literals and helper functions, but a module registering routes through a
   custom wrapper would be missed by both passes.
2. **Dynamic topic strings on wsbus** are recorded as their format expression
   (for example `fmt.Sprintf("workspace:%d", workspaceID)`); the sweep does not
   evaluate them. All observed instances resolve to the three documented topic
   families.
3. **Non-sqlc raw SQL in Go** is captured only where it uses squirrel builders
   or literal FROM/INTO/UPDATE/JOIN keywords. A query assembled by string
   concatenation would be missed. Two residual artifact rows in
   `backend_tables.csv` (`config`, `cache`) come from prose words adjacent to
   SQL keywords in platform Go files; they are retained rather than hand-filtered
   so the CSV stays purely generated.
4. **Jobs registered outside the `registry.Register(jobs.Job{...})` pattern**
   would be missed. `grep RegisterScheduledJobs` shows six modules using the
   pattern and the worker boot path auto-registers exactly those.
5. **gRPC surface** (`internal/platform/server/grpc`) was not swept: the AGENTS.md
   describes it as secondary and no frontend client references it.
