# FLOW->FIN-01: billable logged hours from flow to finance

- **Class:** CAN
- **Necessity:** Essential
- **Producer:** hr.time_entry.v1 (`TimeEntryPayload`), written by Flow
- **Consumer:** finance (invoicing)
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 1, a cross-domain read from the Finance app
- **Severity:** high

## Benefit

The day after this ships, a Finance user can see which billable hours have been
logged against a project and not yet invoiced, which is revenue the business
has earned and not asked for.

## Why this pairing

**Derivation and leakage.** Flow records hours. Finance issues invoices.
Nothing compares them, so unbilled time is invisible.

The producer side is unusually well prepared. Flow's work day lines persist
into `internal/modules/hr_time_entries` rather than a Flow-private table: the `workspaces`
module writes there directly, and the contract documents the intent
(`packages/voxie/api/flow/time/v1/service.onk:6`). Each entry already carries
`project_id`, `task_id`, `is_billable` and a `status`
(`packages/voxie/api/hr/time_entry/v1/models.onk:12`, `:14`, `:28`, `:30`), and
a `source` field whose values include `flow_import`
(`packages/voxie/api/hr/time_entry/v1/models.onk:35`). Billability is already a
first-class fact; nobody reads it.

**Why Finance and not Flow.** Flow could show "hours logged, not invoiced", but
Flow does not know what has been invoiced and should not: invoices are
Finance's. Putting the comparison in Finance keeps the direction
consumer-depends-on-producer and keeps invoice data where it belongs.

**Why this and not the labour-cost derivation.** `HR->FLOW-01` computes what
project time _cost_. This one is about what project time should have _earned_.
The second is cheaper, needs no new endpoint, and touches money the business is
owed rather than money it has already spent. That is why the roadmap sequences
this one first even though the brief's illustrative example was the other.

**What the user does instead right now.** Exports or reads time by hand and
compares against invoices by eye. There is even a backend endpoint built for
exactly this export and no UI for it: `GET /hr/time-entries/export-billable`
(`~/Development/onion/internal/modules/hr_time/internal/routes/handlers/export_billable_handler.go:15`)
returns per-entry `project_id`, `duration_minutes` and `is_billable`
(`~/Development/onion/internal/modules/hr_time/internal/domain/types.go:90`),
and **no frontend calls it** (zero rows in `internal/modules/inventory/endpoints.csv`). Somebody
already recognised this need and built half of it.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants using Flow with Finance, and logging billable time. The reference tenant, a services business with 6 active projects, is exactly this shape.                                                                                                                                                                                                                                                                                   |
| Manual work today      | a monthly reconciliation of logged hours against issued invoices, done by hand or not at all                                                                                                                                                                                                                                                                                                                                           |
| Modelled monthly value | **Leakage closed, not time saved.** Reference tenant: 40 people, assume 25 log billable time at 120 billable hours a month each, so 3,000 billable hours. At an assumed 60 USD blended billing rate that is 180,000 USD of billable time a month. If **1%** goes unbilled through oversight, that is **1,800 USD per month recovered**. The 1% is a modelled assumption; the arithmetic is shown so a reader can substitute their own. |
|                        | Time reclaimed is secondary: the reconciliation, when done, costs an administrator perhaps 90 minutes a month.                                                                                                                                                                                                                                                                                                                         |
| Build cost             | backend 1 day (add the filters in `06-proposed-contracts.md` item 1, plus pagination), frontend 2 days (an unbilled-time view and the comparison), maintenance ~0.5 days per year                                                                                                                                                                                                                                                      |
| Payback                | immediate on any tenant with real billable volume, given the value is recovered revenue rather than saved minutes.                                                                                                                                                                                                                                                                                                                     |
| Confidence             | **Inferred from code** that the fields, the write path and the unused export all exist. **Modelled from assumption** for every number above: hours, rate and the 1% leakage are illustrative, not measured. Do not quote the 1,800 USD as a finding.                                                                                                                                                                                   |
| Cost of not doing it   | n/a, this is CAN rather than MUST. The cost is opportunity: earned revenue that is never invoiced, and no way to notice.                                                                                                                                                                                                                                                                                                               |

