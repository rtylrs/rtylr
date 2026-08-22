# HR->INS-01: labour cost from hr to insights

- **Class:** CAN
- **Necessity:** High value
- **Producer:** hr (the same `hr.cost.v1` aggregate as `HR->FLOW-01`)
- **Consumer:** insights
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 2, a backend aggregate owned by HR
- **Severity:** medium

## Benefit

The day after this ships, Insights can put a cost side against the revenue it
already reports, so margin becomes a metric the analytics app can show rather
than a calculation someone does elsewhere.

## Why this pairing

**Rollup.** Insights reports revenue, customers, products and operations. Labour
is the largest cost in most service businesses and Insights cannot see it, so
every margin-shaped question is unanswerable in the app whose job is answering
them.

**This is deliberately the same endpoint as `HR->FLOW-01`, not a second one.**
Both want "what did this labour cost". Two endpoints would drift and eventually
give two answers to one question, which is the exact failure the audit is
documenting elsewhere (two stock numbers, two pay rates, two leave systems).
`06-proposed-contracts.md` item 8a is written to serve both: it takes project
ids and a date range, and `HR->INS-01` uses it with a period and a wider set of
projects.

**Why Insights consumes rather than computes.** Insights owns no employment data
and should not: rates and hours are HR's. Insights owns the presentation and the
metric definitions built on top.

**Why this ranks below `HR->FLOW-01` despite being the same endpoint.** Flow's
project page is where the person who can act on the number is standing. An
Insights chart is where the number is reviewed. Build the actionable surface
first; the reporting surface then costs almost nothing.

**What the user does instead right now.** Exports payroll and joins it to
whatever revenue picture they have, outside RTYLR.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running Insights with HR.                                                                                                                                                                                                                    |
| Manual work today      | build the cost side of any margin figure by hand                                                                                                                                                                                                     |
| Modelled monthly value | reference tenant: a monthly margin or utilisation review assembled by hand costs perhaps **2 hours**. Once `HR->FLOW-01` exists, the marginal work here is small, so most of the value is in the first proposal and this one is the cheap follow-on. |
| Build cost             | backend **0.5 days** if `HR->FLOW-01` has shipped (the endpoint exists; this adds period-level querying), otherwise 4 days. Frontend 1.5 days. Maintenance shared with `HR->FLOW-01`.                                                                |
| Payback                | very fast **as a follow-on**. Poor as a standalone, which is why the roadmap sequences it after.                                                                                                                                                     |
| Confidence             | **Inferred from code** that Insights has no labour-cost message and HR has the inputs. **Modelled from assumption** for the 2 hours.                                                                                                                 |
| Cost of not doing it   | n/a, CAN not MUST.                                                                                                                                                                                                                                   |

## The data

Identical inputs to `HR->FLOW-01`: `project_id`, `duration_minutes` and
`employee_id` from `hr.time_entry.v1`
(`packages/voxie/api/hr/time_entry/v1/models.onk:12`, `:20`, `:10`), priced with
`internal/modules/hr_employees.hourly_rate`
(`~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:91`).

The difference is the shape of the question: a period total across projects
rather than one project's cost. The proposed response already carries
`rate_coverage` and `contributor_count`, both of which matter more in a reporting
context because a chart hides the caveats a panel can state.

## What exists today

- Insights capabilities for customer, dashboard, financial, operation and
  product (`packages/voxie/api/insights/dashboard/v1/service.onk:6` onward).
- `insights.StaffUtilization` carries a `staff_id`
  (`packages/voxie/api/insights/operation/v1/models.onk:22`), so Insights already
  models a per-person operational metric, **keyed on `staff_id`** rather than
  `employee_id`. That is HR's two-person-record problem surfacing in the
  reporting domain (`HR->HR-01`).
- No labour cost anywhere in the insights contracts.
- Insights reads **no** foreign business domain today (`01-inventory.md`).
- **Manual workaround:** external spreadsheets.

