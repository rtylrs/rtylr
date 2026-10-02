# Flow Time: daily logging on one shared store

Date: 2026-08-14
Status: **proposal, awaiting approval**
Supersedes: the first draft of this file, which proposed a parallel Flow schema.

## Status: step 1 is built

Work days, lines and work types ship across four repos, verified end to end
against a locally running API. Steps 2 to 8 are unstarted.

**The chain is four repos, not three.** `internal/platform/database/schemas` in
onion is itself a git submodule (`tryrtylr/schemas`), so the order runs
`schemas` then `voxie` then `onion` then `rtylr`, with `packages` alongside for
shared frontend code. An earlier draft of this document said three, and a
migration written into that path in a fresh worktree is silently untracked.

## What this corrects, twice

**First**, time tracking was built as configuration: a settings pane and two panels
tucked inside other screens. Logging hours is a primary Flow capability and needs
its own section with real depth.

**Second**, and larger: the first draft proposed adding status, billable, category,
approval and timesheet tables to Flow. **Nearly all of that already exists in
`hr_time_entries`.** Building it again would have created a second source of truth
for hours in a product that already struggles with one.

What `hr_time_entries` already has: `project_id`, `task_id` (nullable, so
project-level logging works today), `category_id`, `date`, `duration_minutes`,
`start_time`, `end_time`, `description`, `is_billable`, `status`,
`rejection_reason`, `approved_by`, `source`, `billable_invoice_id`.
`hr_timesheets` + `hr_timesheet_approval_actions` already do **multi-level**
approval, which the design doc claimed was not expressible.

`hr_time/internal/infrastructure/flow_adapter.go` already reads `flow_time_logs`,
so today there are two stores and an adapter between them. This proposal ends
that: **Flow writes `hr_time_entries`.** That is HANDOFF's open "D1, the shared
time service", decided.

## The shape of a day

Two levels, because a person has one working day containing several pieces of work.

```
Work day        date · work mode · start · end · breaks · timezone · status
  └── lines     project · task · work type → subtype · description · hours · billable
```

The day is new and Flow-owned. The lines are `hr_time_entries`.

### Why the day is not `hr_attendance_records`

That table models live clocking against a shift: `is_late`, `late_minutes`,
`clock_in_method`, `clock_in_lat/lng`, and it feeds HR's lateness policy. People
entering yesterday retrospectively would generate meaningless lateness data and
could trip attendance rules they were never subject to. A work day is a work log,
not a punch clock, and says so.

## Schema

**New: `flow_work_days`**

| Column                                                     | Note                                                                     |
| ---------------------------------------------------------- | ------------------------------------------------------------------------ |
| `employee_id`                                              | Everyone has an HR person record, including non-employees (below)        |
| `date`, `timezone`                                         | The person's own zone, stored per record, so a day means their local day |
| `work_mode`                                                | `onsite` / `remote` / `hybrid`, the day's default                        |
| `start_time`, `end_time`                                   | Optional; a day can be lines-only                                        |
| `break_minutes`                                            | Unpaid, so paid hours are correct                                        |
| `status`                                                   | `draft` / `submitted` / `approved` / `rejected` / `reopened`             |
| `approved_by`, `approved_at`, `reopened_by`, `reopened_at` |                                                                          |

Unique on `(owner_id, employee_id, date)`.

**New: `flow_work_day_actions`**. Approve, reject, reopen, with actor and
comment. Mirrors `hr_timesheet_approval_actions` rather than reusing it, whose
`timesheet_id` means something else.

**New: `flow_work_types`**. Workspace-scoped, `parent_id` self-reference for
exactly two levels, `hr_category_id` mapping to HR's payroll category, plus
archived and sort order. Admin or manager maintains the tree.

**Altered: `hr_time_entries`** (all nullable or defaulted; the table has live rows)

| Column         | Why                                                                                    |
| -------------- | -------------------------------------------------------------------------------------- |
| `work_day_id`  | Parent day. Null for legacy rows.                                                      |
| `work_type_id` | Flow's two-level type. `category_id` stays as the HR payroll mapping.                  |
| `work_mode`    | Null means inherit the day; set means this line differed.                              |
| `timezone`     | As above.                                                                              |
| `client_key`   | Client-generated, unique per owner. The idempotency key that makes offline retry safe. |

