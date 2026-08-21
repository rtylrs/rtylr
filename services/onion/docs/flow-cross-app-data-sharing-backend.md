# Flow cross-app data sharing: the backend half

Audit date: 2026-08-21
Repo: `tryrtylr/onion`, branch `main` at `4cbb8625`
Contracts audited against: `voxie` submodule at **`0e48b902`** (working tree and recorded gitlink agree)
Schemas submodule: `2aca6bd3`
Database observations: local container `onion-env-db-1` (MySQL 8.0.46), database `onion_dev`. Read only. Production was never queried.

This is the server-side mirror of the Flow cross-app audit that ran in `tryrtylr/rtylr`. Its job is to say which of the frontend's proposals are buildable, what mechanism each should use, and what has to be fixed first.

---

## 1. Executive summary

Six things matter more than the rest of this document.

**1. Flow has three backend modules, not six.** The commonly cited six-owner map is a grep artifact. `flow_[a-z_]+` with no left anchor matches `cash_flow_minor_units` and `cash_flow_report`, which is why `insights`, `finance_reports` and `echo` appeared on the list. None of the three touches a `flow_*` table. Evidence in section 3.

**2. The Flow-to-HR seam is already built, and it is invisible to the contract.** `workspaces` receives `hr_core` and `hr_leave` by constructor injection (`cmd/api/main.go:462`) and genuinely uses them (`workspaces/presence.go:52-74`). Its `sqlc.yaml` already lists the `hr_time` and `hr_core` schema directories with a written rationale (`workspaces/sql/sqlc.yaml:8-20`). The endpoint that consumes all of this, `/flow/presence/today`, has no contract entry. Any audit that reads contracts alone will propose building what already ships.

**3. `owner_id` means three different things on the Flow-to-HR path, and two of them collide in one table.** `hr_time_entries.owner_id` holds a resolved business owner user id when `workspaces` writes the row (`worktime_service.go:62`) and holds the business id itself when `hr_time` writes it (`hr_time/internal/infrastructure/repository.go:90`). `flow_workspaces.owner_id` holds the acting user's own id (`workspaces/internal/infrastructure/repository.go:1290`). This is the single blocker on the project-cost example. Detail in sections 6 and 8.

**4. Flow has two live time stores, not one.** `flow_time_logs` is actively written (`workspaces/sql/queries/workspaces.sql:139`), drives `flow_tasks.actual_hours` (`helpers.sql:25`), and is the table the HR importer reads (`hr_time/internal/infrastructure/flow_adapter.go:68`). `hr_time_entries` carries the separate work-day lines. The premise that Flow's hours already live entirely in `hr_time_entries` is false, which changes the recommendation on the import chain.

**5. The best mechanism in this repo is not on the usual menu, and it is already proven.** `internal/platform/financejournal/contract.go` declares a DTO plus a port; `journal` implements it, `hr_payroll` consumes it, and neither module imports the other. It is wired in both binaries (`cmd/api/main.go:494`, `cmd/worker/main.go:129`) with a no-op default so the consumer works unwired. It crosses the generation boundary cleanly: `hr_payroll` is legacy chi, `journal` is onekit. Section 7 recommends it for most new Flow links.

**6. The tenancy exposure is real but narrower than a row count suggests, and it resolves differently per table.** `flow_views` is a confirmed live defect: the INSERT omits `business_id` (`features.sql:75`) and 55 of 55 rows in `onion_dev` sit in the zero bucket. `flow_workspaces` is a confirmed live defect reachable on every registration and every login (`cmd/api/main.go:729-733`). `flow_tasks` shows 359 of 363 zero rows but its INSERT has named `business_id` since 2026-06-19, so those rows cannot come from today's handler path. Same symptom, three different causes, three different fixes.

---

## 2. Coverage statement per layer

| Layer | Coverage | What is not covered |
|---|---|---|
| Schema | Complete for `flow/` (8 files, 22 tables) and for the counterpart tables named in the links below (`hr_time_entries`, `hr_employees`, `finance_invoices` by column inspection) | Whole-schema FK census across all 30-plus domain directories |
| Repository | Complete for `workspaces`, `flow_insights`, `hr_time`. Cross-schema census done two ways: by annotation comment and by `sqlc.yaml` schema borrowing | Line-by-line audit of the roughly 40 `insights` cross-schema reads; sampled, not exhaustive |
| Service | Complete. Every `NewModule` call in `cmd/api/main.go` taking more than `ctx, db` was enumerated; all three `hub_*_provider.go` files and both platform-port packages characterised | Nothing material |
| Handler | Complete for Flow. All 6 hand-registered Flow chi routes, all 79 Flow contract method-and-path pairs, and the `hr_time` route set diffed against `hr/time_entry/v1` | The platform-wide handler-versus-contract diff. rtylr#873's count of 109 of 227 stands as the baseline; my independent literal grep gives 238 paths inside `*Routes()` bodies, consistent with it |

I did not use `contracts:audit-requests` as evidence anywhere, per the standing caution that it is blind to roughly half the route surface.

---

## 3. The real Flow ownership map

### Correction to the accepted map

`insights`, `finance_reports` and `echo` do not reference Flow. What matched:

| File:line | Matched token |
|---|---|
| `insights/internal/application/financial_cash_flow.go:14` | `flow_minor_units` |
| `insights/tool_executor.go:269` | `flow_minor_units` |
| `finance_reports/tool_executor.go:16` | `flow_report` |
| `echo/internal/application/tool_registry.go:152` | `flow_report` |

All four are substrings of `cash_flow_*`. The lesson generalises: in this codebase neither an unanchored `flow_` nor a `\b`-anchored pattern is safe, because `\b` treats underscore as a word character. Anchor on the quote, backtick or whitespace that actually delimits a table name.

### The three real modules

| Module | Generation | Role | Contract surface |
|---|---|---|---|
| `workspaces` | onekit, hybrid | **Owner.** Registers 12 `flow/*/v1` services (`module.go:47-135`), plus 3 hand-written chi routes | `flow/{workspace,project,label,task,dashboard,insights,milestone,cycle,view,notification,setting,time}/v1` |
| `flow_insights` | legacy chi | Read-only aggregates over `flow_projects`, `flow_tasks`, `flow_milestones` | None. 3 chi routes, 2 of which duplicate `workspaces` |
| `hr_time` | legacy chi | **Consumer.** Reads `flow_tasks` and `flow_time_logs` for the import chain (`flow_adapter.go:67-68`) | 5 of its 12 routes appear in `hr/time_entry/v1` |

Three consequences.

`workspaces` is a hybrid, not clean onekit. It has `RegisterAPI` for 12 services and `RegisterRoutes` for `/flow/ws`, `/flow/ws-ticket` and `/flow/presence/today` (`module.go:137-148`). `/flow/ws` is its only anonymous route, and the reason given is sound: browsers cannot set an `Authorization` header on a WebSocket handshake, so the ticket is validated inside the handler.

