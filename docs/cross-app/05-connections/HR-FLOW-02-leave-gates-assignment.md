# HR->FLOW-02: approved leave from hr to flow

- **Class:** CAN
- **Necessity:** High value
- **Producer:** hr.leave.v1 (`LeaveRequestPayload`)
- **Consumer:** flow (assignment and cycle planning)
- **Why (taxonomy):** 6 gating and correctness, 5 context in place
- **Shape:** 1, a cross-domain read from the Flow app
- **Severity:** medium

## Benefit

The day after this ships, assigning someone work due while they are on approved
leave shows a warning, and cycle planning shows who is actually available.

## Why this pairing

**Gating and correctness.** HR approves leave. Flow assigns work. Neither knows
about the other, so a manager can commit a person's time to a week they are
already away for, and only find out when the work is late.

The producer data is in good shape, in the newer of HR's two leave systems:
`internal/modules/hr_leave_requests` carries `employee_id`, `start_date`, `end_date` and `status`
(`~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:24`),
and `LeaveService.listLeaveRequests` serves it
(`packages/voxie/api/hr/leave/v1/service.onk:6`). Leave types are configurable
rows rather than an enum, which matters for the multi-tenant rule: a law firm
and a restaurant name their leave differently.

**Critically, this system keys on `employee_id`**, the same id Flow work days
use (`packages/voxie/api/flow/time/v1/models.onk:38`). The older
`internal/modules/hr_time_off_requests` keys on `staff_id` and hardcodes its types
(`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:56`).
**Use `hr.leave.v1`.** Choosing the other one would require crossing HR's
nullable person bridge (`HR->HR-01`).

**Why a warning and not a block.** Assigning work to someone on leave is
sometimes correct: leave gets cancelled, deadlines are negotiable, and the
person may be the only one who can do it. A hard block would be the product
asserting a business rule that varies by tenant, which the configurability
rules forbid. A warning informs without deciding.

**What the user does instead right now.** Checks the HR leave calendar in
another tab, or does not.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running Flow with HR.                                                                                                                                                                                                                                                                     |
| Manual work today      | opening HR to check availability before assigning, or discovering the clash later                                                                                                                                                                                                                 |
| Modelled monthly value | **Errors avoided.** Reference tenant: 40 people, ~25 days of leave taken per month in aggregate across the team. Assume assignments clash with approved leave a handful of times a month, say 4, and each costs an hour of rescheduling and a slipped commitment: **~4 hours per month** avoided. |
|                        | Cycle planning improves separately: knowing true availability before committing a cycle is a decision-quality gain, not a time saving.                                                                                                                                                            |
| Build cost             | backend **1 day** (date-range filter and pagination on `ListLeaveRequestsRequest`, `06-proposed-contracts.md` item 2), frontend 2 days (warning on assign, availability in cycle planning), maintenance ~0.5 days per year                                                                        |
| Payback                | ~2 months on the modelled clash figure.                                                                                                                                                                                                                                                           |
| Confidence             | **Inferred from code** for the leave schema, the id alignment and the missing filter. **Modelled from assumption** for the 4 clashes and the hour each.                                                                                                                                           |
| Cost of not doing it   | n/a, CAN not MUST. Work continues to be assigned into holidays.                                                                                                                                                                                                                                   |

## The data

| Field                    | From             | Cited                                                                           |
| ------------------------ | ---------------- | ------------------------------------------------------------------------------- |
| `employee_id`            | hr leave request | `~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:24` |
| `start_date`, `end_date` | hr leave request | `~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:26` |
| `status`                 | hr leave request | `~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:30` |
| assignee, due date       | flow task        | `packages/voxie/api/flow/task/v1/models.onk:23`                                 |

Rule, not formula:

```
warn if a task's due date falls within [start_date, end_date]
     of any leave request for that assignee with status = 'approved'
```

Only `approved`. Warning on pending leave would train users to ignore the
warning.

## What exists today

- `hr.leave.v1` with 8 RPCs including balances and approval levels
  (`packages/voxie/api/hr/leave/v1/service.onk:6`).
- Flow tasks carry an assignee (`packages/voxie/api/flow/task/v1/models.onk:23`).
- Flow already reads the HR domain 11 times (`01-inventory.md`), so the wire is
  established.