**Altered: projects and tasks**. `default_work_type_id`, so the common case
pre-fills and a phone entry is two taps.

## Rules

**The gap warns, never blocks.** Clocked hours minus breaks, against the sum of the
lines. The unaccounted figure is shown prominently and is reportable, and
submission proceeds regardless. Blocking is how timesheets get padded with invented
lines, which is worse than an honest gap.

**Approval is per day, after the day has ended.** A day cannot be approved while it
is still running. A manager can reopen an approved day; any edit to a reopened day
returns it to the queue, and every transition lands in `flow_work_day_actions`.
`hr_timesheets` stays as the weekly rollup payroll consumes, derived once its days
are approved: the day is the decision, the week is the total.

**Overtime raises a real request.** A day past the threshold creates an
`hr_overtime_request` rather than a warning nobody records. Open question 2 below
covers what happens when one is refused after the work is done.

**Offline merges per line.** Each line syncs independently on its `client_key`, so
two devices adding different lines both succeed. Only genuine edits to the same
line conflict, and those surface for the person to resolve. Nothing is silently
discarded, because for logged work that means losing someone's afternoon.

**A manager may log on someone's behalf.** Field workers lose phones and go absent,
and the hours still happened. The entry records who it was logged for and who
logged it, and both appear on the day, so a manager's entry is never mistaken for
the person's own.

**A refused overtime request leaves the day unapproved.** The hours were worked and
cannot be deleted, so the day stays in the queue with the refusal attached until a
manager resolves it. This is deliberately uncomfortable: an unapprovable day is a
conversation, and silently paying or silently discarding the time are both worse.

**Only days worked need a record.** There is no expectation of a row for a day off,
so the missing-days list has to know a person's working pattern rather than
assuming Monday to Friday. HR already models this in `hr_shifts` and
`hr_shift_assignments`, and the list reads from there.

**One end-of-day reminder**, at a configurable time, then a manager-visible
missing-days list once someone passes a threshold. Within the six-a-week
notification budget the design sets.

## Non-employees

Contractors get a lightweight HR person record flagged non-employee, so they key
into `hr_time_entries` like everyone else and appear in the same reports, while
payroll excludes them by type. One store holds everyone; the type is what differs.

## Plant cost

Project cost is labour plus plant. Labour comes from the lines above; plant needs a
register, rates and usage.

### Register

`hr_assets` stays the single spine for things the business owns: name, category,
serial, purchase price. It is an IT and personal-kit register today, with no rate,
no usage and no project allocation, and `assigned_to` points at a person, which
fits a laptop and not an excavator.

**New: `flow_plant_items`**, referencing `hr_assets` where the item is owned, and
standing alone where it is hired in. Asset identity is not duplicated; costing
attributes hang off it.

| Column                                  | Note                                                                                   |
| --------------------------------------- | -------------------------------------------------------------------------------------- |
| `asset_id`                              | Null for hired plant that the business does not own                                    |
| `ownership`                             | `owned` / `hired`                                                                      |
| `supplier_id`, `hire_rate`, `hire_unit` | Hired items only. Sets up invoice reconciliation later without building it now.        |
| `operating_rate_hour`                   | Cost while working                                                                     |
| `standby_rate_hour`                     | Cost while on site and available. Lower, and the reason idle plant stops looking free. |
| `logging_mode`                          | `operator` / `site_manager` / `both`, which decides who is asked to record it          |

### Usage

**New: `flow_plant_usage`** with `plant_item_id`, `project_id`, optional `task_id`,
`date`, `hours`, `mode` (`operating` / `standby`), optional `operator_employee_id`,
`logged_by`, `work_day_id` where it came in on someone's day, and a `client_key`
for the same offline sync as labour lines.

Two ways in, by item:

- **Operated plant** appears as a second kind of line on the operator's own work
  day. One habit, one screen, and it syncs offline like everything else.
- **Static or shared plant** is recorded by a site manager, per site per day, which
  is the only way standby hours on unmanned items get captured at all.

### Cost

```
project cost = labour   sum(hours x cost rate)
             + plant    sum(operating hours x operating rate)
             +          sum(standby hours   x standby rate)
```