## What is missing

| Wanted                   | Contract reality                                           | Backend reality    | What is missing                                       |
| ------------------------ | ---------------------------------------------------------- | ------------------ | ----------------------------------------------------- |
| Labour cost per period   | no cost message in insights                                | inputs exist in hr | the `hr.cost.v1` aggregate, shared with `HR->FLOW-01` |
| Consistent person keying | `StaffUtilization.staff_id` vs time entries' `employee_id` | two person tables  | `HR->HR-01`, or an explicit mapping in the aggregate  |

**Bucket: needs a query or endpoint**, satisfied by the same endpoint as
`HR->FLOW-01`.

## Precedent in the codebase

`ERP->INS-01` is the sibling proposal and the two should be built to the same
pattern: Insights consuming a source domain's aggregate rather than fetching raw
records. Together they establish how Insights reaches outside itself, and doing
them inconsistently would leave two patterns for the same job.

## How to build it

**Shape 2, the shared HR aggregate.**

**Why not shape 1.** Same reason as `HR->FLOW-01`: reading employee records to
price hours client-side would expose pay rates. Non-negotiable.

**Why not shape 4, a shared hook in `packages/`.** Two apps needing the same
data is exactly the case rule 6 says is **not** enough: the bar is three or more
apps plus a stable shape. Flow and Insights make two. Duplicating a small query
hook in each app is cheaper than a `packages/**` change that redeploys all ten
Cloudflare Pages apps.

Sequence: after `HR->FLOW-01`.

## Coupling cost

- **New dependency:** insights depends on the HR cost aggregate. Edge
  `hr -> insights`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** none. `insights -> hr` is `insights-is-terminal` in the matrix.
- **Owner:** **hr** owns the cost definition. **insights** owns the metrics
  built from it.
- **Degrade when HR is disabled:** cost and margin metrics hide themselves;
  revenue metrics remain. Charts must not plot zero cost, since a zero-cost
  margin chart reads as 100% margin, which is the most misleading possible
  degradation. **Hide the series, do not zero it.**

## Risks and correctness notes

- **Permissions.** Same as `HR->FLOW-01`: aggregate only, role-gated, with
  `contributor_count` available so a small-group total can be suppressed. An
  Insights chart is more likely than a Flow panel to be shared as a screenshot,
  which makes the small-group inference risk slightly higher here.
- **Rate coverage in a chart.** A panel can carry a sentence saying the figure
  is partial; a chart cannot. Either annotate the axis or refuse to plot periods
  whose coverage is below a threshold. Silently plotting an understated cost
  series is the failure mode.
- **Person keying.** `StaffUtilization` uses `staff_id` while cost is keyed on
  `employee_id`. Any screen putting utilisation and cost side by side is
  crossing HR's nullable person bridge. Say so, or fix `HR->HR-01` first.
- **Currency and rounding.** Minor units from the endpoint, rounded once at
  display; no currency authority exists so the currency must be shown.
- **Historical rates.** Current rates only, labelled. This matters more in a
  trend chart than in a single figure: a rate change makes historical periods
  move, which looks like a data bug.
- **Timezone.** Period boundaries in the tenant's timezone.
- **Tenancy:** standard.

## Effort

**S** as a follow-on, **M** standalone.

| Part                | Estimate                                                                                |
| ------------------- | --------------------------------------------------------------------------------------- |
| Backend             | 0.5 days after `HR->FLOW-01`; 4 days standalone                                         |
| Frontend            | 1.5 days: cost and margin series with coverage annotation and hide-not-zero degradation |
| Ongoing maintenance | shared with `HR->FLOW-01`                                                               |

## Backend section (round two)

Confirmed: insights declares no labour-cost message. With the labourcost port
now live (see HR->FLOW-01's backend section), the same provider can serve both
consumers once exposed, preserving the single-number design.

_Round two (2026-08-22): confirmed; serving mechanism identified._
