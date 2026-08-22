# REC->HR-01: hired candidate from recruit to hr

- **Class:** CAN
- **Necessity:** Essential
- **Producer:** recruit (no contract domain; `careers_submissions`)
- **Consumer:** hr (employee records)
- **Why (taxonomy):** 2 double entry elimination, 4 trigger and automation,
  8 identity resolution
- **Shape:** 1, a cross-domain read from the HR app
- **Severity:** high

## Benefit

The day after this ships, marking a candidate hired lets an HR admin create
their employee record from the application they already submitted, instead of
retyping the whole person.

## Why this pairing

**Double entry elimination, at its most literal.** The candidate typed their
own name, email and phone into the public application form. The business then
types the same three things again into HR.

The producer side already has a `hired` state and nothing behind it. The schema
comment spells out the funnel:

```
-- Status funnel: new → reviewed → interviewing → offered → hired
-- (or → rejected at any stage). Kept as a VARCHAR so HR can rename
-- stages later without a migration.
status VARCHAR(32) NOT NULL DEFAULT 'new',
```

`~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:90`

The table holds `candidate_name`, `email`, `phone`, `resume_url` and
`cover_letter`
(`~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:85`).
It holds **no `employee_id`**, so nothing records that this candidate became
that employee, and `hired` is a label with no consequence.

**Why HR pulls rather than recruit pushing.** Hiring is a decision with
paperwork attached: contract type, department, location, start date, salary.
None of that is in the application. An automatic push would create
half-populated employee records that an HR admin then has to complete or
delete. A pull, where HR opens a prefilled create-employee form, keeps the
human decision and removes only the typing. This is also why the shape is a
read and not an event, unlike `HR->HUB-01`.

**What the user does instead right now.** Reads the candidate's details off the
recruit admin screen and retypes them into HR's create-employee form. For the
reference tenant that is the whole person, once per hire.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                           |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants using Recruit with HR. Recruit is a small app (23 requests, all uncontracted) so this is a narrower audience than most proposals.                                                                                                                       |
| Manual work today      | retype name, email, phone; separately store or lose the resume link                                                                                                                                                                                             |
| Modelled monthly value | reference tenant at 20% annual turnover hires ~8 people a year, so **0.67 hires per month**. Retyping and cross-checking a person costs perhaps 6 minutes: **4 minutes per month**. Negligible as time.                                                         |
|                        | The value is error avoidance and provenance: a mistyped employee email breaks their invite and their payslip delivery, and there is currently no record linking an employee back to the application they were hired from, which matters for any hiring audit.   |
| Build cost             | backend 1.5 days (a contracted read of one submission, plus the `employee_id` link), frontend 1 day (prefill the create-employee form), maintenance ~0.25 days per year                                                                                         |
| Payback                | not justified on time saved for a 40-person tenant. Justified for tenants that hire at volume, and as the removal of a status value that currently means nothing. **This is the weakest ROI among the Essential items**, and the roadmap places it accordingly. |
| Confidence             | **Inferred from code** for the schema, the funnel and the absent link. **Modelled from assumption** for hiring volume and the 6 minutes.                                                                                                                        |
| Cost of not doing it   | n/a, CAN not MUST. A `hired` status that produces nothing is a small lie the product tells about itself.                                                                                                                                                        |

## The data

| Field                              | From                | Cited                                                                                   |
| ---------------------------------- | ------------------- | --------------------------------------------------------------------------------------- |
| `candidate_name`, `email`, `phone` | careers_submissions | `~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:85` |
| `resume_url`                       | careers_submissions | `~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:88` |
| `status`                           | careers_submissions | `~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:93` |
| employee created                   | hr employee         | `packages/voxie/api/hr/employee/v1/service.onk:9` (`createEmployee`)                    |

No formula. Note the name shape mismatch: recruit stores one
`candidate_name VARCHAR(160)`, HR stores separate first and last names. Splitting
on the first space is wrong for a great many real names, so the prefill should
put the whole string in the first-name field and let the admin correct it, or
present two fields pre-split with the split clearly editable. **Silently
guessing a surname is the kind of small wrongness that is worse than typing.**

## What exists today

- `careers_submissions` with the funnel and the candidate's own details.
- `EmployeeService.createEmployee` (`packages/voxie/api/hr/employee/v1/service.onk:9`)
  and `inviteEmployee` (`:15`).
- The recruit app's admin submissions screen, reached through uncontracted
  `/recruit/submissions` paths (23 requests, `01-inventory.md`).