Three modules claim Flow insights. `workspaces` registers `flowinsightsv1.RegisterInsightsServiceServer` (`module.go:73`) and implements all three operations with pagination (`workspaces/internal/api/insights_service.go:22,89,137`). Legacy `flow_insights` registers the same two paths as chi routes. Meanwhile onekit `insights` stubs the same operation names as `notImplemented` (`insights/internal/api/setting_service.go:102-114`). One concern, three implementations, one of them dead.

`flow_businesses` is dead. Defined at `schemas/flow/schema.sql:202-211`, referenced by zero Go, SQL or `.onk`.

### The documentation contradiction, settled

`CLAUDE.md:977` records `echo` depending on `memory` as the "only cross-module dependency". That is false, and cross-module dependency is an accepted pattern here rather than an accumulated exception:

- `hr_core` is injected into **11** modules, not 5: `hr_attendance`, `hr_leave`, `hr_overtime`, `hr_payroll`, `hr_time`, `hr_equipment`, `hr_shifts`, `benefits`, `payroll`, `staff`, `workspaces` (`cmd/api/main.go:446-451, 462, 516, 555, 563, 576, 608`). `internal/modules/CLAUDE.md` also says 5 and is also stale.
- `echo` takes roughly 70 modules in one `NewModule` call (`cmd/api/main.go:542`).
- `insights` carries roughly 40 annotated cross-schema reads spanning `hr_employees`, `hr_shift_assignments`, `hr_leave_requests`, `finance_invoices`, `pos_orders`, `crm_customers`, `erp_products`.
- Two platform-port packages exist and are wired (`financejournal`, `customer_sync`).

The answer to the question the contradiction poses: cross-module dependency is accepted, with a house style. The house style is what section 7 formalises.

---

## 4. Four-layer inventory

### 4.1 Schema layer

`internal/platform/database/schemas/flow/` holds 8 files, not the 7 usually cited: `schema.sql`, `v6_status_backlog.sql`, `v7_task_project_status_index.sql`, `v8_time_tracking.sql`, and four `zz_` files including `zz_flow_work_types_description.sql`.

**Zero declared foreign keys in `flow/`.** Every join key, within Flow and across to HR, is convention only. This is the reason a schema-level reading cannot find the seams.

Cross-domain columns that already exist:

| Column | Table | Points at | Status |
|---|---|---|---|
| `hr_category_id` | `flow_work_types` (`v8:76`) | HR payroll category | Wired end to end: SQL, repo, domain, mapper, contract (`voxie/api/flow/time/v1/models.onk:21`). Stored and echoed; no validation against an HR category table found |
| `work_day_id`, `work_type_id`, `work_mode`, `timezone`, `client_key`, `logged_by` | `hr_time_entries` (`v8:88-106`) | Flow work days and work types | Wired. Flow adds columns to HR's table rather than keeping a second store |
| `project_id`, `task_id` | `hr_time_entries` (`hr_time/schema.sql:8-9`) | `flow_projects`, `flow_tasks` | Wired both ways |
| `is_billable`, `source` | `hr_time_entries` | Flow marks lines `source='flow'` (`worktime.sql:74`), importer marks `source='flow_import'` (`hr_time/.../service.go:274`) | Wired |
| `billable_invoice_id` | `hr_time_entries` (`hr_time/schema.sql:20`) | `finance_invoices` | **Poisoned. See 4.2** |
| `hourly_rate` | `hr_employees` (`hr_core/schema.sql:91`) | The cost side of project cost | Read by `hr_payroll` (`repository.go:428`), absent from every contract |

Nine of Flow's 19 populated tables carry `business_id`, all `DEFAULT 0`. Ten do not, and are scoped transitively through `workspace_id`, `project_id` or `task_id`. That distinction is load-bearing and section 8 depends on it.

Index coverage for the joins proposed below:

- `zz_flow_tenant_indexes.sql` adds four `business_id`-leading indexes, on `flow_projects` and `flow_tasks` only. Nothing on `flow_workspaces`, `flow_time_logs` or `flow_work_types`.
- `v8:96` adds `idx_time_entries_work_day (owner_id, work_day_id)`. Owner-leading, no `business_id`.
- `hr_time_entries`' four original indexes are all `owner_id`-leading (`hr_time/schema.sql:23-26`). **There is no index on `hr_time_entries` whose leading column is `business_id`**, and `hr_time` filters exclusively by `business_id`. Every `hr_time` list query is therefore a scan or an owner-index misuse.

### 4.2 Repository layer

Two sweeps, because the `// cross-schema read from` annotation is not a reliable index of the seams. The annotation sweep finds roughly 40, nearly all in `insights`. The `sqlc.yaml` sweep finds the ones the annotation missed.

`internal/modules/workspaces/sql/sqlc.yaml` borrows three schema directories it does not own, each with a rationale comment:

```yaml
- "../../../../internal/platform/database/schemas/hr_time/schema.sql"
- "../../../../internal/platform/database/schemas/flow/v8_time_tracking.sql"
- "../../../../internal/platform/database/schemas/hr_core/schema.sql"
```

The comments record why: `hr_time_entries` is where Flow's lines live, `hr_time` must be listed before `v8` so the ALTERs have a table to alter, and `hr_employees.user_id` is the only link between a workspace member (a user id) and a work day (an employee id). This is the best-documented cross-module seam in the repo and it is the model for the rest.

Flow's cross-boundary reads, by owner:

| Query | Module | Tables spanned | Tenant filter | Location |
|---|---|---|---|---|
| `ListWorkDayLines` | workspaces | `hr_time_entries` | `owner_id` only | `worktime.sql:55` |
| `GetTimeEntryByClientKey` | workspaces | `hr_time_entries` | `owner_id` only | `worktime.sql:60` |
| `CreateWorkDayLine` | workspaces | `hr_time_entries` | writes both | `worktime.sql:70` |
| `UpdateWorkDayLine`, `DeleteWorkDayLine` | workspaces | `hr_time_entries` | `owner_id` and `business_id` | `worktime.sql:77,83` |
| `SumWorkDayLineMinutes` | workspaces | `hr_time_entries` | `owner_id` only | `worktime.sql:89` |
| `GetInsightsTimeByMember` | workspaces | `flow_time_logs`, `users`, `flow_tasks` | `workspace_id` only | `features.sql:123` |
| `GetCompletedTasksForEmployee` | hr_time | `hr_employees`, then `flow_tasks` and `flow_time_logs` | `business_id` only, no `owner_id` | `flow_adapter.go:35,61` |

`flow_adapter.go` is the only place a legacy module reads Flow's tables, and it does so with squirrel rather than sqlc, marked `// dynamic-query: squirrel`. Both of its reads carry the required annotation.

**`billable_invoice_id` is written with a timestamp, not an invoice id.** `ExportBillableEntries` computes `batchID := time.Now().UnixMilli()` (`hr_time/internal/application/service.go:290`) and `MarkEntryExported` writes it into `billable_invoice_id` (`time_entries.sql:66`). The column that should be the Flow-to-Finance join key holds epoch milliseconds. `ListApprovedBillableEntries` then uses `billable_invoice_id IS NULL` (`time_entries.sql:60`) as an "unexported" flag. Any future join to `finance_invoices` on that column matches nothing, silently.