## The data

| Field              | From             | Cited                                                  |
| ------------------ | ---------------- | ------------------------------------------------------ |
| `project_id`       | hr time entry    | `packages/voxie/api/hr/time_entry/v1/models.onk:12`    |
| `duration_minutes` | hr time entry    | `packages/voxie/api/hr/time_entry/v1/models.onk:20`    |
| `is_billable`      | hr time entry    | `packages/voxie/api/hr/time_entry/v1/models.onk:28`    |
| `status`           | hr time entry    | `packages/voxie/api/hr/time_entry/v1/models.onk:30`    |
| invoiced amounts   | finance invoices | `packages/voxie/api/finance/invoices/v1/service.onk:5` |

Formula, and the honest version of it:

```
unbilled_minutes(project, period) =
    sum(duration_minutes where project_id = P
                          and is_billable
                          and status = 'approved'
                          and date in period)
  - minutes already represented on invoices for P
```

The subtrahend is the problem: **an invoice has no project reference and no
line-level link to time entries.** So the first shippable version compares
_billable approved hours_ against _invoices the user associates with the
project by hand_, or simply reports billable approved hours with an
"invoiced?" flag the user sets. Claiming an automatic reconciliation would
require a `project_id` on invoices, which is the rejected `FIN->FLOW-01`
territory. **This dossier proposes the visibility, not the automatic match**,
and says so on the screen.

## What exists today

- Flow writes into `internal/modules/hr_time_entries` (verified across
  `~/Development/onion/internal/modules/workspaces/sql/queries/worktime.sql`,
  whose statements touch `flow_work_days`, `flow_work_types`, `internal/modules/hr_employees`
  and `internal/modules/hr_time_entries`).
- `is_billable` and `project_id` are contract fields, populated.
- `ListTimeEntriesRequest` exists (`packages/voxie/api/hr/time_entry/v1/service.onk:6`).
- An unused billable-export route exists, described above.
- **Manual workaround:** eyeball comparison.

## What is missing

| Wanted                                              | Contract reality                  | Backend reality             | What is missing                                                                                                                                                                 |
| --------------------------------------------------- | --------------------------------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Billable approved hours for a project over a period | fields all exist                  | rows exist, written by Flow | `ListTimeEntriesRequest` filters only by `employee_id` and one exact `date` (`packages/voxie/api/hr/time_entry/v1/models.onk:71`). No range, no `project_id`, no `is_billable`. |
| Ask for a page of them                              | response reports `limit`/`offset` | ignored                     | the request cannot set them: one of the 13 HR endpoints in `04-gaps.md`                                                                                                         |
| Automatic match to invoices                         | invoices have no project          | none                        | out of scope, and stated as such on the screen rather than faked                                                                                                                |

**Bucket: needs a contract field**, and specifically query filters. Draft in
`06-proposed-contracts.md` item 1.

## Precedent in the codebase

Flow already reads the HR domain 11 times (`01-inventory.md`), so a Flow/HR/
Finance triangle is not novel territory. The precedent that matters more is the
cautionary one: `internal/modules/hr_category_id` is a Flow-to-HR link that was declared and
never consumed (`02-entity-graph.md`). This proposal is the opposite bet, since
every field it needs is already populated by a live write path, and the only
missing piece is the ability to ask for them.

## How to build it

**Shape 1, a cross-domain read.** Finance calls
`hr.time_entry.v1 listTimeEntries` with a project and date range.

**Why shape 1 is sufficient.** The Finance unbilled-time screen reads one
foreign domain (hr). Two would be the limit; one is fine. The data is not
more sensitive than what a Finance user already sees, with one exception noted
below.

**Why not shape 2, an aggregate.** It would be justified if the screen needed
several foreign domains or if hours had to be priced, which is
`HR->FLOW-01`'s job and carries the salary risk. Reporting minutes needs no
aggregate. If a later version prices the hours, it must switch to shape 2 for
the reason given in `04-gaps.md`.

