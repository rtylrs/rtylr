# Flow cross-app data sharing: what is shared, what could be, what should not be

An audit of every data exchange between Flow and its sibling apps, in both
directions, plus a design for the exchanges worth building.

Companion to `docs/flow-time-tracking.md`, which records the contract gaps inside
Flow's own time domain. This document looks outward instead: at the seams between
Flow and HR, Finance, CRM, ERP, Insights and Hub.

Audited 2026-08-21 against `packages` submodule `7596d813` and `packages/voxie`
`0e48b902`, with backend claims checked against onion `ccf239fc`.

---

## 1. Executive summary

Five findings, in the order that should change what gets built.

**Finding 1: the Flow-to-HR hours seam already exists, one layer below the
contracts, and no audit reading only `.onk` files would find it.**
`~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:4-8` states the
first half plainly: "The lines themselves are hr_time_entries." It used to go on
to say "Flow does not keep a second store of hours", which was false and has been
corrected (#890, tryrtylr/schemas#39). `Repository.CreateWorkDayLine`
(`~/Development/onion/internal/modules/workspaces/internal/infrastructure/worktime_repository.go:219-255`)
writes into HR's time-entry table via `GetTimeEntryByClientKey` and
`CreateWorkDayLine` sqlc queries. `flow_work_day_lines` is not a table; the
`flow` schema creates only `flow_work_days`, `flow_work_day_actions` and
`flow_work_types` (`~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:14,49,65`). The work-day seam
is real and nothing needs building for it.

**Correction, from #889 and #890: the "single store" half of this finding was
wrong.** A Flow-side hours table already exists. `flow_time_logs`
(`~/Development/onion/internal/platform/database/schemas/flow/schema.sql:127`) holds per-task durations written by
`flow.task.v1 CreateTimeLog`, which the task detail time panel calls on a live
route. So payroll reads one store, but Flow writes two, and this document's
advice to treat a Flow-side hours table as a regression described a table that
was already there. What is actually wrong is narrower and is filed as **#903**:
`RecalcActualHours` and `GetInsightsTimeByMember` aggregate `flow_time_logs`
alone, so a work-day line carrying a `task_id` never reaches that task's
`actual_hours` or the insights panel. Do not resolve that by summing both stores
without reading #903 first; the same hour can legitimately be in each.

**Finding 2: two live cross-app endpoints exist, are absent from the contracts,
and are called by nobody who benefits.** `POST /hr/time-entries/import-flow`
(`the import route (deleted from onion by #701, "Delete the flow time import chain")`)
and `POST /hr/time-entries/export-billable`
(`~/Development/onion/internal/modules/hr_time/internal/routes/handlers/export_billable_handler.go:15`) are both registered and implemented. Neither
appears anywhere in `packages/voxie/api`. The import has exactly one frontend
caller, `useImportFromFlow` in `apps/hr/src/shared/hooks/queries/useTimeEntries.ts:86-114`,
which has **zero call sites** and whose declared response type does not match what
the handler returns. Driven under vitest against the handler's real response
shape, that hook renders the toast `Imported [object Object],[object Object]
entries`. The billable export has no caller at all. The platform built the pipes
and never connected them.

Neither is an isolated omission: **109 of onion's 227 hand-registered route
literals have no matching path in any `.onk` file**, so nearly half the
hand-written API is outside the contracts and `pnpm contracts:audit-requests`
cannot see it. That is how the response-shape disagreement above survived. Filed
as #873.

**Finding 3: Flow already asks the user for the hourly rate that HR owns, stores
it in `localStorage`, and reads it back nowhere.**
`apps/flow/src/shared/stores/timeSettingsStore.ts:50` declares
`ratesByUserId: Record<string, number | null>`, written from a full
"Billable rates, one per person" settings UI at
`apps/flow/src/features/settings/sections/TimeSection.tsx:409-462`. Grepping every
`.ts` and `.tsx` in `apps/flow/src` for `ratesByUserId` returns the store and that
one screen. Meanwhile `packages/voxie/api/hr/staff/v1/models.onk:17-19` carries `salary`,
`hourly_rate` and `overtime_rate`, and `packages/voxie/api/hr/employee/v1/models.onk:28-32` carries
`hourly_rate`, `monthly_salary` and an ISO-4217 `currency`. This is the audit's
purest case of one app re-asking for what another already knows, made worse by
the answer being discarded.

**Finding 4: Flow has no money and no customer, anywhere.** A grep of all 24 Flow
`.onk` files for budget, cost, billable, customer, client, invoice, rate, price,
currency, revenue and margin returns exactly three real hits: `is_billable` on
`WorkDayLine` (`packages/voxie/api/flow/time/v1/models.onk:106,207,223`), two `client_key`
idempotency fields (`:111,197`), and an unrelated `completion_rate`
(`packages/voxie/api/flow/insights/v1/models.onk:7`). `flow.project.v1.Project`
(`packages/voxie/api/flow/project/v1/models.onk:3-38`) has no `customer_id` and no `budget`. Flow can
mark an hour billable. It cannot say what that hour costs, what it is worth, or
who to bill. Symmetrically, `project_id` and `cost_center` appear **zero** times
across the entire `erp/` and `finance/` contract domains, so no purchase or
expense can be attributed to a Flow project either.

**Finding 5: no app outside Flow reads any Flow data through the frontend.** A
grep of `apps/` excluding `apps/flow/` for `/flow/` endpoint literals and
`voxie/flow/` imports returns nothing. Meanwhile `insights.dashboard.v1` already
defines a complete Flow-shaped projection, `GetInsightsResponse` with
`avg_cycle_time_days` and `velocity`, `RoadmapProject`, and `WorkloadEntry`
(`packages/voxie/api/insights/dashboard/v1/models.onk:3-99`), and every one of its four handlers
returns `notImplemented`
(`~/Development/onion/internal/modules/insights/internal/api/setting_service.go:102,106,110,114`).
The contract for sharing Flow's delivery data platform-wide was written and then
abandoned on both sides.

### The shape of the work

Flow's cross-app problem is not that it is isolated. It reads HR staff,
departments, leave and employee identity today, and its hours already live in
HR's table. The problem is that **the exchange stops exactly where money starts**.
Every link that ends in a number a business would act on, a project's cost, a
client's profitability, an invoice line, is missing, and in two places Flow has
built a local imitation of it instead.

### Decisions already taken

Two questions were put to the product owner before the design was written, because
both change the shape of the roadmap rather than one row of it.

- **Flow serves both internal teams and client delivery**, so `customer_id` on a
  project is nullable and the client-facing UI is gated on a workspace setting.
- **Cost is server-computed only.** Flow never receives an hourly rate. onion
  joins hours to rates and returns a money total, so no individual's pay is
  derivable from a Flow API response.

The second decision has precedent in this codebase and section 7 builds on it.

---

## 2. Coverage statement

### What was swept

`apps/flow/src`, all 188 non-test `.ts` and `.tsx` files, in full for the API
layer and targeted for the UI layer:

- `shared/services/flowApi.ts`, all 900 lines. This is the real API seam: 70 of
  Flow's 120 `safe.*` calls live in this one file.
- All 27 feature-local `use*.ts` hooks under `features/*/`.
- All 3 files in `shared/hooks/queries/`, plus `shared/hooks/useMyEmployeeId.ts`.
- `shared/lib/api.ts`, `router.tsx` (364 lines), `layouts/FlowLayout.tsx`,
  `shared/stores/timeSettingsStore.ts`.
- The 18 files under `features/` containing a `<form`, `Dialog` or `Modal`, and
  every `type="number"` input (7 sites), to catch findings that issue no request.

Counterpart surfaces:

- All 12 contract domains in `packages/voxie/api`, service-level. Message-level
  reads for hr (54 files), finance (28), erp (24), crm (16), insights (16),
  hub (48), flow (24).
- `apps/hr`, `apps/finance`, `apps/insights`, `apps/crm`, `apps/recruit` for
  their Flow-adjacent call sites.
- onion `ccf239fc`: `internal/modules/hr_time`, `internal/modules/workspaces`,
  `internal/modules/insights`, and `internal/platform/database/schemas/`.

### Which idioms were grepped

CLAUDE.md documents one API idiom. Flow uses three, and counting only the
documented one undercounts by more than half.

| Idiom                                                   | Call sites in Flow | Where                                                                                                                   |
| ------------------------------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| `safe.get/post/put/patch/delete` via `safeFetch`        | 120                | 70 in `shared/services/flowApi.ts`, the rest in `useMyEmployeeId.ts`, `router.tsx`, `useProfile.ts`, `usePresence.ts`   |
| Generated OneKit clients from `@voxire/voxie/**/client` | 9 imports          | 6 files: `useDirectory.ts`, `useWorkDay.ts`, `useProfile.ts`, `useLeave.ts`, `useFlowTeam.ts`, `useFlowSubscription.ts` |
| Raw `axios`                                             | 0                  | one mention in `apps/flow/src/layouts/FlowLayout.test.tsx:73`, a comment, not a call                                    |

A related correction to `CLAUDE.md`: it describes query hooks as living in
`shared/hooks/queries/`. In Flow that directory holds 3 files. The live pattern is
a feature-local `use<Feature>.ts` beside the screen, plus the `flowApi.ts`
monolith.

### What was deliberately skipped, and why

- **`pos` (58 files), `booking` (14), `support` (10), `console` (2).** Endpoint
  scan only, no UI sweep. Their apps are not in this repo, so a UI-level finding
  could not be verified to a `file:line` and would be speculation.
- **`apps/menu`.** Grepped, not read. It is the public unauthenticated menu, with
  no tenant-scoped operator surface to exchange with Flow.
- **`apps/dash` and `apps/auth`.** Swept for Flow references only, not audited in
  full. Their exchange with Flow is Hub-mediated (team, permissions, subscription)
  and already class A.
- **Row-level SQL verification of every proposed join.** Section 7 states the join
  keys and the tenancy column, verified against the schema files. It does not
  claim the queries were executed.
- **Any query against a production database, and any call to a live endpoint.**
  `onion/.env` points `DB_HOST` at an `*.rds.amazonaws.com` host with
  `DB_NAME=voxire` while `APP_ENV=development`, and onion has no `godotenv`, so
  `resolveMainDSN` (`~/Development/onion/internal/platform/database/dbclient/mysql.go:198-211`)
  resolves to `localhost:3306/onion_dev` from a clean shell and to that RDS
  instance one `source .env` away. The tenancy figures in section 7.4 come from
  the local `onion_dev` container only, confirmed at the connection, read-only.
  No `insights` or `hr/time-entries` endpoint was called anywhere.
- **One thing was verified at runtime**, unlike the first round: the toast in row
  O3, driven under vitest with the API stubbed and i18next initialised the way
  `apps/hr/src/shared/services/i18next.ts:5-21` initialises it. No server involved.

### A method note on absence proofs

Three measurements in this round returned a confidently wrong number rather than
an error. All three are the same failure: **a check whose failure mode is a quiet
under-report, which is indistinguishable from a pass.** They are recorded here
because the next person measuring anything in this repo will hit one.

1. **`GROUP_CONCAT` truncates silently.** `group_concat_max_len` defaults to 1024.
   The tenancy audit query in section 7.4 generates one `SELECT` per table; at the
   default it emitted 0 of 145 branches and returned an empty result set, which
   reads exactly like a clean bill of health. **Always `SET SESSION
group_concat_max_len` before generating SQL with it**, and check the branch
   count against the table count before trusting the output.
2. **`information_schema.COLUMNS` does not distinguish a table from a view.** The
   same query counted `product_stock_total`, which is a view
   (`~/Development/onion/internal/platform/database/schemas/erp/migrations/002_unified_stock.sql:73`),
   double-counting rows it had already counted in `stock_levels`. Join
   `information_schema.TABLES` and filter `TABLE_TYPE = 'BASE TABLE'`.
3. **A word-boundary regex does not cross an underscore.** `\b` treats `_` as a
   word character, so `\bbillable` never matches `is_billable` and `\bsalary`
   never matches `monthly_salary`. An absence proof built on `\b` reports absence
   for every snake_case field whose name is a suffix. This is how the brief this
   round started from acquired two wrong "verified facts", corrected in section 1.

The general rule: **an absence proof needs its negative result falsified before it
is believed.** Run it against a case you know is present. Every absence claim in
section 3 and section 4 was checked that way, and section 5's class B rows name a
file, message and field precisely so a reader can falsify them cheaply.

### What is inference rather than verification

Labelled inline throughout with **(inference)**. The main ones:

- That `useImportFromFlow`'s response-type mismatch produces a wrong toast is read
  from the code, not observed in a browser. The type mismatch itself is verified.
- Effort estimates in section 8 are judgement, not measurement.
- Section 9's claims about what users do instead are reasoned from the absence of
  a UI path, not from user research.

---

## 3. Inventory, inbound: Flow as consumer

Direction is "Flow reads" for every row in this catalogue. Gap classes are defined
in section 5.

| #   | Data                               | Counterpart | Contract path, message and field                                                                                                                            | Frontend site                                                                                                                                                   | Trigger                                                                                       | Class | Consequence today                                                                  |
| --- | ---------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | ----- | ---------------------------------------------------------------------------------- |
| I1  | Caller's own `hr_employees.id`     | HR          | none, uncontracted `GET /hr/employees/me`; handler `~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:148`         | `apps/flow/src/shared/hooks/useMyEmployeeId.ts:16`                                                                                                              | Automatic on Leave screen mount                                                               | A     | Works. Precedent for id-space resolution                                           |
| I2  | Business staff roster              | HR          | `hr/staff/v1/models.onk` `Staff`, via `listStaff`                                                                                                           | `apps/flow/src/features/directory/useDirectory.ts:25`                                                                                                           | Directory route load                                                                          | A     | Works, but unpaginated                                                             |
| I3  | Department names                   | HR          | `hr/department/v1` `listDepartments`                                                                                                                        | `apps/flow/src/features/directory/useDirectory.ts:26`, `apps/flow/src/features/profile/useProfile.ts:31`                                                        | Directory and Profile load                                                                    | A     | Works                                                                              |
| I4  | Own staff record                   | HR          | `hr/staff/v1/models.onk` `Staff`; no self-lookup or `user_id` filter exists                                                                                 | `apps/flow/src/features/profile/useProfile.ts:30-33`                                                                                                            | Profile route load                                                                            | C     | Flow lists the whole roster and filters client-side. Unbounded read                |
| I5  | Own staff detail (extended)        | HR          | **none**; `GET /hr/staff/detail` is absent from contracts and absent from onion                                                                             | `the profile call site (rewritten upstream)`                                                                                                                    | Profile route load                                                                            | D     | The query cannot succeed. See issue in section 10                                  |
| I6  | Leave types                        | HR          | `hr/leave/v1` `listLeaveTypes`                                                                                                                              | `apps/flow/src/features/leave/useLeave.ts:23`                                                                                                                   | Leave screen load                                                                             | A     | Works                                                                              |
| I7  | Own leave requests                 | HR          | `hr/leave/v1` `listLeaveRequests`                                                                                                                           | `apps/flow/src/features/leave/useLeave.ts:34`                                                                                                                   | Leave screen load                                                                             | A     | Works                                                                              |
| I8  | Own leave balances                 | HR          | `hr/leave/v1` `listLeaveBalances`                                                                                                                           | `apps/flow/src/features/leave/useLeave.ts:45`                                                                                                                   | Leave screen load                                                                             | A     | Works                                                                              |
| I9  | Team membership and roles          | Hub         | `hub/team/members/v1/types` `StaffUserPayload`                                                                                                              | `apps/flow/src/shared/hooks/queries/useFlowTeam.ts:41-53`                                                                                                       | Layout mount                                                                                  | A     | Works                                                                              |
| I10 | Flow app grant                     | Hub         | `hub/team/permissions/v1`                                                                                                                                   | `apps/flow/src/shared/hooks/queries/useFlowTeam.ts:15`                                                                                                          | Team screen action                                                                            | A     | Works                                                                              |
| I11 | Subscription and feature gate      | Hub         | `hub/subscription/v1`                                                                                                                                       | `apps/flow/src/shared/hooks/queries/useFlowSubscription.ts:3`                                                                                                   | Layout mount                                                                                  | A     | Works                                                                              |
| I12 | Own user identity                  | Hub         | `hub/user/v1`                                                                                                                                               | `apps/flow/src/shared/services/flowApi.ts:890`, `router.tsx`                                                                                                    | AppGuard                                                                                      | A     | Works                                                                              |
| I13 | Business list and id               | Hub         | `packages/voxie/api/hub/business/v1/models.onk:25-34` `BusinessEntry`                                                                                       | `apps/flow/src/shared/services/flowApi.ts:432`                                                                                                                  | Workspace manager load                                                                        | A     | Works                                                                              |
| I14 | **Business currency**              | Hub         | `packages/voxie/api/hub/business/v1/models.onk:16` `Business.currency`, via `getBusiness`                                                                   | absent; would go in `shared/services/flowApi.ts` beside `getBusinesses:432`                                                                                     | Would load with business config                                                               | B     | Flow has no currency at all. Any money it renders would be unlabelled              |
| I15 | **Employee hourly rate**           | HR          | `packages/voxie/api/hr/employee/v1/models.onk:28` `hourly_rate`, `:32` `currency`; `packages/voxie/api/hr/staff/v1/models.onk:18` `hourly_rate`             | `the rate settings UI (since removed with the ratesByUserId write path)` reads `localStorage` instead                                                           | Settings, Time, Billable rates                                                                | C     | An admin retypes every person's rate into Flow, per browser, and nothing reads it  |
| I16 | **Overtime rate**                  | HR          | `packages/voxie/api/hr/staff/v1/models.onk:19` `overtime_rate`; `packages/voxie/api/hr/employee/v1/models.onk:69`                                           | absent                                                                                                                                                          | would gate the OT threshold in `apps/flow/src/features/settings/sections/TimeSection.tsx:215` | C     | Flow's OT threshold is a local number with no pay consequence                      |
| I17 | **Contracted hours per week**      | HR          | `packages/voxie/api/hr/work_schedule/v1/models.onk:20` `contracted_hours_per_week`; **no list or get RPC exists**, and `:32` says so in its own doc comment | `apps/flow/src/layouts/FlowLayout.tsx:380` hardcodes `capacity: 40`; `apps/flow/src/features/workspace-dashboard/WorkspaceDashboard.tsx:128` falls back to `40` | Layout mount                                                                                  | C     | Every workload bar in Flow measures against a fictional 40 hours                   |
| I18 | **Working days and OT thresholds** | HR          | `packages/voxie/api/hr/work_schedule/v1/models.onk:12,22,24` `working_days`, `daily_ot_threshold_minutes`, `weekly_ot_threshold_minutes`                    | `apps/flow/src/shared/stores/timeSettingsStore.ts:37,38` holds browser-local equivalents                                                                        | Settings, Time                                                                                | C     | The whole `timeSettingsStore` seam has a real HR owner it cannot reach             |
| I19 | **Approved leave for capacity**    | HR          | `hr/leave/v1` `listLeaveRequests` with status filter; already reachable                                                                                     | absent from planning screens; `features/roadmap/useRoadmap.ts`, `features/team/useTeamWorkload.ts`                                                              | Roadmap and Workload load                                                                     | B     | A sprint is planned against people who are on holiday                              |
| I20 | Attendance clock records           | HR          | `hr/attendance/v1` `listAttendance`                                                                                                                         | absent; `apps/flow/src/features/time/TimeReports.tsx:142` gates a Clocked column on a local `trackAttendance` flag                                              | Time Reports                                                                                  | B     | Clocked-versus-logged reconciliation is offered as a toggle with no data behind it |
| I21 | Shift roster                       | HR          | `hr/shift/v1` `getRoster`                                                                                                                                   | absent                                                                                                                                                          | would inform Calendar, `features/calendar/`                                                   | B     | Flow's calendar cannot see that someone is on shift elsewhere                      |
| I22 | Staff availability                 | HR          | `hr/shift/v1` `listAvailability`                                                                                                                            | absent                                                                                                                                                          | would inform assignment in `features/tasks/`                                                  | B     | Work is assigned to people who declared themselves unavailable                     |
| I23 | **Customer for a project**         | CRM         | `packages/voxie/api/crm/customer/v1/models.onk:4-27` `Customer`; **no link field exists on Flow's side**                                                    | absent; would go on `features/projects/ProjectSettings.tsx`                                                                                                     | Project settings save                                                                         | D     | A client project records its client in its name string, or not at all              |
| I24 | Project budget                     | Finance     | **none**. No budget concept exists in `finance/` or `flow/`                                                                                                 | absent                                                                                                                                                          | Project settings                                                                              | D     | Budget lives in a spreadsheet                                                      |
| I25 | Expenses against a project         | Finance     | `finance/expenses/v1`; **no `project_id` anywhere in `finance/`**                                                                                           | absent                                                                                                                                                          | Project cost view                                                                             | D     | Non-labour cost cannot reach a project                                             |
| I26 | Purchase orders and materials      | ERP         | `erp/purchase_order/v1`; **no `project_id` anywhere in `erp/`**                                                                                             | absent                                                                                                                                                          | Project cost view                                                                             | D     | Materials cost cannot reach a project                                              |
| I27 | Invoice status for a project       | Finance     | `finance/invoices/v1` `listInvoices`; no project or customer FK on the invoice                                                                              | absent                                                                                                                                                          | Project header                                                                                | D     | Nobody in Flow can see whether the client paid                                     |

Inbound totals: 12 class A, 6 class B, 4 class C, 5 class D.

## 4. Inventory, outbound: Flow as producer

Direction is "Flow writes" or "counterpart reads Flow" for every row. Note that
**every single outbound link is unbuilt on the consumer side**: no app outside
Flow contains a Flow endpoint literal or a `voxie/flow/` import.

| #   | Data                                                          | Counterpart            | Contract path, message and field                                                                                                                                                                     | Frontend site                                                                                        | Trigger                           | Class                              | Consequence today                                                                                                |
| --- | ------------------------------------------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------- | ---------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| O1  | Logged hours, with project, task, work type and billable flag | HR                     | `packages/voxie/api/flow/time/v1/models.onk:88-116` `WorkDayLine`, persisted directly as `hr_time_entries` (`~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:4-8`)  | `features/time/useWorkDay.ts`, `features/time/AddHours.tsx`                                          | Log Hours submit                  | A                                  | Works, and is the strongest precedent in the platform                                                            |
| O2  | Work type to payroll category mapping                         | HR                     | `packages/voxie/api/flow/time/v1/models.onk:21` `WorkType.hr_category_id`                                                                                                                            | `features/settings/sections/TimeSection.tsx` category list                                           | Settings, Time                    | A                                  | Contract and column exist. The Flow UI carries a `GateNotice` saying nothing reads the category                  |
| O3  | Completed tasks as draft time entries                         | HR                     | **none**; live `POST /hr/time-entries/import-flow`                                                                                                                                                   | `apps/hr/src/shared/hooks/queries/useTimeEntries.ts:86` exists with **zero call sites**              | none                              | B                                  | An unreachable chain: five working backend links, one broken frontend link, two absent. #869 decides its fate    |
| O4  | Approved billable entries for invoicing                       | Finance                | **none**; live `POST /hr/time-entries/export-billable`, returns `BillableExportItem` with `project_id` and `task_id` (`~/Development/onion/internal/modules/hr_time/internal/domain/types.go:90-99`) | none in any app                                                                                      | none                              | B                                  | Billable hours are marked, batched, and exported to nowhere                                                      |
| O5  | Task counts, velocity, cycle time                             | Insights               | `packages/voxie/api/insights/dashboard/v1/models.onk:53-68` `GetInsightsResponse`; handler stubbed at `~/Development/onion/internal/modules/insights/internal/api/setting_service.go:102`            | `src/modules/insights/shared/hooks/queries/useInsightsDashboard.ts` calls only `getDashboardSummary` | none                              | C                                  | Company-wide delivery metrics do not exist outside Flow                                                          |
| O6  | Project roadmap                                               | Insights               | `packages/voxie/api/insights/dashboard/v1/models.onk:18-35` `RoadmapProject`; handler stubbed at `~/Development/onion/internal/modules/insights/internal/api/setting_service.go:110`                 | none                                                                                                 | none                              | C                                  | Duplicate of `flow/insights/v1` `RoadmapProject`, with a different shape and `int64` ids where Flow uses strings |
| O7  | Per-person workload                                           | Insights               | `packages/voxie/api/insights/dashboard/v1/models.onk:3-16` `WorkloadEntry`; handler stubbed at `~/Development/onion/internal/modules/insights/internal/api/setting_service.go:114`                   | none                                                                                                 | none                              | C                                  | An owner cannot see delivery load beside headcount and payroll cost                                              |
| O8  | Business-wide KPI tile counts                                 | Insights               | `packages/voxie/api/insights/dashboard/v1/models.onk:72-85` `GetKPIsResponse`; handler stubbed at `~/Development/onion/internal/modules/insights/internal/api/setting_service.go:106`                | none                                                                                                 | none                              | C                                  | `GetKPIsResponse` has no delivery field to carry Flow's contribution                                             |
| O9  | **Project labour cost**                                       | Finance, CRM, Insights | **none.** No cost field anywhere in `flow/`                                                                                                                                                          | absent                                                                                               | Project header, P&L, account view | D                                  | The motivating example. See section 6                                                                            |
| O10 | **Billable hours as invoice lines**                           | Finance                | `finance/invoices/v1` `InvoiceItem` (`:4-20`) can hold them today with `unit: "hour"`; nothing records the source                                                                                    | `apps/finance/src/features/invoices/NewInvoice.tsx:122` hardcodes `unit: 'item'`                     | New Invoice                       | B for creating, D for linking back | Every billable line is retyped by hand from a Flow report                                                        |
| O11 | Estimate versus actual                                        | Insights, Finance      | `packages/voxie/api/flow/task/v1/models.onk:33` `estimated_hours` and summed `WorkDayLine.duration_minutes` both exist; **no aggregation endpoint**                                                  | absent                                                                                               | Project header                    | C                                  | Overrun is invisible until the invoice is disputed                                                               |
| O12 | Delivery dates and blocked state                              | CRM                    | `flow/project/v1/models.onk` `due_date`, `status`; no CRM-side field to receive it                                                                                                                   | absent                                                                                               | Account view in CRM               | D                                  | An account manager phones the project manager                                                                    |
| O13 | Milestones                                                    | CRM, Finance           | `packages/voxie/api/flow/milestone/v1/models.onk:8-20` `Milestone` with `target_date`, `computed_progress`                                                                                           | absent                                                                                               | Milestone billing trigger         | D                                  | Milestone-based invoicing is manual                                                                              |
| O14 | Presence and work mode                                        | HR                     | **none**; live `GET /flow/presence/today` (`~/Development/onion/internal/modules/workspaces/presence.go:98`), a server-side join of Flow work days, HR employees and HR leave                        | `apps/flow/src/features/presence/usePresence.ts:15`                                                  | Presence screen load              | A                                  | Works. The best existing model for a reduced cross-app projection                                                |
| O15 | Cycles and sprints                                            | Insights               | `flow/cycle/v1`                                                                                                                                                                                      | absent                                                                                               | none                              | D                                  | No consumer has asked                                                                                            |

Outbound totals: 3 class A, 3 class B, 4 class C, 5 class D.

---

## 5. Gap analysis by class

### Class A: wired and working

Twelve inbound, three outbound. Three are worth copying and one is worth
copying carefully.

**Copy this: `features/directory/useDirectory.ts`.** Nine lines of setup for a
two-service cross-app read.

```ts
const staffClient = createOneKitClient(StaffServiceClient, 'flow');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'flow');
```

The `'flow'` app id scopes the API client. Tenancy and permission are both
server-side: the doc comment at `:14-19` records that `view_directory` produces a
real 403 rather than a hidden button, and that field-gated sensitive columns
simply arrive absent. Nothing about tenancy is asserted by the client.

**Copy this: `onion internal/modules/workspaces/presence.go`.** The best model in
the codebase for the pattern section 7 needs. `ListPresenceForBusiness` joins
Flow work days, HR active employees, HR approved leave and HR permissions
server-side, and returns a **deliberately reduced projection**. Lines 14-18 say
why:

> mode and status only, deliberately no minutes: a company-wide view watched by
> every employee, including people with no management relationship to each other,
> is a different exposure than the workspace-admin-gated team view.

That is the "server-computed only" decision already applied once, to a different
sensitive quantity, by the same reasoning.

**Copy this, but know what it costs: `flow_work_day_lines` does not exist.** Flow
writes work-day hours straight into `hr_time_entries`, and payroll reads exactly
what Flow wrote there. That is one store for the work day, not one store for
Flow's hours overall: `flow_time_logs` is a second, and nothing reconciles them
(see Finding 1's correction and #903). It also means a Flow bug writes into
payroll's table, which is why section 7 treats hours as the highest-consequence
row in the audit.

**Copy the mechanism, not the shape: `WorkType.hr_category_id`.** The join-key
precedent the starting brief pointed at is real
(`packages/voxie/api/flow/time/v1/models.onk:19-21`) and its doc comment states the intent: "so one
entry serves both project reporting and payroll." Every new link in section 7
uses this mechanism, a nullable id on the Flow-side record pointing at the
owning app's record, rather than inventing a new one.

### Class B: contract exists, frontend does not call it

Shippable without touching onion. Each row below names the contract file, message
and field, per the rule that an uncitable claim is not class B.

| Row                | Contract citation                                                                                                                                       | Work                                                                                                                                                                      |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| I14 currency       | `packages/voxie/api/hub/business/v1/models.onk:16` `Business.currency`, returned by `getBusiness` (`packages/voxie/api/hub/business/v1/service.onk:18`) | Two calls: `/businesses/mine` for the id Flow already fetches, then `getBusiness`. `ApiBusinessSummary` (`apps/flow/src/shared/services/flowApi.ts:46-50`) drops it today |
| I19 approved leave | `hr/leave/v1` `listLeaveRequests`, already called at `apps/flow/src/features/leave/useLeave.ts:34`                                                      | Reuse the existing hook on Roadmap and Workload                                                                                                                           |
| I20 attendance     | `hr/attendance/v1` `listAttendance`                                                                                                                     | New hook; gates the existing Clocked column at `apps/flow/src/features/time/TimeReports.tsx:142`                                                                          |
| I21 roster         | `hr/shift/v1` `getRoster`                                                                                                                               | New hook for Calendar                                                                                                                                                     |
| I22 availability   | `hr/shift/v1` `listAvailability`                                                                                                                        | New hook for assignment                                                                                                                                                   |
| O3 flow import     | Uncontracted but live: `the import route (deleted from onion by #701, "Delete the flow time import chain")`                                             | A button. The hook already exists at `apps/hr/src/shared/hooks/queries/useTimeEntries.ts:86` with zero callers                                                            |
| O4 billable export | Uncontracted but live: `~/Development/onion/internal/modules/hr_time/internal/routes/handlers/export_billable_handler.go:15`                            | A screen in Finance or HR                                                                                                                                                 |
| O10 invoice lines  | `packages/voxie/api/finance/invoices/v1/models.onk:4-20` `InvoiceItem` accepts `description`, `quantity`, `unit`, `unit_price`                          | Prefill from Flow hours. `apps/finance/src/features/invoices/NewInvoice.tsx:122` hardcodes `unit: 'item'`                                                                 |

**A caution on O3 and O4.** Both endpoints are live and neither is in the
contracts. Calling an uncontracted endpoint is what `useMyEmployeeId.ts` and
`usePresence.ts` already do, so it is established practice here, but it means
`pnpm contracts:audit-requests` cannot check the request shape. Adding the `.onk`
definitions should be part of the same work, not a follow-up, and issue #6 in
section 10 says so.

### Class C: endpoint exists, field does not

Needs an `.onk` change plus an onion change. Two of these are sharper than the
class label suggests, and filing them as class B would put dead work in wave one.

**C-contract-complete: the four Insights handlers.** O5 through O8 need **no
`.onk` change at all**. `insights/dashboard/v1/models.onk` is complete, down to
`avg_cycle_time_days` and `velocity_change`. The four handlers return
`notImplemented`
(`~/Development/onion/internal/modules/insights/internal/api/setting_service.go:102,106,110,114`).
Pure onion work, then pure frontend work.

**C-message-without-RPC: contracted hours.** I17 and I18 need no new message
either. `packages/voxie/api/hr/work_schedule/v1/models.onk:33` defines `ListWorkSchedulesResponse`
and its own doc comment says: "Not currently wired to any RPC in service.proto
(no request message or list RPC exists) - kept for parity with the source proto."
The payload exists, the RPC does not. Adding one `listWorkSchedules` RPC plus its
request message unblocks Flow's entire capacity story and retires most of
`timeSettingsStore`.

The remaining class C rows are genuine field additions: I4 needs a `user_id`
filter or self-lookup on `hr.staff.v1`; I15 and I16 need the cost projection of
section 7, not the rate itself; O11 needs an aggregation endpoint.

### Class D: concept does not exist

Needs data model design. Section 7 designs D1 through D3; section 9 rejects the
rest.

- **I23 and O12, the customer link.** `flow.project.v1.Project` has no
  `customer_id`. Decided: nullable, gated on a workspace setting.
- **O9, project cost.** No cost field in Flow, and no rate may reach the client.
  Designed as a server-computed projection.
- **O10's return path.** `finance.invoices.v1` has no `project_id`, so an invoice
  raised from Flow hours cannot be traced back.
- **I25 and I26, non-labour cost.** `project_id` appears zero times in the whole
  of `erp/` and `finance/`. Attributing an expense or a purchase order to a
  project is a change to those domains, not to Flow. Deferred, with reasons in
  section 9.
- **I5 is a special case.** `GET /hr/staff/detail` is called at
  `the profile call site (rewritten upstream)`, is absent from every `.onk` file, and does
  not exist in onion: `grep -rn "/hr/staff" internal/` outside generated code
  returns nothing. This is not a gap to design, it is a bug to file.

---

## 6. The worked example, end to end: what a Flow project costs

The brief's motivating calculation, traced through every layer as it would
actually run. This section is the design of O9 and the justification for I15.

### What the user wants

A project manager opens a Flow project and wants one line: this project has cost
44,800 USD in labour, against an estimate of 40,000.

### What that requires

`project labour cost = sum over entries of (duration_minutes / 60 * rate for that person)`

Three operands, owned by three different places.

**Operand 1, the hours. Already correct, already shared.** Every Flow work-day
line is a row in `hr_time_entries` carrying `project_id`, `task_id`,
`duration_minutes` and `is_billable`
(`packages/voxie/api/flow/time/v1/models.onk:88-116`, and `~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:4-8` for the storage
fact). No work needed. This is the finding that makes the rest cheap.

**Operand 2, the rate. Exists in HR, and Flow currently retypes it.**
`packages/voxie/api/hr/employee/v1/models.onk:28` `hourly_rate`, `:32` `currency`. Flow's own
imitation is `apps/flow/src/shared/stores/timeSettingsStore.ts:50` plus the settings UI at
`apps/flow/src/features/settings/sections/TimeSection.tsx:409-462`, per browser, read by nothing.

**Operand 3, the currency. In Hub, and Flow cannot see it.**
`packages/voxie/api/hub/business/v1/models.onk:16`. The endpoint Flow already calls,
`/businesses/mine`, returns `BusinessEntry` (`:25-34`), which has no currency
field. Reachable in one extra call to `getBusiness`.

### Where the multiplication happens, and why it matters

Not in the browser. `packages/voxie/api/hr/staff/v1/models.onk:134-135` already states the platform's
position:

> It never carries salary, hourly*rate, overtime_rate, address, or
> emergency_contact*\* - those stay HR-admin-only

If Flow computed cost client-side it would need the rate in a Flow API response,
and the network tab would become a permission bypass for every project manager.
So onion computes it and returns the total. A sum cannot be inverted into a
salary.

There is one residual leak the design must handle explicitly: **on a
single-person project, total cost divided by hours is that person's exact rate.**
Section 7 handles it with a minimum-headcount threshold, because no amount of
server-side computation fixes an aggregate of one.

### The chain, step by step

1. PM opens `/w/:workspaceId/p/:projectId`. Route defined in
   `apps/flow/src/router.tsx`.
2. A new `useProjectCost(projectId)` hook calls a new
   `GET /flow/project/v1/projects/{id}/cost`.
3. onion resolves the tenant: `tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)`,
   the same call `the import handler tenant resolution (route deleted by onion #701; pattern preserved at internal/modules/workspaces/presence.go)` makes. The client never sends a
   business id.
4. onion checks a Flow-side permission for cost visibility, then joins
   `hr_time_entries` filtered on `project_id` and `business_id` against
   `hr_employees.hourly_rate` on `employee_id`, both filtered on `business_id`.
5. onion returns `{ total_cost, currency, hours, entry_count, distinct_employees, rate_coverage }`.
   No per-person rows, no rates.
6. Flow renders it. `rate_coverage` below 1.0 means some contributor has no rate
   on file, and the number is labelled a partial rather than shown as if
   complete.

### What Finance and CRM then do with it

Finance reads the same figure for a project P&L line. CRM reads it, joined
through `Project.customer_id`, for account profitability. Neither recomputes it:
one owner, one number, which is the whole point of doing it server-side.

### What breaks if it is wrong

This row touches money. A cost figure feeding an invoice can overbill a client;
one feeding a P&L misstates a company's accounts. Two specific failure modes the
design must foreclose:

- **A missing rate silently reading as zero.** `apps/flow/src/shared/stores/timeSettingsStore.ts:49` already
  gets this right in its comment: "`null` is explicitly unset, never zero." The
  server projection must carry `rate_coverage` so the UI can distinguish "cost is
  40,000" from "cost is at least 40,000."
- **The tenancy default.** `~/Development/onion/internal/platform/database/schemas/hr_time/schema.sql:4` declares
  `business_id BIGINT NOT NULL DEFAULT 0`. A join that omits the filter does not
  fail, it returns another tenant's hours. See section 7.4.

---

## 7. Data model and event design

Only for the links section 8 schedules. Every one extends the
`hr_category_id` precedent: a nullable id on a record pointing at the owning
app's record, resolved server-side, never a copied value.

### 7.1 Read live or read a projection

| Data                                            | Pattern                                        | Why                                                                                |
| ----------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
| Staff, departments, leave, availability, roster | Live read of the owner's endpoint              | Already the established pattern. Staleness is the user's own recent action         |
| Contracted hours                                | Live read, cached by React Query               | Changes rarely; a stale capacity bar is a small error                              |
| Project cost                                    | Server-computed projection                     | The operands must not cross the wire. Not a caching decision, an exposure decision |
| Customer name on a project                      | Live read via `customer_id`                    | Never snapshot the name. See 7.3                                                   |
| Billable hours to an invoice                    | Snapshot at invoice creation, with a link back | An invoice is a legal record. It must not change when a time entry is edited later |

### 7.2 Cost projection

New RPC on Flow's project service, since a project is the thing being costed and
Flow owns projects.

```
// flow/project/v1/models.onk
message ProjectCost {
  /// Total labour cost over the requested range, in the business currency.
  total_cost: float64
  /// ISO-4217 code, resolved server-side from the business record.
  currency: string
  /// Hours the total is computed from.
  hours: float64
  /// Entries summed, so the UI can distinguish empty from zero-cost.
  entry_count: int32
  /// Distinct contributors. Below the disclosure threshold the total is withheld.
  distinct_employees: int32
  /// Fraction of hours whose contributor had a rate on file, 0.0 to 1.0.
  /// Below 1.0 the UI must label the figure a partial, never a total.
  rate_coverage: float64
}
```

`GET /flow/project/v1/projects/{id}/cost`, with optional `from` and `to`.

Three rules the handler must enforce, each closing a specific hole:

1. **Minimum headcount.** If `distinct_employees < 2`, return the hours and
   `total_cost` unset rather than a number, because a one-person total divided by
   its own hours is that person's rate. Aggregation is not anonymisation when
   n equals 1.
2. **Never substitute zero for a missing rate.** Exclude the hours from
   `total_cost`, count them in `hours`, and let `rate_coverage` fall below 1.0.
3. **Permission-gated, Flow-side.** Cost visibility is a Flow project permission,
   not an HR one: the question is "may this person see this project's economics",
   and the answer differs per project. Follow `useDirectory`'s model, a real 403
   rather than a hidden button.

### 7.3 Customer link

```
// flow/project/v1/models.onk, on Project and on Create/Update
/// crm_customers.id, or null for an internal project. Never a copied name:
/// a renamed customer must not leave a stale label on a project.
customer_id: int64? @nullable
```

Owner: CRM. Flow stores the id and reads the name live, exactly as
`useDirectory` reads department names to decorate staff rows.

**The mistake not to repeat is already in the codebase.**
`apps/finance/src/shared/components/ClientSelect.tsx:25` reads `/crm/customers`
live, which is right, and then `apps/finance/src/features/invoices/NewInvoice.tsx:130` writes `clientName: client.name`
into `CreateInvoiceRequest`, which has `client_name`, `client_email`,
`client_address` and `client_tax_id` as free-text strings and **no `customer_id`**
(`finance/invoices/v1/models.onk`, `CreateInvoiceRequest`). Finance therefore
cannot answer "what have we billed this customer" without string matching. Flow
must store the id.

On an invoice the snapshot is correct, because an issued invoice must not mutate.
The fix there is to keep the snapshot **and add** the FK, not to replace it.

Gating: a nullable `customer_id` plus a `client_projects_enabled` workspace
setting. Per `docs/flow-time-tracking.md`'s rule, this is a genuine server-owned
setting and not the browser-local seam that document describes, so it belongs on
`flow/setting/v1`, not in a Zustand store.

### 7.4 Tenancy, which is the part that can go worst

Three layers, and the first two are already how the platform works.

**Layer 1, the handler resolves the tenant.** `tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)`,
as at `the import handler tenant resolution (route deleted by onion #701; pattern preserved at internal/modules/workspaces/presence.go)`. The client never sends `business_id` and
cannot influence it. Every proposed endpoint uses this.

**Layer 2, every query filters on it.** `the flow adapter scoping example (module code since reworked; convention documented in onion CLAUDE.md)` shows the shape:
`sq.Eq{"id": employeeID, "business_id": businessID}` on the HR side and
`sq.Eq{"t.assignee_id": userID, "t.business_id": businessID}` on the Flow side.
A cross-schema join filters on **both** sides, not one.

**Layer 3, and this is the hazard.** `~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:9-12` documents it:

> business_id is NOT NULL with no default on every new table here, deliberately.
> The older tables default it to 0, which turns a forgotten value into a silently
> shared tenant rather than an error.

Verified, and larger than a file count suggests: `business_id BIGINT NOT NULL
DEFAULT 0` is declared on **145 distinct tables** across **57 schema files**. The
live schema agrees, so this is not a parsing artefact: of 170 columns named
`business_id` in the local `onion_dev` database, **146** carry
`COLUMN_DEFAULT = '0'`. Affected tables include every one of the nine in
`~/Development/onion/internal/platform/database/schemas/flow/schema.sql` (`:4,29,65,116,127,139,156,172,205`),
`~/Development/onion/internal/platform/database/schemas/hr_time/schema.sql:4` for `hr_time_entries`, `~/Development/onion/internal/platform/database/schemas/crm/schema.sql:4` for customers,
and `~/Development/onion/internal/platform/database/schemas/voxire/expenses.sql:4`. Only the three v8 tables
(`~/Development/onion/internal/platform/database/schemas/flow/v8_time_tracking.sql:17,52,68`) are safe.

So every join this document proposes runs over at least one table where a dropped
`business_id` yields tenant 0 rather than an error.

**The rows exist, and tracing them showed they are not a leak.** Running the
audit read-only against the local `onion_dev` container, 34 of those 145 base
tables hold rows at `business_id = 0`, and five hold such rows across more than
one `owner_id`, which `packages/voxie/api/flow/project/v1/models.onk:8` defines as
the tenant. Tracing all five to their INSERT explained all five:

| Table                          | Why                                                                                                                                                                                                                                                                                         | Leak |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- |
| `flow_workspaces`              | Deliberate. `CreatePersonalWorkspace` omits the column (`~/Development/onion/internal/modules/workspaces/sql/queries/members.sql:27-29`) and `GetWorkspacesByOwner` re-scopes those rows on `owner_id` (`~/Development/onion/internal/modules/workspaces/sql/queries/workspaces.sql:16-21`) | No   |
| `stock_locations`              | Explicit literal `0` in `InsertDefaultWarehouseLocation` (`~/Development/onion/internal/platform/stock_core/sql/queries/stock_core.sql:6-9`)                                                                                                                                                | No   |
| `hr_configs`                   | Column never written, never read; both queries scope `owner_id` (`~/Development/onion/internal/modules/hr_setup/sql/queries/hr_setup.sql:13-23`)                                                                                                                                            | No   |
| `stock_levels`, `stock_ledger` | Inserts omit it; no read in that file filters it                                                                                                                                                                                                                                            | No   |

`flow_workspaces` is documented in the code at `~/Development/onion/internal/modules/workspaces/sql/queries/workspaces.sql:8-11`: a personal
workspace "has no business_id (it isn't tied to one tenant)", and zero is the
intended encoding. Its rows are production data, written on every registration
and login (`~/Development/onion/cmd/api/main.go:723-727`), and correct.

**So the real finding is not exposure.** `owner_id` is this codebase's actual
tenant key; `business_id` is a newer column, inconsistently populated and on many
tables never read. The hazard is that **the first query to trust it on one of
these tables is wrong in the quiet direction**, returning an empty set rather than
an error. That is the `ownerID` trap of
`~/Development/onion/internal/modules/workspaces/presence.go:40-46` generalised,
and it matters here because every cross-app join either audit round proposes plans
to filter `business_id`. Two consequences for the roadmap:

- **Any new cross-app query needs a test that asserts isolation**, not only that
  it returns rows. A query missing its filter passes a happy-path test.
- **The `DEFAULT 0` removal is worth its own issue**, filed in section 10 as a
  platform prerequisite rather than folded into a feature. It needs a data audit
  first, as the v8 comment says.

**A second, subtler hazard, already documented.** `~/Development/onion/internal/modules/workspaces/presence.go:40-46`:

> The three ids passed to different modules are deliberately not the same value:
> hr_core/staff resolve "ownerID" to the tenant's businessID, while hr_leave's own
> handlers pass the caller's raw hub user id as what it calls "ownerID"

So "ownerID" means different things in different modules, and querying leave with
the wrong one "would silently return nothing rather than fail loudly." Any new
join touching HR leave must pass the caller's hub user id, not the business id.
This is a correctness trap that produces an empty result, which reads as "no
leave booked" rather than as an error.

### 7.5 Event flow, and why there is almost none

Every link above is a synchronous read, on purpose. Two exceptions.

**Invoice creation from billable hours** is the one place a write crosses apps,
and it needs a batch marker so the same hour is not billed twice.
`ExportBillableEntries` (`~/Development/onion/internal/modules/hr_time/internal/application/service.go:290-324`)
already does exactly this: `MarkEntryExported(ctx, businessID, e.ID, batchID)`
before adding the item, so a re-export skips it. The design reuses that
mechanism rather than adding one. Note the ordering: it marks first and appends
only on success, so a failure to mark drops the item from the batch rather than
double-billing it. That is the right direction to fail.

**Realtime** stays inside Flow. `RealtimeClient` in `@voxire/api` and Flow's
`useRealtimeStatus` handle in-app freshness. Pushing Flow task events to HR or
Finance would add a second consistency model for no benefit that a React Query
invalidation does not already give.

### 7.6 One contract hygiene note

`RoadmapProject` is defined twice, with different shapes:
`packages/voxie/api/flow/insights/v1/models.onk:45-56` uses `string` ids and has `color`,
`done_count`, `workspace_id`; `packages/voxie/api/insights/dashboard/v1/models.onk:18-35` uses
`int64` ids and has `progress_percent`, `total_tasks`. Whichever survives, the id
type matters at the wire: **OneKit `int64` crosses as a string**. A JSON number
fails to unmarshal, and the string `"0"` passes unmarshalling then fails `@gt(0)`
validation. Both id-carrying fields proposed in this document, `customer_id` and
the cost endpoint's path id, inherit that. Note also that `Boolean('0')` is
`true` in JavaScript, so a guard written as `if (id)` does not catch it.

---

## 8. Roadmap

Ranked by value per unit of effort, respecting dependencies. Effort is judgement
**(inference)**, not measurement.

### Wave 1: frontend only, ships today

No onion work, no `.onk` change. These prove the seam.

| Item                                                | Class | Effort | Depends on                  |
| --------------------------------------------------- | ----- | ------ | --------------------------- |
| I19, approved leave on Roadmap and Workload         | B     | S      | nothing; hook exists        |
| I14, business currency                              | B     | S      | nothing                     |
| I22, availability in task assignment                | B     | M      | nothing                     |
| I21, shift roster in Calendar                       | B     | M      | nothing                     |
| I20, attendance behind the existing Clocked toggle  | B     | M      | nothing                     |
| O3, a button for the Flow import that already works | B     | S      | fix the response type first |

**What a user can newly do:** plan a sprint that accounts for who is on holiday
and who is unavailable; see a currency label on every hour; and import their
completed Flow work into a timesheet from the HR app, which the platform has been
able to do for some time with no way to ask.

Wave 1 also retires a class of bug: `apps/flow/src/features/time/TimeReports.tsx:142` currently gates a
Clocked column on a local flag with no data behind it.

### Wave 2: unblock capacity, one small RPC

| Item                                                                                                                                                  | Class | Effort | Depends on |
| ----------------------------------------------------------------------------------------------------------------------------------------------------- | ----- | ------ | ---------- |
| I17 and I18, `listWorkSchedules` RPC plus Flow read                                                                                                   | C     | M      | onion      |
| Retire the capacity fiction at `apps/flow/src/layouts/FlowLayout.tsx:380` and `apps/flow/src/features/workspace-dashboard/WorkspaceDashboard.tsx:128` | C     | S      | above      |
| Migrate `timeSettingsStore`'s target and OT fields to the server values                                                                               | C     | M      | above      |
| I4, `user_id` filter on `hr.staff.v1` `listStaff`                                                                                                     | C     | S      | onion      |

**What a user can newly do:** see workload measured against their real contracted
hours instead of a hardcoded 40, and stop maintaining Flow's browser-local copy of
a schedule HR already owns. I4 also removes an unbounded roster read that Flow
performs on every profile view.

This wave is placed second because it is the cheapest class C in the audit and it
deletes code rather than adding it.

### Wave 3: cost, the motivating example

| Item                                            | Class    | Effort | Depends on                         |
| ----------------------------------------------- | -------- | ------ | ---------------------------------- |
| O9, the `ProjectCost` projection of 7.2         | D        | L      | I14 currency, wave 1               |
| Delete `ratesByUserId` and the rate settings UI | C        | S      | above                              |
| The `DEFAULT 0` tenancy audit                   | platform | L      | nothing, but blocks nothing either |

**What a user can newly do:** open a Flow project and see what it has cost, with
an honest partial label when someone has no rate on file. A PM stops exporting
hours to a spreadsheet, and an admin stops retyping salaries into a settings
screen that discards them.

The tenancy audit is listed here rather than as a wave-1 blocker deliberately.
Every proposed query filters `business_id` explicitly, so the default is a
latent hazard rather than an active one. It should not gate delivery, and it
should not be forgotten.

### Wave 4: the customer link and its consequences

| Item                                                    | Class     | Effort | Depends on |
| ------------------------------------------------------- | --------- | ------ | ---------- |
| I23, `customer_id` on `Project` plus the workspace gate | D         | L      | 7.3        |
| O12, delivery status on the CRM account view            | D         | M      | above      |
| O10, prefill invoice lines from billable hours          | B, then D | M      | wave 3     |
| `customer_id` and `project_id` on `finance.invoices.v1` | C         | M      | above      |

**What a user can newly do:** raise an invoice for a client's project without
retyping a single hour, and answer "is this account profitable" by joining cost
to billings. This is also where the audit fixes Finance's existing free-text
client problem rather than copying it.

### Wave 5: the platform view

| Item                                                   | Class | Effort | Depends on                         |
| ------------------------------------------------------ | ----- | ------ | ---------------------------------- |
| O5 to O8, implement the four stubbed Insights handlers | C     | L      | pick one `RoadmapProject`, see 7.6 |
| O11, aggregation endpoint for estimate versus actual   | C     | M      | nothing                            |
| O4, a screen for the billable export that already runs | B     | M      | wave 4                             |

**What a user can newly do:** see delivery metrics beside revenue and headcount on
one Insights dashboard, which today is four contract messages and no
implementation.

Placed last despite containing class B and contract-complete work, because
`getDashboardSummary` is the only Insights dashboard call the frontend makes, so
this wave delivers a screen nobody is currently asking for. Its value is real and
its urgency is not.

---

## 9. Not worth building

A report where every proposed integration is worth building is not a report
anybody should trust. Six.

**1. Flow reading ERP purchase orders and materials against a project (I26).**
`project_id` appears zero times in the entire `erp/` domain. Making materials
cost land on a Flow project means adding project attribution to purchase orders,
stock transfers and inventory adjustments: a change to ERP's data model to serve
a Flow screen. Do it when a customer asks for true project cost including
materials, and do it as an ERP project with a Flow consumer, not as a Flow
integration. **What would change my mind:** a construction or manufacturing
customer, where materials dominate labour.

**2. Flow reading Finance expenses against a project (I25).** Same reasoning,
same zero-occurrence evidence, slightly cheaper. Worth revisiting immediately
after wave 4, since an invoice carrying `project_id` establishes the FK that an
expense would reuse.

**3. A project budget field in Flow (I24).** Tempting and wrong at this stage. A
budget only means something next to actuals, and until wave 3 Flow has no
actuals. Shipping a budget field first produces a number the user types in and
compares to nothing, which is the `ratesByUserId` mistake again. Build it in
wave 3 or later, against cost.

**4. Recruit to Flow onboarding tasks.** Recruit has **no contract domain at
all**: `packages/voxie/api` has no `recruit` directory, and `apps/recruit` talks
to `/recruit/*` through `safeFetch` with zero `voxie` imports. So this is class D
on both sides. It is also the wrong app: `hr/onboarding/v1` already has checklist
templates, per-employee checklists, triggering and item completion. A new hire's
onboarding belongs in HR's checklist, not duplicated as Flow tasks. Building it
in Flow would create two competing onboarding systems.

**5. Pushing Flow task events to HR, Finance or CRM over realtime.** Every
consumer identified in this audit wants an aggregate on screen load, not a live
feed. Adding a second consistency model buys nothing a React Query invalidation
does not already provide, and it would put Flow task churn on a path that
touches payroll.

**6. Per-person rates reaching the Flow client, in any form.** Including "just
for admins." `packages/voxie/api/hr/staff/v1/models.onk:134-135` already documents the platform's
position, and section 7.2 shows the aggregate can serve every identified need
without it. The one place aggregation genuinely fails, a single-contributor
project, is handled by withholding the total rather than by relaxing the rule.
A permission-gated rate read would be easier to build and would put salary data
in a browser that has no reason to hold it.

**One more, listed separately because it is a bug rather than a design choice.**
`the profile call site (rewritten upstream)` calls `GET /hr/staff/detail`. That path is in
no `.onk` file and in no onion route: `grep -rn "/hr/staff" internal/` outside
generated code returns nothing. Do not design around it. Fix or delete it.

---

## 10. Relationship to the platform-wide audit

This document is the second round of the audit whose other 34 documents are its
siblings in this directory. The first round was platform-wide: every ordered pair
of the ten apps and twelve contract domains, 21 proposals and 18 rejections,
issues #834 to #862, labelled `audit:cross-app-2026-08-21`. This round is
Flow-centric, labelled `audit:flow-cross-app-2026-08-21`, and it exists because
one app read in depth surfaces things a 132-cell matrix cannot hold.

It lands on the same branch deliberately. Two overlapping cross-app audits
arriving on `main` separately would be worse than one, and this way the shared
rows below are reconciled before either is merged rather than after.

**Three of its issues are Flow links and this document does not re-propose them.**

| Their issue                                         | This document's row | Note                                                                                                                                                                                                                                                         |
| --------------------------------------------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| #837 `HR->FLOW-01: project labour cost`             | O9, section 6       | Same conclusion by the same reasoning: HR owns the aggregate, returns totals only, carries a `rate_coverage` figure, and suppresses the total on a one-contributor project. Their `cost_minor_units: int64` is better than the `float64` in 7.2; take theirs |
| #838 `HR->FLOW-02: approved leave gates assignment` | I19                 | Same link                                                                                                                                                                                                                                                    |
| #836 `FLOW->FIN-01: unbilled billable hours`        | O4, O10             | Same link                                                                                                                                                                                                                                                    |

Their `docs/cross-app/11-rejected.md` also rejects `FLOW->HR-01` (workload on the
employee page), `FLOW->HR-02` (cycle commitments versus shifts), `FIN->FLOW-01`
(budget consumed) and `FLOW->INS-01` (delivery metrics in Insights). Section 9 of
this document reaches the same verdict on the budget item independently, and
section 5 disputes the evidence behind the `FLOW->INS-01` rejection without
disputing its conclusion.

**Where this document adds something.** A Flow-centric sweep found six things a
platform-wide sweep did not. Verified by searching the first round's documents
in this directory at `d62ad614`, the commit this round was branched from:

| Finding                                                                                          | Their coverage                                                                                       | Filed |
| ------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- | ----- |
| Flow's `ratesByUserId` localStorage rate table with no reader                                    | zero hits for `ratesByUserId`                                                                        | #867  |
| `capacity: 40` hardcoded, `contracted_hours_per_week` has no read RPC                            | zero hits for `capacity: 40`; `contracted_hours_per_week` appears only as a CSV inventory row        | #868  |
| `useImportFromFlow` has zero call sites and a response-type mismatch                             | the endpoint is in `inventory/endpoints.csv:584`; the dead hook and the mismatch are not             | #869  |
| All four `insights.dashboard.v1` handlers return `notImplemented`                                | zero hits for `notImplemented`; `docs/cross-app/11-rejected.md:64-69` assumes the aggregate is built | #870  |
| `business_id` unreliable across 145 tables, so a join trusting it silently returns the wrong set | zero hits for `DEFAULT 0`                                                                            | #871  |
| Flow's currency is unreachable through the endpoint it already calls                             | `BusinessEntry` is mentioned; the gap is not                                                         | #872  |

The seventh, that `flow_work_day_lines` is not a table and Flow's work-day hours
are rows in `hr_time_entries`, is recorded in section 1 rather than filed,
because it is a fact to know rather than work to do. It did become work to do,
but not the work it looked like: the accompanying claim that Flow keeps no
second store of hours was false, and tracing it produced #890 and #903. Zero hits for `flow_work_day_lines` in their
documents.

They also found something this document did not, and it is worth reading before
building any of the above: `docs/cross-app/10-doc-drift.md:50` independently
identifies that `GET /hr/staff/detail` has no onion route, matching row I5, and
correctly labels the production 404 as inferred rather than verified.

## 11. Filed issues

All labelled `audit:flow-cross-app-2026-08-21` on `tryrtylr/rtylr`.

| Issue | Link                                                                                                   | Class                                              | Blocked on onion                              |
| ----- | ------------------------------------------------------------------------------------------------------ | -------------------------------------------------- | --------------------------------------------- |
| #867  | Flow re-asks for the HR hourly rate, stores it per-browser, reads it back nowhere                      | I15, class C                                       | Partly: sequenced behind #837                 |
| #868  | Flow hardcodes a 40-hour capacity because HR's `contracted_hours_per_week` has no read RPC             | I17, I18, class C                                  | Yes                                           |
| #869  | HR's Flow time import is live but unreachable, and its response type does not match the handler        | O3, class B plus a bug                             | Frontend fix no; two correctness problems yes |
| #870  | `insights.dashboard.v1` declares four Flow-shaped endpoints and every handler returns `notImplemented` | O5 to O8, class C                                  | Yes                                           |
| #871  | `business_id` is unreliable across 145 tables, so a join that trusts it silently returns the wrong set | platform prerequisite                              | Yes, entirely                                 |
| #873  | Contract drift: 109 of onion's 227 hand-registered routes have no `.onk` definition                    | platform, and the reason O3's mismatch went unseen | Partly                                        |
| #874  | `onion/.env` points a development-labelled config at a production RDS host                             | platform                                           | Yes, entirely                                 |
| #875  | Flow and HR both call `/hr/staff/detail`, which does not exist, and both swallow the failure           | I5, a live bug                                     | Option A yes, Option B no                     |
| #872  | Flow has no currency: the `/businesses/mine` row it already reads has no currency field                | I14, class B                                       | No                                            |

Not filed, and why:

- **Rows I19 to I22 (leave, attendance, roster, availability).** I19 is #838. The
  other three are the same HR-to-Flow planning link at three levels of detail;
  filing them separately would produce three issues one PR closes. They belong as
  scope on #838, and a comment there says so.
- **Rows I23 to I27 and O12 to O15.** Class D, rejected or deferred in section 9,
  or already covered by #836.
- **Row I5, `/hr/staff/detail`.** Already in their `docs/cross-app/10-doc-drift.md:50`.
- **Row O11, estimate versus actual.** Overlaps #855 (`Contract: add range,
project and billability filters to hr time entries`), which is the aggregation
  this needs. Better as scope there than as a new issue.