**The same export has no transaction.** `MarkEntryExported` runs once per entry inside a loop that logs and `continue`s on error (`service.go:293-297`). A partial failure leaves some entries flagged exported and some not, with no rollback, and the marked ones are then invisible to `ListApprovedBillableEntries`. Billable hours can be lost with a successful HTTP 200.

### 4.3 Service layer

Five mechanisms exist in this repo. A sixth, events or projections, does not.

| Mechanism | Where | Shape |
|---|---|---|
| Cross-schema read in the consumer's repository | `insights` (roughly 40), `hr_time/flow_adapter.go`, `workspaces/worktime.sql` | Consumer's `sqlc.yaml` borrows the owner's schema; query annotated |
| Constructor injection of the owner's `*Module` | `hr_core` into 11 modules; `workspaces` takes `hr_core` and `hr_leave` | Consumer imports the owner's package. Coupling is total but visible |
| Provider interface in the owner's module | `addons/hub_addon_provider.go`, `api_keys/hub_api_keys_provider.go`, `webhooks/hub_webhooks_provider.go` | Owner ships an adapter; consumer calls `SetXProvider` (`cmd/api/main.go:726-727`). Consumer still imports the owner |
| Aggregate read module | `insights`, `flow_insights` | A module that owns no tables and joins other domains' for reporting |
| **Platform-level port** | `internal/platform/financejournal/contract.go` | DTO plus interface in `platform/`. Owner implements, consumer consumes, **neither imports the other**. Wired at `cmd/api/main.go:494` and `cmd/worker/main.go:129`, with `noopFinancePoster` as the default (`hr_payroll/.../service.go:78`) |
| **Shared platform write helper** | `internal/platform/customer_sync` | Platform package with its own `sql/` and `persistence/`, takes a `DBTX` so it enlists in the caller's transaction. Called from 9 modules' repositories |

**Events and projections do not exist.** No `EventBus`, `Publisher`, `Subscriber` or `Outbox` type anywhere in `internal/`. `internal/platform/wsbus` has `Publish` and `Subscribe` but it is a WebSocket fanout hub with a Redis backplane for realtime UI delivery: no durability, no replay, no projection tables. Introducing events would be introducing a mechanism, and section 11's not-worth-building section treats it as such.

### 4.4 Handler layer

**Population one, contract-backed.** 79 method-and-path pairs across 60 distinct paths under `voxie/api/flow/`, served by `workspaces`. Normal.

**Population two, handler with no contract.** All six of Flow's hand-registered chi routes:

| Route | Module | Verdict |
|---|---|---|
| `/flow/ws` | workspaces (`module.go:139`) | Legitimately uncontractable. WebSocket upgrade |
| `/flow/ws-ticket` | workspaces (`module.go:140`) | Contractable, POST |
| `/flow/presence/today` | workspaces (`presence.go:98`) | **The working Flow-to-HR cross-module read has no contract** |
| `/flow/insights` | flow_insights | Contract calls the same thing `/flow/overview` |
| `/flow/team/workload` | flow_insights | **Duplicate.** Also served by `workspaces` |
| `/flow/roadmap` | flow_insights | **Duplicate.** Also served by `workspaces` |

`hr_time` registers 12 routes (`hr_time/internal/routes/routes.go`) against 5 declared operations in `hr/time_entry/v1/service.onk`. The 7 with no contract include both Flow chain endpoints. Both are **POST**: `routes.go:49` for import-flow, `routes.go:52` for export-billable. The prior drift document records export-billable as a GET; it is not.

**Population three, contract with no handler.** Worse than it has been reported.

- `insights/internal/api/setting_service.go:102,106,110,114`: four `notImplemented`, confirmed exactly. Plus two more in `report_service.go:73,77`, and further instances in `business/` and `hr_dashboard/`.
- `hr.work_schedule.v1.WorkScheduleService` is declared in `voxie/api/hr/work_schedule/v1/service.onk` and **has zero implementors anywhere in the repo**. No module registers it. This is strictly worse than `notImplemented`: the route 404s while the contract advertises it, and `contracted_hours_per_week` (`models.onk:20,50,74`) is unreachable as a result. The contract also declares only create and update, so there is no read operation to implement even if a module were wired.

**Caller liveness for the Flow chains.** Neither `/hr/time-entries/import-flow` nor `/hr/time-entries/export-billable` appears in any `.onk`, so no generated client can reach them, and the frontend audit found no caller. Both handlers are complete, tenant-scoped through `tenant.ResolveBusinessID`, and use `onekiterrors` correctly. They are not stubs. They are finished features with no way in.

---

## 5. Gap analysis by class

### Class A: wired and working

| Link | Mechanism | Evidence | Copy or reject |
|---|---|---|---|
| Flow work-day lines stored as HR time entries | Cross-schema read plus schema extension | `flow/v8_time_tracking.sql:88-106`, `workspaces/sql/sqlc.yaml:8-20`, `worktime.sql:54-90` | **Copy.** Best-documented seam in the repo |
| Flow presence reads HR directory and approved leave | Constructor injection | `cmd/api/main.go:462`, `workspaces/presence.go:52-74` | Copy with care. Works, but couples `workspaces` to two HR packages and the endpoint has no contract |
| Flow work types map onto HR payroll categories | Column plus contract field | `v8:76`, `worktime.sql:107`, `voxie/api/flow/time/v1/models.onk:21` | Copy. Note the id is never validated against an HR category |
| HR payroll posts totals to the finance journal | **Platform-level port** | `platform/financejournal/contract.go`, `journal/module.go:27`, `cmd/api/main.go:494` | **Copy this above all others.** Cross-generation, no module-to-module import |
| CRM customer upsert from 9 modules | Shared platform write helper | `platform/customer_sync/customer_sync.go:30` | Copy for shared writes only |

### Class B: data exists, nothing exposes it

| Item | Evidence | Cost |
|---|---|---|
| `hr_employees.hourly_rate` is read and converted to cents by `hr_payroll` but appears in no contract | `hr_core/schema.sql:91`, `hr_payroll/.../repository.go:428`, `service.go:215`; zero hits in `voxie/api` | Contract addition plus one read path. Cheapest real work in this audit |
| `hr_time_entries.project_id` plus `is_billable` plus `duration_minutes` already carry everything needed for unbilled billable hours per project | `hr_time/schema.sql:7,14,10` | One aggregate query. No migration |
| `flow_work_types.hr_category_id` is stored but never joined to an HR category | `v8:76` and no join anywhere | Either validate it or document that it is advisory |

### Class C: endpoint exists, column or field does not

| Item | Evidence |
|---|---|
| `billable_invoice_id` exists but holds a timestamp, so the Flow-to-Finance link needs a real invoice id column or a corrected write | `service.go:290`, `time_entries.sql:66` |
| `hr/time_entry/v1` has no date-range, project or billability filter, so any consumer must over-fetch | `voxie/api/hr/time_entry/v1/service.onk` declares 5 operations, none filtered |

### Class D: concept does not exist

