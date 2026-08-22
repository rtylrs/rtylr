# HR->HUB-01: employment status change from hr to hub

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** hr.employee.v1 (`EmployeePayload.status`)
- **Consumer:** hub (app entitlement and session), and through it every app
- **Why (taxonomy):** 4 trigger and automation, 6 gating and correctness
- **Shape:** 3, an event emitted by HR and consumed by hub
- **Severity: critical.** This is the only finding in the audit that is a
  security exposure rather than a missing feature.

## Benefit

The day after this ships, terminating an employee in HR actually ends their
access to RTYLR, instead of leaving a working login behind until somebody
remembers to remove it by hand in a different app.

## Why this pairing

**Gating and correctness.** An action in one app (terminate) has a consequence
in another (access) that is currently not applied at all.

Terminating an employee runs one statement. The handler sets a single field and
returns:

```
emp := domain.EmployeeRecord{ID: id, Status: "terminated"}
_, err := h.Service.UpdateEmployeeForWriter(ctx, businessID, device.User.UserID, emp)
```

`~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:426`

Nothing else happens. The employee's `internal/modules/hr_employees.user_id`
(`packages/voxie/api/hr/employee/v1/models.onk:10`) still points at a live hub
user. Their app entitlements, which hub models with `is_enabled`
(`packages/voxie/api/hub/app/v1/models.onk:8`), are untouched. Their Flow
workspace membership (`packages/voxie/api/flow/workspace/v1/models.onk:37`)
is untouched.

**Why hub and not each app.** The tempting design is for HR to tell Flow, then
POS, then Finance. That is one edge per surface, each of which can be forgotten
exactly the way the manual process is forgotten today. hub already owns the
entitlement every app reads through `AppGuard`
(`packages/auth/src/guards/AppGuard.tsx`), so revoking there revokes
everywhere, including apps built later. This is also why `HR->POS-01` was
rejected and folded into this dossier rather than proposed separately
(`11-rejected.md`).

**What the user does instead right now.** An administrator opens dash, goes to
the team screen, and removes the member. The audit found the endpoint they use
(`DELETE /hub/team/members/v1/members/{user_id}`,
`packages/voxie/api/hub/team/members/v1/service.onk:12`). Nothing prompts
them, nothing reminds them, and nothing reports which terminated employees
still have access.

## ROI

Reference tenant as defined in `08-roadmap.md`.

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                            |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | any tenant using HR plus at least one other app. For the reference tenant, all of them.                                                                                                                                                                                                                                                                          |
| Manual work today      | remove the team member in dash separately from terminating in HR, with no prompt or cross-check                                                                                                                                                                                                                                                                  |
| Modelled monthly value | a 40-person services business at a mid-range 20% annual turnover has 8 departures a year, so 0.67 per month. The manual revocation takes about 3 minutes when remembered. Time saved is negligible: **0.67 x 3 = 2 minutes per month.** The value is not the time.                                                                                               |
|                        | The value is the exposure removed. Assume, as a modelled figure, that 1 in 5 departures has revocation delayed or missed. That is **1.6 lingering accounts per year**, each with access to payroll, customer records and finance for an unbounded period.                                                                                                        |
| Build cost             | backend 3 days (emit on status change, consume in hub, revoke entitlements and invalidate sessions), frontend 0.5 days (surface the result and a "still has access" warning list), maintenance ~0.25 days per year                                                                                                                                               |
| Payback                | not meaningful in time saved. Justified as risk removal, and it is the cheapest item on the critical list.                                                                                                                                                                                                                                                       |
| Confidence             | **Inferred from code** for the mechanism: the termination handler and its single statement were read directly. **Modelled from assumption** for turnover and the 1-in-5 miss rate; neither is measured.                                                                                                                                                          |
| Cost of not doing it   | An ex-employee retains a working login with their previous permissions. For a tenant using HR and Finance that is standing access to payroll and banking data by someone the business believes it has removed. It is also the kind of finding that fails a customer's own security review, which for a B2B product is a sales problem as well as a security one. |