- **Manual workaround:** checking HR separately.

## What is missing

| Wanted                        | Contract reality                                                                                                       | Backend reality     | What is missing                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------- | ------------------------------------------------------------------- |
| Who is away between two dates | `ListLeaveRequestsRequest` filters by `employee_id` and `status` only (`packages/voxie/api/hr/leave/v1/models.onk:87`) | rows have the dates | a `from_date`/`to_date` filter with stated overlap semantics        |
| A bounded read                | response reports `limit`/`offset`                                                                                      | ignored             | the request cannot page: one of the 13 HR endpoints in `04-gaps.md` |

**Bucket: needs a contract field**, specifically query filters.

## Precedent in the codebase

Flow reading HR is established practice, and the audit found the existing
instances (11 requests). One of them is a warning: Flow calls
`GET /hr/staff/detail` (the profile call site (rewritten upstream)) and
**no onion module registers that route**, so it is a cross-domain read to a
non-existent endpoint. Flow's own schema file comments that the endpoint is ad
hoc and outside `hr.staff.v1`
(`the removed schema file`). This proposal must go through the
contracted `hr.leave.v1`, not repeat that pattern.

## How to build it

**Shape 1, a cross-domain read.** Flow reads approved leave for the relevant
people over the relevant window.

**Why shape 1 is sufficient.** Leave dates are not sensitive the way rates are:
a colleague's absence is generally visible in any workplace calendar. One
foreign domain, within the read budget. No aggregation needed.

**Why not shape 2.** Only if leave dates turn out to need permission filtering
per tenant, which is a policy question worth asking but not assuming. Note the
risk below about leave _type_.

**Why not shape 3, an event.** Nothing is triggered; a screen is informed.

## Coupling cost

- **New dependency:** flow depends on hr leave. Edge `hr -> flow`, the same
  edge `HR->FLOW-01` uses, so no new node pair is introduced.
- **Direction:** consumer depends on producer. Correct. No cycle.
- **Owner:** **hr** owns leave. Flow owns the warning.
- **Degrade when HR is disabled:** no warning appears and cycle planning shows
  no availability column. Assignment works exactly as today. The warning is
  additive and must never block a save when the data is unavailable, otherwise
  an HR outage stops work being assigned.
- **Two-foreign-reads limit:** if a Flow project page shows both the cost panel
  (`HR->FLOW-01`) and availability, both read hr, which counts as one foreign
  domain. Still within budget.

## Risks and correctness notes

- **Permissions, and this is the real one: leave type leaks health data.** Leave
  types are configurable and commonly include sick leave. Exposing the _type_
  to any Flow member discloses that a colleague was off sick, which in several
  jurisdictions is special-category personal data. **Flow must read dates and
  approval status only, never the leave type or the reason.** `notes` and
  `rejection_reason` exist on the request
  (`~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:29`,
  `:32`) and must not cross. If the contract cannot express a projection that
  omits type, that is the trigger to switch to shape 2.
- **Tenancy.** Server-scoped, but the client must not assume an unpaginated list
  is complete: with no `limit`, behaviour on a large team is undefined.
- **Identity.** Correct only because `hr.leave.v1` uses `employee_id`. If a
  tenant records leave in the older `staff_id` system, this warning silently
  never fires. That is a real partial-coverage risk and should be stated in the
  issue, with `HR->HR-01` as the fix.
- **Timezone.** Leave is stored as `DATE`, tasks have due dates. Comparing them
  across timezones can be off by a day at the boundary. Compare calendar dates
  in the tenant's timezone, not instants.
- **Currency, rounding, historical rates:** not applicable.

## Effort

**S/M.**

| Part                | Estimate                                                                                                                         |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 1 day: date-range filter with overlap semantics, pagination, and a projection that omits leave type if the type must be withheld |
| Frontend            | 2 days: assignment warning, availability in cycle planning                                                                       |
| Ongoing maintenance | ~0.5 days per year                                                                                                               |

## Backend section (round two)

Confirmed unchanged: `ListLeaveRequestsRequest` filters by employee id and
status only; no date range, no pagination fields
(`packages/voxie/api/hr/leave/v1/models.onk`).

_Round two (2026-08-22): confirmed unchanged._