| Item | What is missing |
|---|---|
| Project cost as a stored or computed value | No table, no column, no service. Section 6 |
| Account profitability for CRM | No link from a Flow project to a CRM customer at all. `flow_projects` has no `customer_id` |
| A read path for contracted hours | `WorkScheduleService` has no implementor and no read operation |

### Class E: drift

| Item | Evidence | Severity |
|---|---|---|
| `/flow/team/workload` and `/flow/roadmap` served twice, in two generations | `flow_insights/.../handlers/insights_handlers.go:19-20` versus `workspaces/internal/api/insights_service.go:89,137` | High. The legacy chi route is more specific than the onekit `/*` mount (`MountOneKit` does `router.Handle("/*", mux)`, `onekit_routes.go:18`), so the paginated onekit implementations are unreachable. Inference from chi's trie preferring static segments; a live request would settle it |
| `hr.work_schedule.v1` declared, zero implementors | `voxie/api/hr/work_schedule/v1/service.onk` | High. Advertised and 404s |
| 6 Flow chi routes with no contract, including the one cross-module read that works | Section 4.4 | Medium |
| 7 of `hr_time`'s 12 routes have no contract | `routes.go` versus `service.onk` | Medium |
| 4 `notImplemented` in `insights/internal/api/setting_service.go`, plus 2 in `report_service.go` | `:102,106,110,114` and `:73,77` | Medium |
| `flow_businesses` table defined, zero references | `schemas/flow/schema.sql:202` | Low |
| `migrate.sh` records a failed migration as applied | `scripts/migrate.sh:76-78`: `mysql_cmd < "$f" 2>&1 \| grep -v "Warning:" \|\| true` followed by unconditional `record_applied` | **High.** A failed file is marked done, never retried, and the script prints success |

---

## 6. The worked example: labour cost of a Flow project

### The join path that exists today

```
flow_projects.id
  -> hr_time_entries.project_id            convention only, no FK
     hr_time_entries.duration_minutes      the hours
     hr_time_entries.is_billable           billable split
     hr_time_entries.employee_id
       -> hr_employees.id
          hr_employees.hourly_rate         DECIMAL(10,2), hr_core/schema.sql:91
          hr_employees.user_id             the only bridge to a Flow member
```

Every column needed is already on disk. There is no migration in the cost calculation itself.

### What actually blocks it

Not the plumbing. Three things, in order of severity.

**The tenancy key does not agree with itself.** For the same tenant:

| Row written by | `owner_id` holds | `business_id` holds | Evidence |
|---|---|---|---|
| `workspaces` work-day line | resolved business **owner user id** | resolved business id | `worktime_service.go:62` calls `tenant.ResolveOwnerID` |
| `hr_time` time entry | the **business id** | the business id | `hr_time/.../repository.go:90`: `OwnerID: businessID` |

`tenant.ResolveOwnerID`'s own doc comment states it "collapses every business of an owner down to one shared owner id and is reserved for billing aggregation" (`platform/tenant/owner.go:57-59`). So a query filtering `hr_time_entries.owner_id` either sees Flow's lines or HR's lines, never both, and if it sees Flow's it may span several businesses of one owner. A labour-cost aggregate must filter `business_id` and must not touch `owner_id`. Confirming this empirically was not possible: `onion_dev` holds 12 `hr_time_entries` rows, all `owner_id = business_id = 1001` and all `source` blank or `manual`, with no Flow-written rows at all. The finding is static, and the static evidence is unambiguous.

**There is no index for the query.** `hr_time_entries` has four indexes, all `owner_id`-leading (`hr_time/schema.sql:23-26`), plus `v8`'s `(owner_id, work_day_id)`. A cost aggregate filtering `business_id` and grouping by `project_id` has no usable index. This needs one migration, and it belongs to `hr_time`'s schema, not Flow's.

**`hourly_rate` is not in any contract.** `hr_payroll` already reads it and converts to cents (`repository.go:428`). Nothing exposes it. This is the only genuinely missing piece, and it is small.

### Where each piece lands

| Piece | Layer | Module |
|---|---|---|
| `LabourCostProvider` port: DTO plus interface | `internal/platform/labourcost/contract.go` | platform |
| Implementation: rate lookup plus minutes aggregate | `infrastructure/` | `hr_payroll` (owns the rate reading already) |
| Wiring | `cmd/api/main.go` and `cmd/worker/main.go` | main |
| Consumption for project cost | `application/` | `workspaces` |
| Consumption for billing and P&L | `application/` | `invoices`, `finance_reports` |

### Tenancy filter, stated at every layer

- Handler: `tenant.ResolveBusinessID` only. It errors rather than returning 0 (`owner.go:74-84`), so the zero bucket is unreachable from a request.
- Service: pass `businessID` through unchanged. Never call `ResolveOwnerID` on this path.
- Repository: `WHERE hr_time_entries.business_id = ? AND hr_employees.business_id = ?` on both sides of the join. No `owner_id` anywhere.
- Index: add `(business_id, project_id, date)` on `hr_time_entries` before shipping the aggregate.

### The CRM leg does not exist

Finance can consume project cost as soon as the above lands. CRM cannot: `flow_projects` has no `customer_id` and nothing links a Flow project to a CRM customer. Account profitability is Class D and needs a data-model decision, not an endpoint. That decision is flagged in section 10.

### OneKit wire format

Every id-carrying `int64` in a new contract crosses the wire as a **string**. A JSON number fails to unmarshal and `"0"` fails `@gt(0)`. This applies to `project_id`, `employee_id`, `business_id` and, critically, to any money field expressed as cents. `hr_payroll` already uses `int64` cents through `internal/platform/money`; a `labour_cost_cents: int64` contract field will be a quoted string on the wire and every consumer must be told so.

---

## 7. Mechanism design decisions

### The decision rule

The five-way menu collapses to three questions.

1. **Is the consumer reading, or writing someone else's record?** Writing means the shared platform write helper (`customer_sync`) or nothing. Reading continues.
2. **Does the consumer need business logic the owner enforces, or just columns?** Just columns, and the owner's schema is stable: cross-schema read. Logic, ownership rules, or a value the owner may change how it computes: a port.
3. **Do owner and consumer sit in different generations?** If yes, a **platform-level port** is the only option that does not couple a legacy module to an onekit one or vice versa.

### Per link

**Project labour cost, HR to Flow, Finance and Insights.**
Owner: `hr_payroll`. It already owns rate reading and the cents conversion, and rate is a business rule the company configures, not a column Flow should interpret.
Mechanism: **platform-level port**, `internal/platform/labourcost`.
Rejected: cross-schema read in `workspaces`, because the rate is a business rule and three consumers would each reimplement it. Constructor injection of `hr_payroll` into `workspaces`, because `hr_payroll` is legacy chi and `workspaces` is onekit, so the import drags a generation boundary into Flow's module graph. Provider interface in `hr_payroll`, because it still forces `workspaces` to import `hr_payroll`. Aggregate read module, because three different consumers want three different shapes. Events, because none exist.
Join key: `hr_time_entries.project_id` to `flow_projects.id`; `hr_time_entries.employee_id` to `hr_employees.id`.
Live, not projected. Cost changes whenever an hour is logged or a rate edited, and a stale projection on money is worse than a slower read.
Tenant filter: `business_id` on both sides. Never `owner_id`.
**Cross-generation note: no conversion required.** The port lives in `platform/`, so neither side changes generation. That is the whole reason to prefer it.