Hired items carry their hire rate as real cash out, so a job can show recovered
cost against actual spend rather than only an internal figure. As with labour
rates, the rate is snapshotted on the usage row: a rate change must not silently
restate last quarter's job costs.

Cost is money, so every plant cost surface sits behind the same dual gate as
labour cost.

## Screens

The sidebar Time group gets real destinations. Settings keeps configuration only.

| Route               | Screen                                                                                                 |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| `/time`             | **My day.** Today first: work mode, start/end/breaks, the lines, the gap. The week is a tab beside it. |
| `/time/review`      | **Review.** Days awaiting approval, oldest first, with reopen.                                         |
| `/time/reports`     | **Reports.** Grouped by person, project, work type, mode, week.                                        |
| `/time/workload`    | **Workload.** Replaces the `/team` stub; removes its flag.                                             |
| `/timeline`         | Replaces the existing stub; removes its flag.                                                          |
| `/w/:ws/p/:id/time` | Project Time, a fifth tab.                                                                             |
| task detail         | The existing panel, gaining work type and billable.                                                    |
| `/settings` → Time  | Week, rounding, targets, work-type tree, rates, reminder time.                                         |

`/time` leads with the day rather than the week because the requirement is daily
logging. The week grid remains, one tab across.

## Build order

Each step is contracts → onion → rtylr, in that order, because the generated
clients are committed.

1. **Work day + lines.** `flow_work_days`, `flow_work_types`, the `hr_time_entries`
   columns, and the RPCs to read and write a day. Frontend: `/time` for today, and
   the task panel gains work type. _Delivers daily logging end to end._
2. **Approval.** Day submit, approve, reject, reopen, and `/time/review`.
3. **Offline and mobile.** `client_key` sync, the queue, conflict resolution, and
   the phone entry form as a first-class surface.
4. **Reminders and missing days.** End-of-day nudge, escalation list.
5. **Plant.** The register, rates, usage capture on the work day and the site-manager
   screen. Project cost becomes labour plus plant.
6. **Reports.** The grouped aggregation RPC, the pivot, the project Time tab, and
   cost by project across both labour and plant.
7. **Workload and Timeline.** Both replace flag-gated stubs; both flags come out.
8. **Billing.** Dual-gated, last, and still blocked on the Finance pull contract.

## What this unwinds

Deliberately, in step 1: `timeSettingsStore` loses its `persist` and becomes a
query cache over real endpoints; the task panel's `GateNotice` reasons come out as
the fields land; and `flow_time_logs` is migrated into `hr_time_entries` and
retired, along with the `flow_adapter` join that exists to bridge them.

## Risks

- **`flow_time_logs` has live rows and an external reader.** Migration has to
  preserve them and cut over `flow_adapter.go` in the same change.
- **Flow gaining a dependency on the HR module.** Direction of dependency in a
  modular monolith is a real architectural call, not a detail.
- **`business_id` defaults to `0`** on both `flow_time_logs` and `hr_time_entries`,
  the same defect already recorded for the HR fixtures. Tenancy scoping needs
  settling before more data lands on it.
- **Two-level approval semantics.** Day-level approval plus a weekly payroll
  rollup means two states that can disagree; the rollup must be derived, never
  independently editable.
- **Scope.** Seven steps across three repos. Step 1 alone is a genuine feature.

## Retention

Work days, their action history, labour lines and plant usage are retained: they
are payroll and job-cost evidence, and a job's cost has to be defensible after the
job closes. The retention period itself is still to be set, and it is a legal
question per jurisdiction rather than a product one.

## Decisions on the four items left open

### 1. Finance shows the hire variance. Flow supplies the evidence.

There is already a precedent and this follows it: Finance pulls approved time, and
Flow renders no invoice and has no send button. Hired plant works the same way.
Flow exposes usage per hired item per period through the same aggregation endpoint
as everything else; Finance's bill screen pulls it and shows expected against
invoiced at the moment someone decides to pay or query.

The variance belongs where the decision is made, and that is the bill screen, not a
plant report. Flow never learns what an invoice says.

What Flow does owe is the thing only a site manager can answer: whether the item
was really on site for that long. That is the usage record, and it has to be
correctable after the fact, because an invoice query usually starts with someone
saying the days are wrong.

### 2. Fuel is out of scope, and will arrive as an expense, not a plant column.

