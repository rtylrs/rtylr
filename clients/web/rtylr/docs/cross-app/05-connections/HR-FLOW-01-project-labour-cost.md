# HR->FLOW-01: employee pay rate from hr to flow, as project labour cost

- **Class:** CAN
- **Necessity:** High value
- **Producer:** hr (`internal/modules/hr_employees.hourly_rate` plus `hr.time_entry.v1`)
- **Consumer:** flow (project pages), and Insights via the same endpoint
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 2, a backend aggregate endpoint owned by HR
- **Severity:** high

This is the case the brief opened with. It is real, and the audit found it to be
cheaper than expected and **not** the highest-leverage item. See `08-roadmap.md`.

## Benefit

The day after this ships, a project page shows what the time logged against it
cost in labour, so a services business can tell a profitable project from an
unprofitable one without a spreadsheet.

## Why this pairing

**Derivation.** Hours live in HR time entries with a `project_id`
(`packages/voxie/api/hr/time_entry/v1/models.onk:12`). Rates live on the
employee record (`~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:91`).
Nothing multiplies them per project.

Payroll already does the multiplication for pay:

```
basePay = hoursWorked * hourlyRateDollars
```

`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:221`

but it uses `internal/modules/hr_timesheets.total_minutes`
(`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:186`),
which is supplied by the client rather than aggregated from time entries
(`~/Development/onion/internal/modules/hr_time/internal/infrastructure/repository.go:215`).
So the arithmetic exists and the project dimension exists, in two different
places that never meet.

**Why HR owns the endpoint, not Flow.** This is a permissions decision, not a
convenience. If Flow read employee records and multiplied client-side, every
Flow workspace member with a browser could read colleagues' pay rates: `hr.Staff`
exposes `salary`, `hourly_rate` and `overtime_rate`
(`packages/voxie/api/hr/staff/v1/models.onk:17`). HR must return the total and
never the inputs.

**Why one endpoint serves Flow and Insights.** `HR->INS-01` wants the same
number for margin reporting. Two endpoints would eventually give two answers to
one question.

**What the user does instead right now.** Exports hours, exports or recalls
rates, multiplies in a spreadsheet. This is the exact workflow the brief
described.

## ROI

| Field                  | Value                                                                                                                                                                                                                     |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running Flow with HR and logging project time. The reference tenant, 6 active projects, exactly.                                                                                                                  |
| Manual work today      | a spreadsheet per project per period, or the question goes unasked                                                                                                                                                        |
| Modelled monthly value | reference tenant, 6 active projects. Building the cost picture by hand costs perhaps 45 minutes per project per month: **6 x 45 = 4.5 hours per month**. At a loaded 40 USD/hour for the person doing it, ~180 USD/month. |
|                        | The larger value is the decision enabled: which projects lose money. Unquantifiable here without pricing data, and deliberately not invented.                                                                             |
| Build cost             | backend **4 days** (new `hr.cost.v1` capability, the aggregate query, rate-coverage handling, role gating), frontend 2 days (Flow project cost panel), maintenance ~0.5 days per year                                     |
| Payback                | ~3 months on the time figure alone for the reference tenant; faster for any tenant with more projects.                                                                                                                    |
| Confidence             | **Inferred from code** that hours carry projects, rates exist, and payroll already multiplies. **Modelled from assumption** for the 45 minutes and the 40 USD.                                                            |
| Cost of not doing it   | n/a, CAN not MUST. Project profitability stays a manual exercise.                                                                                                                                                         |

## The data

| Field              | From          | Cited                                                                          |
| ------------------ | ------------- | ------------------------------------------------------------------------------ |
| `project_id`       | hr time entry | `packages/voxie/api/hr/time_entry/v1/models.onk:12`                            |
| `duration_minutes` | hr time entry | `packages/voxie/api/hr/time_entry/v1/models.onk:20`                            |
| `employee_id`      | hr time entry | `packages/voxie/api/hr/time_entry/v1/models.onk:10`                            |
| `hourly_rate`      | hr_employees  | `~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:91` |

Formula:

```
cost(project, period) = sum over entries e in project, period of
    (e.duration_minutes / 60) * rate(e.employee_id)
```

**`rate` is a function of time and the design must say which time.** An
employee's rate today is not the rate they were on last quarter. The honest
first version uses the current rate and labels the figure as valued at current
rates, because there is no rate history table to read
(`internal/modules/hr_employees.hourly_rate` is a single current value). Claiming historical
accuracy would be a fabrication. See the risks section.

## What exists today

- Time entries with `project_id`, populated by Flow via `flow_import`
  (`packages/voxie/api/hr/time_entry/v1/models.onk:35`).
- `hourly_rate` and `overtime_rate` on the employee record.
- Payroll's own hours-times-rate computation, for pay rather than costing.
- **Nothing** that groups cost by project. Searched the payroll, `flow_insights`
  and `insights` modules; no labour-cost-per-project aggregation exists.
- **Manual workaround:** spreadsheets.

## What is missing

| Wanted                  | Contract reality                          | Backend reality                                                       | What is missing                                                                   |
| ----------------------- | ----------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Labour cost per project | no cost-per-project message in any domain | every input present; nothing joins hours-by-project to rate-by-person | a new HR-owned aggregate RPC (`06-proposed-contracts.md` item 8a)                 |
| Without leaking rates   | `hr.Staff` exposes pay fields             | served                                                                | the aggregate must return totals only, and be gated                               |
| Honest partial totals   | nothing                                   | employees may have no rate                                            | a `rate_coverage` figure so a partial total is labelled, not silently understated |

**Bucket: needs a query or endpoint.**