**Unbilled billable hours, Flow to Finance.**
Owner: `hr_time`, because the rows are `hr_time_entries`.
Mechanism: **cross-schema read in `invoices`' repository**, not a port. This one is columns, not logic: `WHERE business_id = ? AND is_billable AND status = 'approved' AND <not yet invoiced>`.
Rejected: a port, because there is no business rule to hide and a port would add indirection for a single consumer. Reusing `export-billable`, because it mutates and its batch id is a timestamp.
Prerequisite: fix `billable_invoice_id` first. Building on a column that holds epoch milliseconds guarantees a wrong answer. **This is two pieces of work, and the fix must land first.**
Tenant filter: `business_id`.

**Approved leave gates assignment, HR to Flow.**
Already built. `workspaces/presence.go:74` calls `m.hrLeave.ApprovedLeaveEmployeeIDs`. The gap is that it is used for presence and not for assignment, and that `/flow/presence/today` has no contract. Mechanism: keep the existing injection, add the contract entry. No new mechanism.

**Contracted hours for capacity, HR to Flow.**
Blocked, and not by Flow. `WorkScheduleService` has no implementor and declares no read operation. Two pieces of work: contract addition for a read, then an implementing module. Only then can Flow stop hardcoding 40 hours.

**Flow insights consolidation.**
Not a new link. Delete `flow_insights` and let `workspaces` serve the contract it already implements with pagination. See section 10.

### The hard constraint, restated

SQL stays in `infrastructure/`. Every mechanism above puts its query there. A port's interface lives in `platform/`, its implementation's SQL in the owner's `infrastructure/`, and the consumer calls the port from `application/`. No join in `application/` or `api/`, however convenient.

---

## 8. Tenancy assessment

### The shape of the problem

`business_id BIGINT NOT NULL DEFAULT 0` on 146 of 170 `business_id` columns in `onion_dev`, confirmed by `information_schema`. 20 more are `NOT NULL` with no default, 4 are nullable. `flow/v8_time_tracking.sql:9-12` states the reasoning for the newer tables explicitly and does not add to the problem: every table it creates is `NOT NULL` with no default.

The platform helper is sound. `tenant.ResolveBusinessID` returns `ErrNoBusinessScope` rather than 0, and documents why (`owner.go:74-84`). **A request cannot write or read the zero bucket through it.** That single fact is what separates the live defects below from the residue.

### Flow's real authorization model

Flow does not gate on `business_id`. It gates on **workspace ownership or membership**, in SQL, via `EXISTS` subqueries:

```sql
AND (
  EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_views.workspace_id AND w.owner_id = ?)
  OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_views.workspace_id AND wm.user_id = ?)
)
```

That is `UpdateView` (`features.sql:81-85`), `DeleteView` (`features.sql:90-94`) and `RecalcActualHours` (`helpers.sql:30-34`). It is coherent and it explains why 10 of 19 Flow tables carry no `business_id` at all: they are reached only through a gated parent.

The consequence for this audit is the important part. **A cross-module consumer does not pass through the membership gate.** `hr_time`, `insights` and any new port implementation filter `business_id` and nothing else. So for cross-module reads, `business_id` must be correct, and on the Flow side it frequently is not.

### The three causes, distinguished

The brief asked for one INSERT traced to settle whether the zero rows are fixture residue or a live path. The answer differs per table, and that is the finding.

**`flow_workspaces`: live, handler-reachable.**

```sql
-- members.sql:29
INSERT INTO flow_workspaces (owner_id, name, icon, color, is_default, is_personal, created_by)
VALUES (?, 'Personal', '🏠', '#64C5B9', TRUE, TRUE, ?);
```

No `business_id`, and the table defaults it to 0. Three reachable entry points:

1. `cmd/api/main.go:729-733`: bound to **both** `user.OnUserCreated` and `user.OnUserLogin`. Every registration and every login.
2. `application/service.go:203`: `ResolveWorkspaceID` when the frontend sends the sentinel string `"personal"`.
3. `application/service.go:1218`: `EnsureMyTasksProject` with `workspaceID == 0`.

Chain: `main.go:731` to `module.go:152` to `service.go:192` to `repository.go:1289` to `members.sql:29`. Data agrees: 679 rows in `onion_dev`, **531 with `business_id = 0`, spanning 7 distinct `owner_id` values**, matching the previously reported figures exactly.

Effect: personal workspaces are invisible to `GetWorkspacesByOwner`, which filters `business_id = ?` (`workspaces.sql:11`) and can never be passed 0. They are reachable only through `GetPersonalWorkspaceByOwner`, which filters `owner_id = ?` (`members.sql:24`) where `owner_id` holds the **user's own id** (`repository.go:1290`: `OwnerID: userID`). So the real exposure is not tenant zero: it is that `owner_id` on this table is a user id, a third meaning distinct from both `hr_time_entries` conventions.

**`flow_views`: live, and universal.**

```sql
-- features.sql:75
INSERT INTO flow_views (workspace_id, name, icon, filters, group_by, sort_by, sort_dir, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
```

`flow_views` has `business_id DEFAULT 0`. `ViewInput` (`domain/features.go:105`) has no `BusinessID` field. Data: **55 of 55 rows are in the zero bucket.** Every view ever created. Because reads gate on membership rather than `business_id`, nothing is currently broken. What is broken is any future cross-module join that trusts the column. This is dead tenancy metadata that reads as live.

**`flow_tasks` and `flow_projects`: neither, and this is where a row count misleads.**

359 of 363 `flow_tasks` rows and 326 of 536 `flow_projects` rows sit at `business_id = 0`. But the `flow_tasks` INSERT names `business_id` (`workspaces.sql`, `CreateTask`), `domain.TaskInput` carries a `BusinessID` field (`domain/task.go:125`), and current tests populate it with `testutil.BusinessID`. `business_id` entered that INSERT in commit `6220c959` on **2026-06-19**, seven weeks before the polluted rows were created (a contiguous burst, 2026-08-08 13:09 to 19:48). No fixture file creates Flow rows: `scripts/fixtures/` contains users, businesses, chat sessions, POS, HR, ERP/CRM, config and console admin only.

So these rows were written by a caller that passed 0, after the column was being written, and not by the fixture set. Since `ResolveBusinessID` cannot yield 0, no handler produced them. The remaining explanation is direct repository-level access during a dev or integration session against `onion_dev`, which matches this database's known history as an integration-test target. **Labelled as an inference.** What would settle it: the same query against production. I did not run it and recommend someone with production read access does, because the severity difference between "dev residue" and "a live path I have not found" is the whole question.

### `*Input` structs: the hypothesis, tested