- **Manual workaround:** retyping.

## What is missing

| Wanted                                     | Contract reality                          | Backend reality                                                                                                                              | What is missing                                                                                             |
| ------------------------------------------ | ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| Read one submission from HR                | **recruit has no contract domain at all** | the `careers` module serves plain route strings and imports no generated contract (`~/Development/onion/internal/modules/careers/module.go`) | a contracted read. This is the only proposal in the audit whose producer has no contract surface whatsoever |
| Record that a candidate became an employee | no field                                  | no column                                                                                                                                    | an `employee_id` on the submission, or the reverse                                                          |

**Bucket: needs a query or endpoint**, with the unusual complication that the
producing domain needs a contract before it can have an endpoint worth
consuming. **Do not consume the legacy `/recruit/*` paths from the HR app.**
Cross-domain reads should go through contracts, or the coupling has no owner
and no type safety, which is the whole argument of `07-plumbing.md`.

## Precedent in the codebase

There is no good precedent: no app currently reads recruit. The relevant
precedent is negative and worth stating. Recruit is the only app whose entire
API surface is uncontracted (`01-inventory.md`), which means every cross-app
connection into or out of it starts by paying a contract-creation cost that
other pairs do not have. That cost is the main reason this proposal ranks below
its apparent simplicity.

## How to build it

**Shape 1, a cross-domain read**, once a contract exists.

1. Contract a minimal `recruit.submission.v1` with a `getSubmission` RPC
   returning the candidate fields. Minimal on purpose: HR needs the person, not
   the pipeline.
2. HR's create-employee screen accepts a submission id, reads it, prefills.
3. Add the `employee_id` link on the submission so the provenance is recorded
   and the button becomes idempotent: a submission already converted shows the
   employee instead of offering to create a second one.

**Why not shape 3, an event on hire.** Rejected above: an automatic push
creates incomplete employee records. The human decision is the feature.

**Why not shape 2, an aggregate.** Nothing to aggregate, one record.

**Why not shape 5, a contract FK as the primary mechanism.** The `employee_id`
link _is_ a small FK, but it records the outcome rather than carrying the data,
so the connection's shape is the read.

## Coupling cost

- **New dependency:** hr depends on a recruit contract. Edge `recruit -> hr`.
  This edge has no matrix cell because recruit is not one of the 12 domains
  (`generated-matrix.md`, "connections outside the domain matrix").
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** none. Nothing is proposed from hr into recruit.
- **Owner:** the new **recruit** contract is owned by whoever owns the careers
  module. HR owns the prefill.
- **Degrade when Recruit is disabled:** the HR create-employee form shows no
  "from an application" option at all. It is an additive entry point, not a
  replacement, so the manual form remains the primary path. Entitlement via
  `packages/config/src/appGuards.tsx`.

## Risks and correctness notes

- **Permissions, and this one is sharp.** Candidate applications contain
  personal data from people who are not employees and never will be: names,
  phone numbers, cover letters, CVs. Exposing a submission-read to the HR app
  widens who can see applicant PII. It should be gated to the same role that
  can already see submissions in the recruit admin, and it should return **one
  submission by id**, never a list. A list endpoint here would let an HR user
  browse every applicant.
- **Data protection.** `resume_url` points at stored files. Copying it onto an
  employee record extends the retention of an applicant document into employee
  records, which may cut against a tenant's own retention policy. Recommend
  linking, not copying, and stating it in the UI.
- **Tenancy.** `careers_submissions` carries both `owner_id` and `business_id`
  (`~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:83`),
  and the comment notes `owner_id` is denormalised for tenant-scoped queries.
  The read must scope on `business_id`.
- **Idempotency.** Without the `employee_id` link, clicking twice creates two
  employees. The link is not optional polish.
- **Name splitting.** Covered above. Do not guess.
- **Currency, rounding, timezone, historical rates:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                                                                          |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 1.5 days: a minimal recruit contract with one RPC, the `employee_id` column and link, tenant scoping, role gating                                 |
| Frontend            | 1 day: prefill path on HR create-employee, converted-state handling on the recruit submission row                                                 |
| Ongoing maintenance | ~0.25 days per year, plus the standing cost of a new contract capability that must be kept in step with the legacy routes until those are retired |

## Backend section (round two)

Confirmed: careers serves 23 legacy chi routes with no contract domain
(`backend_handlers.csv`, careers rows); `EmployeeService.createEmployee` exists
as cited. Round one's numbers survive precise re-measurement.

_Round two (2026-08-22): confirmed unchanged._
