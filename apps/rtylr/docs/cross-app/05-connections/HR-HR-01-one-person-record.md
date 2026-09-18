# HR->HR-01: one person record, one pay rate, inside hr

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** hr.staff.v1 (`Staff`) and hr.employee.v1 (`EmployeePayload`)
- **Consumer:** hr, and every proposal that needs to identify a person
- **Why (taxonomy):** 3 single source of truth, 8 identity resolution
- **Shape:** 6, a resolver, scoped to one domain
- **Severity:** high

**This is not a crossing.** It is an intra-domain finding, and it is in the
dossier set because four proposed crossings are degraded or blocked by it.
`generated-matrix.md` lists it under intra-domain findings rather than in a
cell.

## Benefit

The day after this ships, "what do we pay this person" and "is this person on
leave" have one answer each, because there is one person record instead of two
that can disagree.

## Why this pairing

**Single source of truth, violated inside a single domain.** HR stores a person
twice, in two tables, each with its own pay fields.

|                 | `internal/modules/hr_staff`                                                                                   | `internal/modules/hr_employees`                                                      |
| --------------- | ------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Declared as     | `hr.Staff` (`packages/voxie/api/hr/staff/v1/models.onk:4`)                                                    | `hr.EmployeePayload` (`packages/voxie/api/hr/employee/v1/models.onk:4`)              |
| Schema          | `~/Development/onion/internal/platform/database/schemas/hr/schema.sql:15`                                     | `~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql`          |
| `user_id`       | yes                                                                                                           | yes                                                                                  |
| `salary`        | yes (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:29`)                               | yes                                                                                  |
| `hourly_rate`   | yes (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:30`)                               | yes (`~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:91`) |
| `overtime_rate` | yes (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:31`)                               | yes (`~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:92`) |
| Bridge          | `employee_id BIGINT DEFAULT NULL` (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:20`) |                                                                                      |

The bridge is **nullable**. Two records for one person, two pay rates, and a
link that may not be set.

This is not a dormant duplication. Different HR features key on different ids:

- **Time off** keys on `staff_id` (`packages/voxie/api/hr/time_off/v1/models.onk:7`;
  column at `~/Development/onion/internal/platform/database/schemas/hr/schema.sql:55`).