The brief flagged missing `BusinessID` on `*Input` structs as a recurring omission to grep for. Result: 14 of 30 `*Input` structs in `workspaces/internal/domain/` carry it, 16 do not. **Most of the 16 are correct**, because their target table has no `business_id` column. A naive sweep reports 11 untenanted Flow INSERTs; cross-referencing against `information_schema` shows 9 of those target tables with no such column. The genuine list is short:

| Struct | Target table | Verdict |
|---|---|---|
| `ViewInput` (`features.go:105`) | `flow_views`, has `business_id` | **Defect.** Confirmed by 55/55 |
| `UpdateViewInput` (`features.go:116`) | `flow_views` | Gates on membership instead; acceptable but leaves the column unwritten |
| `UpdateCycleInput` (`features.go:35`) | `flow_cycles`, has `business_id` | Needs checking against its query's gate |
| `UpdateMilestoneInput` (`features.go:79`) | `flow_milestones`, has `business_id` | Same |
| `MoveProjectInput` (`project.go:89`) | `flow_projects` | Gated `WHERE id = ? AND owner_id = ?` (`members.sql:66-68`). `owner_id` collapses businesses of one owner, so a project can move across them |
| `TransferProjectOwnershipInput` (`project.go:83`) | `flow_projects` | Same gate, same caveat (`members.sql:61-63`) |

I earlier read `CreateLabel` (`workspaces.sql:147`) as writing neither tenant column. `flow_labels` has neither column, so omitting them is correct. Withdrawn.

### Index coverage for the proposed joins

| Proposed join | Index present | Verdict |
|---|---|---|
| `hr_time_entries` by `business_id` and `project_id` | None. All five indexes are `owner_id`-leading | **Must add before shipping the cost aggregate** |
| `flow_projects` by `business_id` | `idx_flow_projects_business_created` (`zz_flow_tenant_indexes.sql`) | Covered |
| `flow_tasks` by `business_id` | Three business-leading indexes | Covered |
| `flow_work_types` by tenant and workspace | `idx_work_types_tenant_workspace` (`v8:78`) | Covered, though `workspace_id` holds a business id: see below |
| `flow_time_logs` by tenant | None | Add if the Finance link reads it |

### Two further defects found while here

**`flow_work_types.workspace_id` holds a business id.** `worktime_service.go:378` and `:390` both set `WorkspaceID: businessID`. Reads use the same value, so nothing is broken today, but the column name is false, per-workspace work types are impossible without a migration, and any future join of `flow_work_types.workspace_id` to `flow_workspaces.id` will silently match the wrong rows.

**`hr_time_entries` client-key idempotency is owner-scoped.** `v8:106` adds `UNIQUE KEY uk_time_entries_client_key (owner_id, client_key)`. Since `owner_id` collapses an owner's businesses, two businesses under one owner share the idempotency namespace: a key generated in business A blocks the same key in business B. The unique key should include `business_id`.

### The audit query

Run against a local or read-replica database only. `.env` in this repo sets `APP_ENV=development` with `DB_HOST=voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com` and `DB_NAME=voxire`, so a careless run hits production. State the target before querying.

```sql
SET SESSION group_concat_max_len = 1048576;  -- the default 1024 truncates silently

-- 1. Which tables carry the hazard at all.
SELECT TABLE_NAME, COLUMN_DEFAULT, IS_NULLABLE
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'business_id'
ORDER BY COLUMN_DEFAULT, TABLE_NAME;

-- 2. Prove the sweep can find a known positive before trusting any zero.
SELECT COUNT(*) AS total,
       SUM(business_id = 0) AS zero_bucket,
       COUNT(DISTINCT CASE WHEN business_id = 0 THEN owner_id END) AS owners_in_zero
FROM flow_workspaces;

-- 3. Per table, generated rather than hand-written. Inspect the output, then run it.
SELECT CONCAT(
  'SELECT ''', TABLE_NAME, ''' t, COUNT(*) n, SUM(business_id=0) z, ',
  'COUNT(DISTINCT CASE WHEN business_id=0 THEN owner_id END) owners ',
  'FROM `', TABLE_NAME, '` UNION ALL'
)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'business_id';
```

A count alone proves prevalence, never reachability. Pair every number with the INSERT that produced it.

---

## 9. Roadmap

Drift first. It is cheap, and leaving it in place corrupts every later decision: an audit that reads the contract sees three insights implementations and one unserved service, and reasons from that.

### Wave 1: stop the bleeding

| # | Item | Effort | Why first |
|---|---|---|---|
| 1 | `migrate.sh` records failed migrations as applied | Trivial | Every later migration in this roadmap is unsafe until this is fixed |
| 2 | Wrap `ExportBillableEntries` in a transaction | Small | Billable hours are currently losable on a 200 response |
| 3 | `flow_views` INSERT writes `business_id`; add `BusinessID` to `ViewInput` | Small | 55 of 55 rows wrong, and it is one column |
| 4 | `CreatePersonalWorkspace` writes `business_id` | Small, plus a backfill decision | Fires on every login |
| 5 | Delete the `flow_insights` module | Small | Removes two duplicate routes and a whole legacy module. See section 10 |

**Lands:** the tenancy columns this audit's joins depend on become trustworthy, and Flow's route surface stops contradicting its contract.

### Wave 2: make the seam legible

| # | Item | Effort |
|---|---|---|
| 6 | Contract entry for `/flow/presence/today` | Small |
| 7 | Decide the fate of `hr.work_schedule.v1`: implement or delete | Small either way |
| 8 | Fix `billable_invoice_id` to hold an invoice id, or add a separate `export_batch_id` | Medium, needs a data decision |
| 9 | Add `(business_id, project_id, date)` to `hr_time_entries` | Small |
| 10 | Resolve the import and export chains (section 10) | Medium |

**Lands:** the existing Flow-to-HR link becomes discoverable by contract, and the Flow-to-Finance join key becomes real.

### Wave 3: the links worth building

| # | Item | Mechanism | Effort |
|---|---|---|---|
| 11 | `internal/platform/labourcost` port, implemented in `hr_payroll` | Platform port | Medium |
| 12 | Project cost consumed by `workspaces` | Port consumption | Small once 11 lands |
| 13 | Unbilled billable hours read in `invoices` | Cross-schema read | Small once 8 and 9 land |
| 14 | Expose `hourly_rate` on an HR contract | Contract plus read | Small |

**Lands:** the project-cost example works end to end, and Finance can bill from Flow's hours.

### Wave 4: needs a product decision first

| # | Item | Blocker |
|---|---|---|
| 15 | Flow project to CRM customer link | No column, no concept. Section 10 |
| 16 | `flow_work_types.workspace_id` migration to real workspace scoping | Is per-workspace scoping wanted at all? |

---

## 10. Not worth building

**A projection or event mechanism.** Nothing in this repo does it. `wsbus` is a WebSocket fanout hub, not a durable bus. Introducing one to serve project cost would add a mechanism, an ordering problem and a staleness window to a number that must be right, in exchange for saving a join that has no index yet. Add the index instead.

**A separate Flow labour-cost aggregate endpoint.** `insights` already carries roughly 40 cross-schema aggregate reads and is the sanctioned home for reporting joins. A third insights implementation is what the current mess is made of.