**Why not shape 3, an event.** Nothing is triggered. A user is asking a
question.

Order of work: contract filters and pagination first, then the screen. Do not
build the screen against the unpaginated endpoint.

## Coupling cost

- **New dependency:** finance depends on hr time entries. Edge `hr -> finance`.
- **Direction:** consumer depends on producer. Correct. Note the edge is
  labelled `hr -> finance` even though the data is Flow's, because the contract
  and the rows belong to hr. That matters for ownership: **hr owns this
  contract**, and a Flow change cannot break Finance without going through it.
- **Cycle:** none. `finance -> hr` is marked `would-invert` in the matrix and
  is not proposed.
- **Owner:** **hr** owns `time_entry.v1`. Finance owns the screen.
- **Degrade when Flow is disabled:** the entries simply do not exist, so the
  screen shows an empty state saying no billable time has been logged. Finance
  never errors. When **HR** is disabled the endpoint is unavailable, and the
  section must hide itself rather than error: check entitlement through
  `packages/config/src/appGuards.tsx`, not by catching a 403.

## Risks and correctness notes

- **Permissions, and the one real widening.** A time entry carries an
  `employee_id` and a free-text `description` of what someone did. Exposing
  raw entries to a Finance user shows them who worked on what and their own
  words about it. That is more than "hours per project". **The screen should
  aggregate by project and only drill into per-person detail for roles that may
  see it.** No rates cross, so this is not the salary exposure that
  `HR->FLOW-01` risks, but it is not nothing.
- **Tenancy.** Server-side scoping already applies; the client must not assume
  a filtered list is complete without paging.
- **Timezone.** `date` is the date the work was performed and Flow records a
  per-record timezone on the work day
  (`packages/voxie/api/flow/time/v1/models.onk:48`). A period boundary can
  therefore include or exclude a travelling worker's day depending on
  interpretation. The screen must state which timezone the period is expressed
  in.
- **Status.** Only `approved` entries should count as billable-and-ready.
  Counting drafts would invoice hours nobody has confirmed.
- **Rounding.** Minutes are integers; no rounding until money is involved, and
  this connection stops before money.
- **Currency, historical rates:** not applicable, no pricing here.

## Effort

**S/M.**

| Part                | Estimate                                                                                                                                 |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 1 day: add `from_date`, `to_date`, `project_id`, `is_billable`, `status`, `limit`, `offset` to `ListTimeEntriesRequest`, and honour them |
| Frontend            | 2 days: unbilled-time view grouped by project, period picker, the honest "not automatically reconciled" statement                        |
| Ongoing maintenance | ~0.5 days per year. Slightly higher than a plain read because the aggregation logic lives in the client.                                 |

## Backend section (round two)

Two developments, one closing the gap and one widening it.

Closing: the missing filters landed upstream the day after round one filed.
voxie #46 added `from_date`, `to_date`, `project_id`, `is_billable`, `status`,
`limit` and `offset` to `ListTimeEntriesRequest` (`api/hr/time_entry/v1/models.onk`),
onion #783 served them, packages #66 and rtylr #917 bumped the pointer through.
The landed contract mirrors round-one draft 1 nearly verbatim. The gap bucket
moves from "needs a contract field" to frontend work.

Widening: `internal/modules/hr_time_entries.billable_invoice_id`
(`internal/platform/database/schemas/hr_time/schema.sql:22`) is read by the unbilled filter
(`internal/modules/hr_time/sql/queries/time_entries.sql:52`: `AND billable_invoice_id IS NULL`) and written by
nothing anywhere in the repo. "Unbilled" is currently true of every row forever;
recording that an hour got invoiced exists at no layer. Any implementation of
this dossier must include the invoice-side marker write or the unbilled number
never moves.

_Round two (2026-08-22): corrected-by-progress upstream plus a new storage gap
recorded in 14-layer-reconciliation.md A1/A2._