Fuel is real cost and leaving it out understates every job. It is still the wrong
thing to build here, because fuel is a purchase: it arrives on a receipt, against a
site or a vehicle, and `expenses` already models that. Adding a fuel field to plant
usage would create a second, worse place to record spending that Finance already
captures.

The insurance is cheap and taken now: **project cost is an additive set of typed
components**, not a single computed figure.

```
project cost = sum(component)   where component is one of
               labour · plant-operating · plant-standby · [consumables] · [expenses]
```

Adding fuel later is then a new component type reading project-tagged expenses, not
a rewrite of the cost model. Every cost surface is built to render an open set of
components from the start.

### 3. Split the register by field, not by screen.

Rates are money and fleet admin is not, so gating the whole register on the money
permission makes a fitter and an accountant the same person. Two capabilities:

- **`manage plant`** adds and edits items, serial numbers, status, availability and
  who logs each one. Fleet and site managers hold it.
- **`manage plant rates`** sets operating, standby and hire rates, dual-gated
  exactly like the labour cost surfaces.

A manager can put a new excavator on the register and mark it available this
morning; what it costs is set by someone who is allowed to see money. This is the
same line the design already draws between time and cost.

### 4. New tables carry no `business_id` default. The existing one gets its own fix.

`business_id BIGINT NOT NULL DEFAULT 0` on the existing tables is a tenancy hazard:
a forgotten value is not an error, it is row zero, silently shared. Three rules
here:

1. **New tables declare `business_id BIGINT NOT NULL` with no default**, so an
   insert that forgets it fails loudly instead of landing in a shared bucket.
2. **Every write sets it explicitly**, including writes into `hr_time_entries`,
   whose default this work must not rely on.
3. **Every read filters on it.** Scoping by `owner_id` alone is what turns a
   default into a cross-tenant leak, and indexes lead with
   `(owner_id, business_id, ...)` accordingly.

Correcting the existing default is **not** folded into this work. It needs a data
audit first, because deciding which business owns today's row-zero data is a
question about the data, not the schema, and doing it blind would move real rows to
the wrong tenant. It gets its own change, and this spec deliberately stops adding
to the problem in the meantime.

## Still genuinely open

1. **The retention period.** Retention is settled; the number is a jurisdictional
   question and needs an answer per market rather than a default invented here.
2. **Direction of dependency.** Flow taking a dependency on the HR module, and now
   possibly on ERP for supplier links, is an architectural call that wants an
   explicit yes before migrations are written.

## Amendment (2026-08-16): step 1 didn't meet the original ask

Step 1 shipped and was verified end to end, but it wasn't what was asked for.
Three specific gaps, confirmed directly rather than guessed at:

1. **Project was never surfaced**, though the field already exists.
2. **Time still reads as settings-adjacent**, not a first-class section.
3. **The screen is thin** for something called a main functionality: no history,
   no team visibility, no reports.

This section scopes the fix for all three as one build. It does **not**
renumber steps 2–8 above; it sits alongside them, and mostly ahead of them,
because two of these three are close to free and the third exposes a real gap
step 1 left open.

### 1. Project: already there, just not wired

`hr_time_entries.project_id` exists, is nullable, and the entire chain already
carries it: `WorkDayLine.project_id` is in the voxie contract
(`api/flow/time/v1/models.onk:84`), onion's domain struct, repository and sqlc
queries all read and write it
(`internal/modules/workspaces/internal/infrastructure/worktime_repository.go:107,244,273`).
`apps/flow/src/features/time/useWorkDay.ts` already accepts and forwards a
`projectId`. **No schema, contract, or repository change is needed.** The gap
is exactly one thing: `MyDay.tsx` never rendered a picker and never required
one.

Fix: add a required project picker to the add-line form, reusing
`apps/flow/src/features/projects/useProjects.ts` (already used elsewhere in
Flow) rather than a new query. Required in the UI. The column itself stays
nullable — `hr_time_entries` has other writers, and narrowing a shared column
is a separate, larger decision than this screen. What "required" means here is
enforced where it should be: the Flow work-day-line create/update path in
onion's `application` layer rejects a missing `project_id`, the same layer
that already owns `ErrInvalidWorkMode` and friends. A gap in adoption is not
grounds to leave it in the frontend only, or the constraint stops meaning
anything the moment a second client exists.