**Completing `/hr/time-entries/import-flow` as a user-facing button.** It should not be deleted either, and the reasoning has changed. `flow_time_logs` is not historical: it is actively written (`workspaces.sql:139`), drives `flow_tasks.actual_hours` (`helpers.sql:25`) and feeds the workload insight (`features.sql:123`). So the importer reads real, current data that `hr_time_entries` does not contain, and the "already solved below the contract layer" argument does not apply to it. But a per-task importer is also the wrong shape: it collapses every time log for a task onto the task's `completed_at` (`service.go:270-271`), destroying the per-day attribution payroll needs. **Recommendation: delete the endpoint and file the real question, which is whether Flow should have two time stores at all.** That is a data-model decision, not a button.

**`/hr/time-entries/export-billable` as it stands.** Deleting it is not purely subtractive: `MarkEntryExported` writes, and `billable_invoice_id IS NULL` is load-bearing for `ListApprovedBillableEntries`. Recommendation: keep the query, delete the endpoint, and let `invoices` do the read directly once item 8 lands. The mutation belongs to invoice creation, not to an export button.

**Retrofitting `business_id` onto the 10 Flow tables that lack it.** They are gated through a parent that is itself gated. Adding a column that must then be kept consistent buys nothing over the `EXISTS` gate and adds 10 more places to forget it.

**Converting `flow_insights` to onekit.** Delete it instead. `workspaces` already implements all three of its operations, with pagination it lacks.

---

## 10a. Issues filed

All on `tryrtylr/onion`, labelled `audit:flow-backend-2026-08-21`.

| # | Title | Wave | Class |
|---|---|---|---|
| 677 | `[scripts]` migrate.sh records a failed migration as applied and reports success | 1 | E |
| 678 | `[hr_time]` export-billable marks entries exported without a transaction, losing billable hours on a 200 | 1 | E |
| 682 | `[workspaces]` CreateView writes no business_id: 55 of 55 rows sit in the tenant-zero bucket | 1 | E |
| 681 | `[workspaces]` CreatePersonalWorkspace writes no business_id, on every registration and every login | 1 | E |
| 683 | `[flow_insights]` Delete the module: workspaces already serves all three operations, with pagination | 1 | E |
| 676 | `[journal]` Payroll journal posting names and creates the tenant chart-of-accounts entries itself | 1 | A |
| 687 | `[workspaces]` /flow/presence/today has no contract, hiding the one working Flow-to-HR read | 2 | E |
| 684 | `[hr_work_schedule]` WorkScheduleService is declared in the contract with zero implementors | 2 | E |
| 679 | `[hr_time]` billable_invoice_id holds a timestamp, not an invoice id | 2 | C |
| 688 | `[hr_time]` No index on hr_time_entries leads with business_id, the only column the module filters | 2 | E |
| 685 | `[hr_time]` Delete the Flow import and billable-export endpoints, and settle the two-time-stores question | 2 | E |
| 680 | `[hr_time]` hr_time_entries.owner_id holds a business id or an owner user id depending on the writer | 2 | E |
| 694 | `[workspaces]` Flow route drift: all 6 hand-registered Flow routes have no contract entry | 2 | E |
| 690 | `[workspaces]` client_key idempotency is owner-scoped, so businesses under one owner collide | 2 | E |
| 686 | `[platform]` Add a labourcost port so Flow, Finance and Insights get one project labour cost from HR | 3 | D and B |
| 693 | `[invoices]` Finance cannot see approved billable hours Flow has logged | 3 | B |
| 689 | `[workspaces]` flow_work_types.workspace_id holds a business id | 4 | E |
| 691 | `[schemas]` flow_businesses is defined and referenced nowhere | 4 | E |
| 692 | `[docs]` CLAUDE.md calls cross-module dependency the only one; there are 11 hr_core injections and two undocumented platform mechanisms | 1 | E |

Two require a product decision before they can be scheduled: #689 (are work types per-workspace or business-wide?) and the two-time-stores question inside #685.

---

## 11. Reconciliation with the frontend audit

