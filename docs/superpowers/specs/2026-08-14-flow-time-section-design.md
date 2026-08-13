# Flow Time: a first-class section, backed for real

Date: 2026-08-14
Status: **proposal, awaiting approval**

## What this corrects

The first pass treated time tracking as configuration: a settings pane, a panel in
the task rail, and a panel behind the board's gear icon. Logging hours is a primary
Flow capability, so it needs a section of its own with real destinations and real
depth. Three of fifteen designed screens exist, the rest were reported as blocked,
and the parts that shipped save to a browser rather than a database.

This proposal builds it properly: the contract fields and endpoints in `voxie` and
`onion`, and the screens in `rtylr`. Nothing stubbed, nothing client-only.

## Where the code lives, verified

| Layer        | Repo    | Path                                                                                                                                                             |
| ------------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Contracts    | `voxie` | `api/flow/<capability>/v1/{models,service}.onk`, with committed `gen/go` + `gen/ts`                                                                              |
| Flow backend | `onion` | `internal/modules/workspaces/` — `internal/api/task_service.go`, `internal/application/service.go`, `internal/infrastructure/repository.go`, `sql/queries/*.sql` |
| Schema       | `onion` | `internal/platform/database/schemas/flow/`, MySQL, versioned files (`v6_status_backlog.sql`, `v7_…`), so this adds `v8_time_tracking.sql`                        |
| sqlc codegen | `onion` | `make generate` → `scripts/generate_sqlc.sh`                                                                                                                     |
| Gates        | `onion` | `make lint`, `make test`, `make test-db` (needs `make local-env-up`)                                                                                             |
| Frontend     | `rtylr` | `apps/flow/src/`                                                                                                                                                 |

Today's table is the whole problem in nine lines:

```sql
CREATE TABLE flow_time_logs (
  id, owner_id, business_id, task_id, user_id,
  duration_minutes, description, logged_at, created_at
);
```

No date distinct from `logged_at`, no category, no billable flag, no status, no
approval, no lock. `TimeLog` in `voxie` mirrors it, and only three RPCs exist:
create, list-by-task, delete.

## Information architecture

The sidebar `Time` group gets real destinations. Settings keeps configuration only.

| Route               | Screen                                                                                  | Who             |
| ------------------- | --------------------------------------------------------------------------------------- | --------------- |
| `/time`             | **My week.** The timesheet grid: days across, tasks down, running totals, submit.       | everyone        |
| `/time/review`      | **Review.** Flat queue of submitted weeks and entries to approve or return.             | managers        |
| `/time/reports`     | **Reports.** Saved list plus a pivot; value resolved per row from the entry's own rate. | gated           |
| `/time/workload`    | **Workload.** Capacity six weeks out. Replaces the `/team` stub and removes its flag.   | managers        |
| `/timeline`         | **Timeline.** Replaces the existing stub and removes its flag.                          | everyone        |
| `/w/:ws/p/:id/time` | **Project Time.** A fifth project tab.                                                  | project members |
| task detail         | The panel stays. Logging in context is correct; it gains category, billable and status. | everyone        |
| `/settings` → Time  | Week, rounding, targets, approval lock, categories, rates. Configuration only.          | admin/owner     |

Rationale for `/time` rather than `/time/timesheet`: the week grid is the thing
people open daily, so it earns the bare route. Review, Reports and Workload are
periodic and sit below it.

## Data model, the whole shape

Additions to `flow_time_logs`:

| Column                                                  | Why                                                                                                                                                                                                  |
| ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `entry_date DATE`                                       | The day the work happened, distinct from `logged_at` (when it was recorded). Without it a timesheet cannot exist, and today's form is stuck on "today".                                              |
| `category_id BIGINT NULL`                               | FK to `flow_time_categories`. Nullable, because existing rows have none.                                                                                                                             |
| `billable BOOLEAN NOT NULL DEFAULT 0`                   | Per entry, defaulted from the category.                                                                                                                                                              |
| `status ENUM('draft','pending','approved','rejected')`  | Default `draft`.                                                                                                                                                                                     |
| `approved_by BIGINT NULL`, `approved_at TIMESTAMP NULL` | Who cleared it and when.                                                                                                                                                                             |
| `locked_at TIMESTAMP NULL`                              | Set by approval-lock or the auto-lock job. A locked row is read-only.                                                                                                                                |
| `rate_snapshot DECIMAL NULL`                            | The rate **at the time of logging**. This is what makes Reports resolve value per row instead of multiplying a rate table client-side, and it is the one constraint the design calls non-negotiable. |

New tables:

- `flow_time_categories` — workspace-scoped: name, `billable_default`, archived flag. Deleting one with hours reassigns first.
- `flow_time_rates` — person-scoped hourly rate, with the per-project override column the design reserves.
- `flow_workspace_time_settings` — week start, rounding, weekly target, overtime threshold, `approval_locks_entry`, `auto_lock_days`. Replaces the browser-only store.
- `flow_timesheets` — user, `week_start`, status, `submitted_at`, `approved_by`, `approved_at`. The unit Review acts on.

New RPCs, per capability:

- `flow/time/v1` — `updateTimeLog` (the missing PATCH), `listTimeLogs` (by user and date range, paginated), `getTimeSummary`.
- `flow/timesheet/v1` — `getWeek`, `submitWeek`, `withdrawWeek`, `listForReview`, `approve`, `reject`. Bulk approve is **not** a transaction: partial failure names the entry that refused and keeps the successes.
- `flow/timereport/v1` — one grouped aggregation RPC (group by person, project, task, category, week), paginated, which Reports, Portfolios, the project Time tab and Workload's logged half all read.
- `flow/timesetting/v1` — read/write the workspace settings, categories and rates.

Every list RPC paginates. `paginationUtils.ts` exists in `rtylr` with zero call
sites; these are its first.

## Decomposition

Too large for one spec. Six sub-projects, each its own spec → plan → three PRs
(voxie, then onion, then rtylr, in that order because the generated clients are
committed).

**1. Entry richness.** The columns above, categories, rates and workspace settings,
`updateTimeLog`, and a paginated `listTimeLogs`. Frontend: the task panel gains
category, billable and status; Settings → Time drops the browser store for real
endpoints. _Unblocks everything else, and finishes the three screens already built._

**2. My week.** `/time` and the timesheet grid: `role="grid"`, arrow-key movement,
the ten-state cell, submit. Nothing in Flow has grid keyboard navigation to copy.

**3. Review.** `/time/review`, the flat queue, approve/return, the lock semantics,
and named partial failure on bulk actions.

**4. Reports.** The aggregation RPC, the pivot with both header axes scoped and a
sticky first column, saved reports reusing the Views pattern, and the project Time
tab.

**5. Workload and Timeline.** `/time/workload` and `/timeline`, both replacing
flag-gated stubs and removing `APP_FLOW_PREVIEW_ROUTES` from each.

**6. Cost, rates and billing surfaces.** Dual-gated on Flow admin **and** HR
payroll permission. Last, because it is the most sensitive and the Finance pull
contract is still open.

## What already-shipped work changes

Sub-project 1 revisits it, deliberately:

- `timeSettingsStore` loses its `persist` and becomes a query cache over the real
  endpoints. The seam did its job; it should not survive the endpoint.
- The task panel reinstates status pills, the Billable switch and category chips —
  now backed by columns, so they persist.
- The `GateNotice` reasons naming missing fields come out as the fields land.
- The board panel's WIP limit and prompts-for-hours need a column-level decision:
  they are still not in `flow_project_columns`. Either sub-project 1 adds them or
  they stay gated; flagged as an open question below.

Two fixes need none of this and are being done immediately: entry durations move to
clock notation (`3:15`, matching the mock and aligning the column), and dark
becomes the actual default, since `themeStore` boots `light` while every mock and
`HANDOFF.md` say dark is production.

## Risks

- **Three repos per sub-project.** A frontend change cannot merge before the
  generated client it imports. Order is contracts → backend → frontend, every time.
- **`flow_time_logs` has live rows.** Every added column is nullable or defaulted,
  and `entry_date` backfills from `DATE(logged_at)`.
- **HR already reads this table.** `internal/modules/hr_time/internal/infrastructure/flow_adapter.go`
  reads `flow_time_logs` today. Any column change has to be checked against it, and
  the reconciliation the handoff describes runs through it.
- **`status` plus `locked_at` plus the lock setting is a state machine.** It wants
  defining once, server-side, not re-derived per screen.
- **Scope.** Six sub-projects is months, not days. Sequencing by what unblocks the
  most is what keeps it shippable.

## Open questions

1. **WIP limit and prompts-for-hours** — add to `flow_project_columns` in
   sub-project 1, or leave the board panel's rows gated?
2. **Timer.** Still no start/stop schema. Ship the degraded quick-log the design
   settles on, or add `started_at`/`ended_at` and build it properly?
3. **Approval chain depth.** Single-level is baked in. Confirm that, or model a
   chain now while the tables are new?
4. **Business scoping.** `flow_time_logs.business_id` defaults to `0`, the same
   defect already recorded for the HR fixtures. Worth fixing here or out of scope?