No backfill: there is no production data on this table's Flow-owned columns
yet, confirmed directly rather than assumed. Nothing to migrate.

### 2. A real security gap, found while scoping the team view

Designing "who can see whose day" surfaced that **nobody can see whose day is
already unenforced**. `WorkTimeAPI.scope()`
(`internal/modules/workspaces/internal/api/worktime_service.go:33`) resolves
tenant scope and then trusts a caller-supplied `employee_id` with no
capability check at all. Any authenticated user in the business can read or
write any other employee's work day today, by passing their id — GetWorkDay,
ListWorkDays and CreateWorkDayLine all take this path. This predates the team
view; it was already true for the "manager logs on someone's behalf" case in
step 1, which was built assuming a permission check would arrive with it and
one never did.

The frontend hid this by hardcoding the gate closed:
`TaskDetail.tsx:648` passes `canLogForOthers={false}`, a literal, not a
computed permission — so the one screen in step 1 that could have exercised
this path never did, and the hole went unnoticed. That is what live
verification without a second account will miss: the button was correctly
disabled, for the wrong reason.

**Shipped** (`3d754ec7`): `scope()` now denies any `employeeID` argument that
differs from the caller, unconditionally. Not "unless the caller holds a
permission" — investigating what that permission would check against found
there is nothing to check yet (next section), and closing a live hole should
not wait on the feature that will eventually relax it. This is a correctness
fix independent of everything else in this amendment, landed first, on its
own commit.

### 3. One new permission, and it is not where the first draft of this section said

The first draft of this section proposed adding `flow:time` to
`@voxire/permissions` (`packages/permissions/src/capabilities.ts`), the
business-wide `Resource × Action` matrix configured through Dash's
`/permissions` and `/roles` screens. Investigating how to enforce it
server-side found the premise wrong: **that matrix is not enforced anywhere in
onion.** It is stored correctly - the `roles` module's `AllPermissions()` and
`UpdatePermissionMatrix` genuinely persist a business's role → capability
grid - but nothing in the backend ever reads it to gate a request. Confirmed
by grepping every other module for an import of `modules/roles`: there are
none. Adding `flow:time` there would have shipped a Dash toggle that looks
like it restricts who sees a colleague's hours and does not, which is worse
than no control at all - a toggle a business owner unchecks believing access
is now denied is a lie, not a gap.

What is actually enforced, anywhere in this backend, for anything resembling
"does this caller have standing over that resource":

- `hr_payroll`'s single `manage_payroll` check, via `hr_core`'s freeform
  per-employee permission blob. Narrow, payroll-only, not reusable.
- Flow's own `Service.RequireProjectRole` / `RequireWorkspaceRole`
  (`internal/modules/workspaces/internal/application/service.go:269-286`),
  gating on `flow_project_members.role` / `flow_workspace_members.role`
  (`viewer` < `member` < `admin`). Real, tested
  (`internal/modules/workspaces/internal/api/task_role_test.go`), and already
  used by every mutating Flow task/comment/attachment/time-log handler.

Decision: model `manage` and `view_team` as workspace-role requirements using
the pattern that is real, not the one that only looks real. Concretely:

- **Viewing someone else's day** requires the caller hold at least
  `FlowRoleAdmin` on a workspace the target employee is also a member of.
  `RequireWorkspaceRole` already exists; what's new is resolving an
  `employee_id` to the workspaces they belong to, which needs the
  `hr_employees.user_id` link (confirmed unique) since workspace membership is
  keyed on auth `user_id`, not HR `employee_id`. This resolution does not
  exist in the workspaces module today and is new plumbing, not a config
  toggle - which is exactly why it lands with the team view (next section)
  rather than as a standalone permission commit: there is nothing to gate
  until that resolution exists.
- **Logging for someone else** (`canLogForOthers`) waits for the same
  resolution rather than getting its own shortcut. `scope()`'s unconditional
  deny above already closes the hole; loosening it correctly and loosening it
  first are different things, and only the first is safe to do under time
  pressure.

This is a real, product-wide gap worth its own attention independent of Flow
time: every other resource in `capabilities.ts` - `hr:staff:manage`,
`erp:inventory:adjust`, `crm:customers:manage`, `finance:invoices:create`, all
of it - is equally unenforced. Flagged separately; not this amendment's job to
fix the whole product's authorization story.