- **Payroll** keys on `employee_id`
  (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:198`).
- **Flow work days** key on `employee_id`
  (`packages/voxie/api/flow/time/v1/models.onk:38`).

And there are **two leave systems**, not one:

|             | `internal/modules/hr_time_off_requests`                                                                                                          | `internal/modules/hr_leave_requests`                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------- |
| Keys on     | `staff_id`                                                                                                                                       | `employee_id`                                                                   |
| Leave types | **hardcoded ENUM** `('vacation','sick','personal','unpaid','other')` (`~/Development/onion/internal/platform/database/schemas/hr/schema.sql:56`) | configurable rows, `internal/modules/hr_leave_types`                            |
| Contract    | `hr.time_off.v1` (`TimeOffService`, 2 RPCs)                                                                                                      | `hr.leave.v1` (`LeaveService`, 8 RPCs, with balances and approval levels)       |
| Schema      | `~/Development/onion/internal/platform/database/schemas/hr/schema.sql:51`                                                                        | `~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:20` |

The older one hardcodes leave types, which directly violates this repo's own
configurability rule that a leave type name must never be hardcoded (root
`CLAUDE.md`, section 8, which uses leave types as its worked example). The
newer one gets it right.

**Why this is in the dossier set.** `HR->FLOW-02` (leave gates assignment)
has to decide which leave table is authoritative. `HR->FLOW-01` and
`HR->INS-01` (labour cost) have to decide which `hourly_rate` is authoritative.
`HR->HUB-01` (revoke on termination) has to watch both `status` columns or be
bypassable. Four proposals, one prerequisite.

**What the user does instead right now.** Unknown, and that is the point:
whichever screen a tenant uses determines which record they maintain, and the
other silently rots. **Inferred, not verified** that real tenants have divergent
records; confirming it needs production data, which this audit does not touch.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                              |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | every HR tenant.                                                                                                                                                                                                                                                                                                                                   |
| Manual work today      | none visible, which is what makes it dangerous: the duplication is invisible until two screens disagree.                                                                                                                                                                                                                                           |
| Modelled monthly value | **Errors avoided, not time saved.** If a tenant maintains `internal/modules/hr_staff.hourly_rate` while payroll reads `internal/modules/hr_employees.hourly_rate`, payroll pays the wrong amount and nothing flags it. For the reference tenant, 40 people paid monthly, a single stale rate is one wrong payslip per month until someone notices. |
|                        | Value is therefore "correct payroll", which does not reduce to minutes. The countable part is the reconciliation nobody currently does.                                                                                                                                                                                                            |
| Build cost             | backend **6 to 10 days**: decide the canonical record, migrate, redirect the older leave system or retire it, keep both contracts serving during the transition. Frontend 2 days to follow the surviving ids. Maintenance goes **down** afterwards.                                                                                                |
| Payback                | not expressible as a payback period. This is debt removal that unblocks four proposals worth more than it costs.                                                                                                                                                                                                                                   |
| Confidence             | **Inferred from code** for every structural claim: both schemas, both contracts, both leave tables and the nullable bridge were read directly. **Modelled from assumption** that divergence occurs in practice.                                                                                                                                    |
| Cost of not doing it   | Two answers to what someone is paid, with payroll reading one of them. Two answers to whether someone is on leave. Any leave-gating or labour-costing feature built on top inherits the ambiguity and produces confidently wrong numbers, which is worse than producing none. Four proposals stay degraded.                                        |

## The data

Nothing moves between apps. What changes is which row is authoritative.

| Decision           | Options                                                                                                                                      |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Canonical person   | `internal/modules/hr_employees` is the stronger candidate: payroll, leave balances, Flow work days and time entries all key on `employee_id` |
| Canonical pay rate | whichever record survives; payroll already reads `internal/modules/hr_employees.hourly_rate`                                                 |
| Canonical leave    | `internal/modules/hr_leave_requests`, because it keys on `employee_id`, has configurable types, balances and approval levels                 |

**This dossier does not make those decisions.** They are data-ownership calls
for the HR owners, with a migration behind each. The audit's contribution is
establishing that the decisions are outstanding and that four other proposals
are waiting on them.

## What exists today

- Both tables, both contracts, both leave systems, as tabulated above.
- One working resolver, hub user to hr employee
  (`~/Development/onion/internal/modules/hr_core/sql/queries/employees.sql:47`).
- The `staff` module reads both `internal/modules/hr_staff` and `internal/modules/hr_employees`, so the backend
  already has the join in one place.
- `hr.time_off.v1` documents its own vestigial state: a message says "Not
  currently wired to any RPC in service.proto - kept for parity with the source
  proto" (`packages/voxie/api/hr/time_off/v1/models.onk:25`).

## What is missing

| Wanted            | Contract reality                   | Backend reality                                      | What is missing                  |
| ----------------- | ---------------------------------- | ---------------------------------------------------- | -------------------------------- |
| One person record | two messages, both with pay fields | two tables, nullable bridge                          | a decision and a migration       |
| One leave system  | two contracts, 2 RPCs vs 8         | two tables, two key spaces, one with hardcoded types | retire or redirect the older one |
| One pay rate      | duplicated in both                 | duplicated in both                                   | collapse to the canonical record |

**Bucket: needs a new concept**, in the sense that "the canonical person" is
not currently a concept the schema expresses.

## Precedent in the codebase

The `hr.time_off.v1` doc comment quoted above is the precedent worth noting:
the team already recognises parts of that contract are vestigial and said so in
the contract rather than deleting it. That is the honest instinct this proposal
extends: finish the sentence.

The counter-precedent is `internal/modules/hr_category_id` (`02-entity-graph.md`): a link left
unconsumed. `internal/modules/hr_staff.employee_id` is the same species, a nullable bridge that
nothing enforces.

## How to build it

**Shape 6, a resolver, deliberately scoped to one domain.**

`07-plumbing.md` says shape 6 should be proposed once, for the identity
problem, not per connection. This is the intra-HR instance of it and it is
separated from `IDENTITY-01` for a practical reason: it can be done first, by
one team, without designing a product-wide party concept. It is the pilot.

Sequence:

1. **Report before migrating.** Count people with rows in both tables, and how
   many have divergent `hourly_rate`, `salary` or `status`. That number decides
   how careful the migration has to be, and it is the only measurement anyone
   will get of the current damage.
2. Pick the canonical record. Keep both contracts serving, with the
   non-canonical one reading through to the canonical row.
3. Migrate leave: `internal/modules/hr_time_off_requests` rows mapped onto `internal/modules/hr_leave_requests`
   with their enum values becoming configured leave types.
4. Only then retire the duplicate columns.

**Why not a looser shape.** There is no read, aggregate or event that makes two
authoritative pay rates into one. Every looser shape leaves the ambiguity in
place and hides it better.

## Coupling cost

- **New dependency:** none. This **removes** coupling: four proposals currently
  depend on understanding both tables and would afterwards depend on one.
- **Direction and cycles:** not applicable, intra-domain.
- **Owner:** **hr**, unambiguously.
- **Degrade when a producing app is disabled:** not applicable, no crossing.
  The relevant guarantee is different: during the transition **both contracts
  must keep working**, because apps in this monorepo deploy independently and a
  Cloudflare Pages app pinned to an older bundle will keep calling the older
  endpoints after the migration ships.

## Risks and correctness notes

- **Migration risk is the whole risk, and it is money.** Choosing the wrong
  canonical `hourly_rate` changes what people get paid. The report in step 1 is
  not optional, and divergent rows should be resolved by a human rather than by
  a rule like "most recently updated wins".
- **Permissions.** Both records carry salary. Collapsing them must not widen
  who can read pay: the surviving contract should keep pay fields behind the
  same role gate the current employee record uses. Note that `hr.Staff` exposes
  `salary` and `hourly_rate` in a message some screens may fetch broadly
  (`packages/voxie/api/hr/staff/v1/models.onk:17`); the migration is an
  opportunity to narrow that, and a risk of widening it by accident.
- **Tenancy.** Both tables carry `business_id`. Matching people across the two
  must be scoped by tenant, and matching on email alone across tenants would be
  a cross-tenant leak.
- **Historical rate changes.** Payroll entries already written reference rates
  as they were. The migration must not retro-price historical payroll: existing
  `internal/modules/hr_payroll_entries` are immutable records.
- **Rounding and currency.** Payroll works in integer cents internally
  (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:236`)
  while the contracts expose `float64`. A migration that round-trips rates
  through `float64` can move a value by a cent. Migrate in minor units.
- **Timezone:** not applicable.

## Effort

**L.**

| Part                | Estimate                                                                                                                                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 6 to 10 days: divergence report, canonical decision, dual-serving period, leave migration including enum-to-rows, retirement of duplicate columns |
| Frontend            | 2 days: follow surviving ids, remove references to the retired leave surface                                                                      |
| Ongoing maintenance | **negative.** This is the only item in the audit that reduces long-run maintenance rather than adding to it.                                      |

## Backend section (round two)

Confirmed unchanged at the storage layer: `internal/platform/database/schemas/hr/schema.sql:29-30` and
`internal/platform/database/schemas/hr_core/schema.sql:90-91` both carry salary and hourly_rate (plus overtime),
bridged by nullable `internal/modules/hr_staff.employee_id`. One scope-widening fact: team's
invite flow INSERTs into `internal/modules/hr_employees` cross-schema
(`internal/modules/team/sql/queries/team.sql:194`), so hub-side seat
provisioning co-owns the employee table. The migration plan must treat team as
a stakeholder module, not just the HR modules.

_Round two (2026-08-22): confirmed; scope widened to include team provisioning._
