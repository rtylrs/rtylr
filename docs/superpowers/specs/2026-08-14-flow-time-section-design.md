# Flow Time: daily logging on one shared store

Date: 2026-08-14
Status: **proposal, awaiting approval**
Supersedes: the first draft of this file, which proposed a parallel Flow schema.

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

**One end-of-day reminder**, at a configurable time, then a manager-visible
missing-days list once someone passes a threshold. Within the six-a-week
notification budget the design sets.

## Non-employees

Contractors get a lightweight HR person record flagged non-employee, so they key
into `hr_time_entries` like everyone else and appear in the same reports, while
payroll excludes them by type. One store holds everyone; the type is what differs.

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
5. **Reports.** The grouped aggregation RPC, the pivot, the project Time tab.
6. **Workload and Timeline.** Both replace flag-gated stubs; both flags come out.
7. **Rates, cost and billing.** Dual-gated, last, and still blocked on the Finance
   pull contract.

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

## Open questions

1. **Who may log for someone else?** A manager filling in for an absent field
   worker is common, and nothing above covers it.
2. **A refused overtime request after the hours are worked.** The hours happened;
   refusing the request cannot delete them. Does the day stay unapproved, or does
   the time move to an unpaid category?
3. **Weekends and non-working days.** Does every calendar day need a record, or
   only days worked? Affects the missing-days list and how the week reads.
4. **Retention.** How long do work days and their action history survive, given
   they become payroll evidence?