## The data

One fact moves: an employee's employment status became terminal.

| Field                              | Source          | Cited                                                                                            |
| ---------------------------------- | --------------- | ------------------------------------------------------------------------------------------------ |
| `status` (to `terminated`)         | hr employee     | `packages/voxie/api/hr/employee/v1/models.onk:19`                                                |
| `user_id` (whose access to revoke) | hr employee     | `packages/voxie/api/hr/employee/v1/models.onk:10`                                                |
| `business_id` (tenant scope)       | request context | `~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:415` |

No formula. The consequence is a state change, not a calculation.

Note the identity dependency: this works because `internal/modules/hr_employees.user_id` is one
of the four real joins in the product (`02-entity-graph.md`), and there is
already a resolver query in the opposite direction
(`~/Development/onion/internal/modules/hr_core/sql/queries/employees.sql:47`).
**An employee whose `user_id` is unset has no login to revoke**, which is
correct behaviour, and the warning list should say so rather than reporting
success.

## What exists today

- `hr.EmployeePayload.status` already carries `terminated` as a documented
  value (`packages/voxie/api/hr/employee/v1/models.onk:19`).
- `internal/modules/hr_staff.status` separately carries `terminated` in an ENUM
  (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:28`),
  which is the second person record problem (`HR->HR-01`). **Both must be
  handled or the revocation is bypassable through whichever record the tenant
  actually maintains.**
- hub models per-app entitlement with `is_enabled`
  (`packages/voxie/api/hub/app/v1/models.onk:8`).
- A webhooks module exists in the backend
  (`~/Development/onion/internal/modules/webhooks`), and a notifications
  module (`~/Development/onion/internal/modules/notifications`), so event
  plumbing does not have to be invented.
- Manual workaround: removing the team member in dash.

## What is missing

| Wanted                                    | Contract reality                                            | Backend reality                                                                                                                                        | What is missing                                                                                                                          |
| ----------------------------------------- | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Terminating an employee ends their access | `status` includes `terminated`; hub `App.is_enabled` exists | the termination handler sets one column and returns (`~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:426`) | the consequence. An in-process event from hr_core on status transition to `terminated`, consumed by the hub entitlement and session code |
| Seeing who slipped through                | nothing                                                     | nothing                                                                                                                                                | a reconciliation read: employees with terminal status whose `user_id` still has enabled entitlements                                     |

**Bucket: needs a query or endpoint.** No contract field is missing. Nothing in
`apps/**` can fix it, because the fix has to happen where the status changes.

## Precedent in the codebase

The nearest precedent is a warning, not a model. `internal/modules/hr_category_id` shows what
happens when a cross-domain link is declared and no consumer is written
(`02-entity-graph.md`): it sits unread. This proposal is the inverse shape, a
consequence with no declaration, and it should be built consumer-first: write
the hub revocation and the reconciliation list before or with the event, so
there is never a period where an event fires into nothing.

The genuine precedent for the mechanism is hub's own entitlement enforcement,
which every app already honours through `AppGuard`
(`packages/auth/src/guards/AppGuard.tsx`). This proposal adds a writer to a
system that already has readers everywhere, which is the cheapest possible
position to be in.

## How to build it

**Shape 3, an event.** HR emits `employee.employment_ended` carrying
`business_id`, `employee_id` and `user_id`. hub consumes it and disables that
user's app entitlements and invalidates sessions.

**Why not shape 1, a cross-domain read.** A read cannot work: the consumer
would have to poll for terminations, and the whole problem is that the
consequence must be immediate and unconditional. There is no screen involved.

**Why not shape 2, a consumer-owned aggregate.** Same reason. Nobody is asking
a question; a state change needs propagating.

**Why an event and not a direct call from hr_core into hub.** Both modules live
in the same Go binary, so a direct call is genuinely simpler and should be
considered first. The argument for the event is rule 8, a named owner: HR owns
"employment ended", hub owns "what that means for access", and POS or a future
app will want the same signal. A direct call makes HR responsible for knowing
every consumer, which is the coupling this audit is trying to avoid. **If the
team prefers a direct call for the first version, that is a defensible
narrowing and the reconciliation list matters more than the mechanism.**

Order of work, and it matters:

1. The reconciliation read first: which terminated employees still have access.
   It is independently valuable, it quantifies the existing exposure, and it is
   the test oracle for the rest.
2. The hub-side revocation, callable and tested on its own.
3. The emission from `internal/modules/hr_core`, covering both `internal/modules/hr_employees` and `internal/modules/hr_staff`.

## Coupling cost

- **New dependency:** hub depends on an HR event. One edge, `hr -> hub`.
- **Direction:** consumer (hub) depends on producer (hr). Correct per rule 2.
  No cycle: hub has no proposed outbound edge to hr (`HUB->HR-01` was rejected,
  `11-rejected.md`), so `hr -> hub` is the only edge between them.
- **Owner:** the event contract is owned by **hr**. The revocation semantics
  are owned by **hub**.
- **Degrade when the producing app is disabled:** if a tenant has no HR licence
  there are no HR employees, so no events, and hub behaves exactly as it does
  today. Nothing in hub becomes conditional on HR existing. The reconciliation
  list, which lives in dash, shows an empty state when HR is absent rather than
  an error. This is the easiest degrade story in the audit because the event is
  additive: its absence is the current behaviour.

## Risks and correctness notes

- **Tenancy.** The event must carry `business_id` and hub must scope revocation
  by it. A user who belongs to two businesses (hub supports switching,
  `packages/voxie/api/hub/user/v1` `switchBusiness`) must lose access to the
  terminating tenant only. **Getting this wrong locks a consultant out of
  every client.** This is the single highest-risk detail in the proposal.
- **Permissions.** Emission is a side effect of an action already gated by HR's
  own permission check. The reconciliation read exposes "who is terminated",
  which is HR data, so it must be gated to administrators, not shown to every
  dash user. No role gains sight of data it cannot see today: the list contains
  facts an HR admin already sees.
- **The second person record.** `internal/modules/hr_staff` and `internal/modules/hr_employees` both carry
  status. If only one is wired, a tenant maintaining the other gets no
  revocation and a false sense of safety. This is a hard dependency on
  understanding `HR->HR-01`, though not on completing it.
- **Reversal.** Terminations get reversed, and rehiring is normal. Revocation
  must be re-grantable, and the event must not be treated as a delete.
- **Timezone and rounding:** not applicable, no dates or money.
- **Historical rate changes:** not applicable.

## Effort

**S/M.**

| Part                | Estimate                                                                                                                                                  |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 3 days: reconciliation query, hub revocation, emission from both person tables, multi-business scoping tests                                              |
| Frontend            | 0.5 days: the reconciliation list in dash, plus confirmation copy on termination stating what will be revoked                                             |
| Ongoing maintenance | ~0.25 days per year. Low: one event, one consumer, no schema. Rises if future apps take their own dependency on the event rather than on hub entitlement. |

## Backend section (round two)

Verified against onion at `ab6ff7ff`. The claim stands and one fact strengthens
it: seat provisioning already writes across the boundary in the opposite
direction. Team's invite flow INSERTs an `internal/modules/hr_employees` row cross-schema
(`internal/modules/team/sql/queries/team.sql:192-201`), so employment is coupled
to entitlements at hire time. At termination time the coupling disappears:
`internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:425-427` sets `Status: "terminated"` via
`UpdateEmployeeForWriter`, `internal/modules/staff/sql/queries/staff.sql:26`
terminates the hr_staff row, and no module anywhere reads a terminated status to
revoke `user_app_permissions` or edit any apps array.

_Round two (2026-08-22): backend re-verified, claim confirmed; provisioning-side
cross-schema write added as supporting evidence._