Read after sections 1 to 10 were written, so agreement below is corroboration rather than anchoring. Sources: `tryrtylr/rtylr` branches `origin/docs/cross-app-data-sharing-audit` (34 documents, issues #834 to #862) and `origin/docs/flow-cross-app-deep-dive`. Neither is on `rtylr`'s `main`.

A note on contract citations. Their documents cite `packages/voxie/api/...`, the vendored copy in the separate `packages` repo (rtylr's `packages` submodule is at `7596d813`). This audit cites `voxie/api/...` in this repo at `0e48b902`. I found no actual disagreement on any field either side cited, so the pointer difference did not bite here. Worth noting that the divergence described in the brief has partly closed: `54c48f0b` is an ancestor of onion's current `0e48b902`, and `d6e44d1c` is not a commit in voxie's history at all.

### Confirmed

| Their finding | My evidence | Note |
|---|---|---|
| `hr_employees.hourly_rate` exists and nothing multiplies it per project | `hr_core/schema.sql:91`; zero hits in `voxie/api` | Exact match |
| Payroll computes the money rather than storing a typed figure | `hr_payroll/.../service.go:215`, `repository.go:428` | Confirmed. They cite `:221` and `:186`, adjacent lines in the same function |
| HR must own the cost endpoint so Flow never sees the inputs | `hr.Staff` exposing `salary`, `hourly_rate`, `overtime_rate` | **Their argument is better than mine.** Section 7 justified HR ownership on "rate is a business rule"; the real reason is that Flow computing client-side would expose colleagues' pay to any workspace member. Adopted |
| `/hr/time-entries/export-billable` is a POST, not a GET | `hr_time/internal/routes/routes.go:52` | Confirmed. Their drift document has it as a GET |
| `flow_workspaces`: 531 rows across 7 tenants (rtylr#871) | `onion_dev`: 679 rows, 531 at `business_id = 0`, 7 distinct owners | Exact match |
| 146 of 170 `business_id` columns default to 0 (rtylr#871) | `information_schema` on `onion_dev` | Exact match |
| 109 of 227 hand-registered routes have no `.onk` (rtylr#873) | 238 path literals inside `*Routes()` bodies | Consistent. The "93 of 195" restatement in my brief was stale; their own number is close to mine |
| Four `notImplemented` in `insights/internal/api/setting_service.go` (rtylr#870) | `:102,106,110,114` | Exact. Population is larger: `report_service.go:73,77` plus `business/` and `hr_dashboard/` |
| `contracted_hours_per_week` has no read RPC (rtylr#868) | `voxie/api/hr/work_schedule/v1/service.onk` declares only create and update | Confirmed **and worse**: the service has zero implementors anywhere, so even those two 404 |

### Contradicted

**HR-FIN-01 is already shipped, and their reason for not shipping it describes a defect that shipped with it.** Their document costs it at "backend 0 days for the read and write paths; both exist", frames the status quo as "a bookkeeper reads one and types the other", and argues against automation: "Posting automatically would require RTYLR to choose a tenant's accounts, which the configurability rules forbid."

The automation exists. `hr_payroll/.../service.go:386` calls `s.financePoster.PostPayrollJournal` on run completion, through the port at `platform/financejournal/contract.go`, implemented at `journal/module.go:27`, wired in both binaries (`cmd/api/main.go:494`, `cmd/worker/main.go:129`). The posting is transactional and idempotent on `(business_id, source_type, source_id)` with `FOR UPDATE` (`journal/.../repository.go:161-195`).

And their objection is correct about the shipped code. `repository.go:200-208`:

```go
expenseAccountID, err := r.ensurePayrollAccount(ctx, tx, input, "Payroll Expense", "expense")
payableAccountID, err := r.ensurePayrollAccount(ctx, tx, input, "Payroll Payable", "liability")
```

RTYLR names and creates the accounts. A chart of accounts is a tenant business rule, so this is a configurability violation of the kind CLAUDE.md section 8 rates as severe as a missing tenant filter. Their reasoning was right and their facts were wrong; my facts were right and their reasoning supplies the severity. rtylr#854 should be reframed from "gap" to "already automated, and the account mapping is hardcoded".

**The six-module Flow ownership map is wrong.** Three modules. Section 3.

**rtylr#869's deletion premise does not hold.** The recommendation rests on Flow's lines already being `hr_time_entries` rows, making an importer redundant. True for work-day lines, false for task time logs: `flow_time_logs` is actively written (`workspaces.sql:139`), drives `flow_tasks.actual_hours` (`helpers.sql:25`) and feeds the workload insight (`features.sql:123`). The named falsifier, "unmigrated historical `flow_time_logs`", is not historical. I still recommend deleting the endpoint, but for a different reason: the importer collapses every log for a task onto its `completed_at` (`service.go:270`), destroying the per-day attribution payroll needs. Section 10.

**HR-FLOW-01's shape is necessary but not sufficient.** "Shape 2, a backend aggregate endpoint owned by HR" is right about ownership and about there being one endpoint for Flow and Insights. It does not mention that the aggregate cannot currently be written correctly, because `hr_time_entries.owner_id` holds two different kinds of value depending on which module wrote the row. Their 4-day backend estimate omits that fix and the missing index. Resolution: build both, with the platform port as the in-process implementation behind the contract, so Flow avoids an HTTP hop and the OneKit int64-as-string round trip while Insights still gets a contract.

### New

Not present in their documents.

| Finding | Location |
|---|---|
| `owner_id` carries three incompatible meanings across the Flow-to-HR path, two of them in one table. The actual blocker on project cost | `worktime_service.go:62`, `hr_time/.../repository.go:90`, `workspaces/.../repository.go:1290` |
| Two live Flow time stores | `workspaces.sql:139` versus `worktime.sql:70` |
| The platform-port mechanism, a sixth option better than the five-way menu, already proven cross-generation | `platform/financejournal/contract.go` |
| The shared platform write helper, used by 9 modules' repositories | `platform/customer_sync/customer_sync.go:30` |
| `billable_invoice_id` holds `time.Now().UnixMilli()`, poisoning the Flow-to-Finance join key | `service.go:290`, `time_entries.sql:66` |
| `ExportBillableEntries` has no transaction; billable hours are losable on a 200 | `service.go:293-297` |
| `flow_views`: 55 of 55 rows in the zero bucket, `ViewInput` has no `BusinessID` | `features.sql:75`, `domain/features.go:105` |
| rtylr#871's open question settled: the `flow_workspaces` path is handler-reachable on every registration **and every login** | `cmd/api/main.go:729-733` |
| rtylr#871's row counts conflate three causes. `flow_tasks`' 359 zero rows are neither fixture residue nor handler-reachable | `6220c959` predates the rows by seven weeks; `scripts/fixtures/` creates no Flow rows |
| `migrate.sh` records a failed migration as applied and reports success | `scripts/migrate.sh:76-78` |
| `hr.work_schedule.v1` has zero implementors | `voxie/api/hr/work_schedule/v1/service.onk` |
| `/flow/roadmap` and `/flow/team/workload` served twice, across two generations, with the paginated implementation shadowed | `flow_insights/.../insights_handlers.go:19-20` versus `workspaces/.../insights_service.go:89,137` |
| `flow_work_types.workspace_id` holds a business id | `worktime_service.go:378,390` |
| `client_key` idempotency is owner-scoped, so businesses under one owner share the namespace | `v8:106` |
| No index on `hr_time_entries` leads with `business_id`, which is the only column `hr_time` filters | `hr_time/schema.sql:23-26` |
| Zero declared foreign keys in `flow/` | `schemas/flow/` |
| `hr_core` is injected into 11 modules, not 5. Both CLAUDE.md files understate it | `cmd/api/main.go:446-608` |
| `flow_businesses` is a dead table | `schemas/flow/schema.sql:202` |

### Their proposals, assessed

| Issue | Verdict |
|---|---|
| rtylr#854 HR to FIN payroll journal | **Already shipped.** Reframe as a configurability defect in the account mapping |
| rtylr#838 HR to FLOW leave gates assignment | **Mechanism already wired.** `presence.go:74` calls `ApprovedLeaveEmployeeIDs`. Only the application to assignment is new, and the endpoint needs a contract |
| rtylr#837 and #862 HR labour cost | **Buildable, blocked.** Fix the `owner_id` convention and add the index first. Two prerequisites, not zero |
| rtylr#836 FLOW to FIN unbilled billable hours | **Buildable, blocked.** `billable_invoice_id` must stop holding a timestamp first. Two pieces of work |
| rtylr#868 contracted hours for capacity | **Blocked on a service with no implementor.** Two pieces of work: a read operation in the contract, then a module to serve it |
| rtylr#867 Flow re-asks the hourly rate | **Confirmed, and their permissions argument makes it more serious than "debt".** A per-browser copy of a pay rate is the leak their own HR-FLOW-01 doc argues against |
| rtylr#872 Flow has no currency | Not assessed here. `internal/platform/money` exists and `hr_payroll` uses `int64` cents; any Flow money field should follow it |
| rtylr#870 four notImplemented | **Confirmed, extend to 6 in `insights` plus other modules.** Their rejection of an Insights aggregate as "already partly built" does not survive: all four of those handlers return `notImplemented`, and the working implementation is in legacy `flow_insights`, which section 10 recommends deleting |

### Repo moves recommended

All five are onion findings on the frontend tracker. File the onion counterpart, cross-link, and narrow the rtylr issue to the frontend half where one genuinely exists.

| rtylr issue | Recommendation |
|---|---|
| #869 import chain | **Close as moved.** Wholly backend. The onion issue carries the corrected premise |
| #871 tenancy | **Narrow.** The reachability answer and the three-cause split are backend. No frontend half |
| #873 route drift | **Narrow to the frontend half**, which is the client-side consequence of calling uncontracted routes. The count and the fix are backend |
| #870 notImplemented | **Close as moved.** Wholly backend |
| #862 aggregate endpoints | **Keep open, add the backend prerequisites.** The frontend half (consuming panels) is real |