### 4. Team/manager view: scoped by workspace, not HR reporting lines

Flow's grouping unit is the workspace (`flow_workspace_members`,
`api/flow/workspace/v1/models.onk`), not an HR org chart — a manager's "team"
for this view is the set of people in their Flow workspace, which is the
scope every other Flow screen already uses. `WorkspaceMember.user_id` is an
auth user id; work days key on HR's `employee_id`. The two are already linked
by `hr_employees.user_id` (unique), so resolving a workspace's member list to
the set of employee ids to query is a lookup, not a new concept.

New: `ListWorkDaysForWorkspace(workspace_id, from, to)` in the same
`TimeService`. Requires `Service.RequireWorkspaceRole(ctx, workspaceID,
callerID, domain.FlowRoleAdmin)` - the same real check every other
Flow-mutating handler already uses, not a new permission system. Resolves
workspace members to employee ids server-side via `hr_employees.user_id` (not
a client-side loop calling `ListWorkDays` once per person - that's an N+1 the
client has no business making), then returns one page of work days across all
of them for the range. New screen: `/time/team`, a roster with one row per
member showing today's status (logged / gap / not started) and a date picker
for other days, reusing the same `WorkDay` shape the My Day screen already
renders.

### 5. Week/history view: mostly free

`ListWorkDaysRequest` already takes `from`/`to`. The week view is a date-range
call to the existing endpoint the caller already has access to for their own
`employee_id` — no backend change. New: a week-strip control on `/time`
alongside the existing day view, and the route keeps working for a single day
when no range is selected, so nothing about today's screen changes for
someone who never touches it.

### 6. Reports v1: day-level only, and grouping by project needs the RPC this section said it didn't

The first draft of this section proposed grouping by `project_id` /
`work_type_id` client-side from `ListWorkDaysForWorkspace`. Building it found
that premise wrong too: `WorkDay` deliberately excludes lines
("the lines are not loaded: a week view needs each day's totals, not every
line on every day" - the same comment on `ListWorkDays` this RPC reuses), and
`project_id`/`work_type_id` live on `WorkDayLine`, not `WorkDay`. There is no
endpoint that returns lines across a range or a workspace, only per-day
(`GetWorkDay`) - calling it once per work day to build a report would be a
real N+1, not a shortcut.

**Shipped instead:** `/time/reports` groups only what `ListWorkDaysForWorkspace`
actually carries - clocked, logged and gap minutes per employee, summed
across a chosen range. Grouping by project or work type is not drawn disabled
with a reason the way a missing permission would be; it is absent from the
screen entirely, because the concept genuinely needs data this call cannot
provide, not a permission this user lacks. That grouped aggregation endpoint
is still step 6 of the original build order - this did not shrink it, it
confirmed it is real, separate work, and this screen's data source swaps to
it once it exists, same seam as `timeSettingsStore`.

### Build order for this amendment

1. **Shipped: the `scope()` fix, alone** (`3d754ec7`). An unconditional deny,
   not gated on a permission, because there was nothing real to gate it on.
2. **Shipped: project picker**, required in the UI, enforced in onion's
   application layer (`ca4ba95a` + the matching rtylr commit). Turned out to
   need no schema or contract change at all - the field already existed
   end to end.
3. **Shipped: `ListWorkDaysForWorkspace` + `/time/team`.** This is where the
   employee-id-to-workspace resolution actually got built, using
   `RequireWorkspaceRole` - the real permission work, folded into the feature
   that needed it rather than shipped as a standalone step with nothing to
   attach to yet. Also added `WorkDay.user_id`, needed for the frontend to
   join a day back to the workspace roster it already has. `canLogForOthers`
   still hardcodes false - it needs the same resolution, applied to a single
   employee rather than a whole roster, not yet built.
4. **Shipped: week strip on `/time`.** No backend change; `startOfWeek`/
   `endOfWeek` gained an optional `weekStart` so the strip honors the
   workspace's configured week start.
5. **Shipped: `/time/reports`**, day-level totals only - see the corrected
   section above for why project/work-type grouping did not ship with it.

4 and 5 are independently shippable once 3 lands; 1 and 2 already have.