## Precedent in the codebase

The precedent is the reason to be careful: `flow.WorkType.hr_category_id` was
added to map Flow work types onto HR payroll categories "so one entry serves
both project reporting and payroll", and **nothing reads it**
(`packages/voxie/api/flow/time/v1/models.onk:21`, analysed in
`02-entity-graph.md`). Somebody has already tried to connect Flow reporting to
HR payroll once, declaratively, and it went nowhere.

The lesson applied here: build the consumer and the endpoint together, and do
not add contract surface ahead of a reader. This dossier proposes no new field
on any existing message for exactly that reason.

## How to build it

**Shape 2, a backend aggregate owned by the producer.**

**Why not shape 1, a cross-domain read.** Because shape 1 means Flow reads
rates. That is the salary exposure. Rule 1 says take the loosest shape that
works, and shape 1 does not work: it cannot satisfy the permission requirement.
This is the clearest instance in the audit of the looser shape being genuinely
insufficient rather than merely less tidy.

**Why not shape 5, a contract FK.** There is nothing to link. The join already
exists through `employee_id` and `project_id`.

Sequence: endpoint and Flow panel in one change. Insights consumes the same
endpoint afterwards (`HR->INS-01`).

## Coupling cost

- **New dependency:** flow depends on an HR aggregate. Edge `hr -> flow`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** none. `flow -> hr` candidates were both rejected
  (`FLOW->HR-01`, `FLOW->HR-02`), partly to keep this direction clean.
- **Owner:** **hr** owns `hr.cost.v1` and the definition of labour cost. Flow
  owns the panel.
- **Degrade when HR is disabled:** the Flow project page hides the cost panel
  entirely. It must not render a zero, because zero cost is a meaningful and
  false statement. Entitlement via `packages/config/src/appGuards.tsx`. Flow
  remains fully usable: this is an additive panel.
- **Two-foreign-reads limit:** the Flow project page reads one foreign domain
  (hr), and only through one call. Within budget.

## Risks and correctness notes

- **Permissions, severity critical if done wrong.** A one-contributor project's
  cost divided by its hours is that person's rate. The endpoint returns
  `contributor_count` so the consuming screen can suppress the figure for small
  projects; the screen must actually do so. Gate the endpoint to roles that may
  see cost data, and never return per-person rows.
- **Rate coverage.** Employees with no rate contribute hours and no cost.
  Without `rate_coverage` the total silently understates. This is the failure
  mode the audit exists to prevent, so the field is mandatory, not optional.
- **Historical rates.** No rate history exists. The figure is "valued at current
  rates" and must be labelled as such. If historical accuracy is later required
  it needs a rate-history table, which is a new concept and out of scope here.
- **Rounding and currency.** Return `cost_minor_units: int64` and a currency
  code. Payroll works in integer cents internally
  (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:236`)
  while Finance contracts use `float64`, and there is no single currency
  authority in the product (`02-entity-graph.md`). Rounding once, at the
  display, is the only way this number agrees with itself across screens.
- **Timezone.** Entry `date` is the work date; the period boundary must state
  its timezone, since work days carry their own
  (`packages/voxie/api/flow/time/v1/models.onk:48`).
- **Identity.** Depends on `employee_id` being the right person, which is
  `HR->HR-01`'s problem. If a tenant maintains rates on `internal/modules/hr_staff` instead, this
  endpoint reads a stale rate. **State that dependency in the issue.**
- **Overtime.** `overtime_rate` exists and this formula ignores it. A first
  version costing all hours at base rate understates overtime-heavy projects;
  say so on the screen rather than implying precision.

## Effort

**M.**

| Part                | Estimate                                                                                                        |
| ------------------- | --------------------------------------------------------------------------------------------------------------- |
| Backend             | 4 days: new capability, aggregate query, coverage and contributor counting, role gating, minor-units arithmetic |
| Frontend            | 2 days: Flow project cost panel with partial-coverage and current-rates labelling                               |
| Ongoing maintenance | ~0.5 days per year, higher if overtime and rate history are added later                                         |

## Backend section (round two)

The aggregate this dossier asked for exists as a live platform port, built in
#743 the day after round one closed:

- Contract: `internal/platform/labourcost/contract.go` defines `ProjectCost`
  (cost cents, costed and uncosted minutes, uncosted employee count) and
  `LabourCostProvider`.
- Implementation: HR-owned SQL joins `internal/modules/hr_time_entries` to
  `internal/modules/hr_employees.hourly_rate` inside hr_payroll
  (`internal/modules/hr_payroll/sql/queries/labour_cost.sql`; repository at
  `internal/modules/hr_payroll/internal/infrastructure/repository.go:450`).
- Consumption: workspaces holds it via `SetLabourCostProvider`
  (`internal/modules/workspaces/internal/application/service.go:113,121`) with
  a noop fallback, wired at `cmd/api/main.go:466`.

Not exposed over any RPC; zero frontend callers. Remaining work is one RPC on
Flow's project surface plus the UI, not a new capability.

Deltas versus draft 8a in 06-proposed-contracts.md, which must be revised to
bind this port rather than propose a parallel shape: the real SQL filters
`is_billable = TRUE` only (non-billable minutes are excluded entirely rather
than reported separately); it answers one project per call (no batch ids); it
takes no date range; coverage arrives as `uncosted_minutes` +
`uncosted_employee_count` rather than `rate_coverage`; and there is no
contributor-count equivalent. The permission posture matches the draft's
intent: rates never cross the boundary, only aggregates.

_Round two (2026-08-22): corrected-by-progress; draft 8a superseded in part._
