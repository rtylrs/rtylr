# IDENTITY-01: a shared entity resolver, hub to every domain

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** hub (the natural owner of tenant-wide identity)
- **Consumer:** every domain
- **Why (taxonomy):** 8 identity resolution, 3 single source of truth
- **Shape:** 6, a shared entity resolver
- **Severity:** high

## Benefit

The day after this ships, "this company" and "this person" mean one thing
across RTYLR, so a customer's invoices, orders and support tickets can be
listed together instead of being three unrelated sets of rows that happen to
share a name.

## Why this pairing

**Identity resolution.** The product stores a person or a company up to eight
times. `02-entity-graph.md` has the full pairwise matrix; the summary is that of
28 unordered pairs, **4 are real joins, 1 is an identity, 5 are loose strings
and 18 have no join at all.**

The eight: hub user, hr employee, hr staff, crm customer, pos customer, support
customer, erp supplier, recruit candidate. Plus a ninth thing that is not an
entity but behaves like one: the finance invoice's four `client_*` strings
(`packages/voxie/api/finance/invoices/v1/models.onk:30`).

**Why hub owns it.** hub already owns tenancy and users, is already read by
every app, and is the one domain whose contracts every other domain may depend
on without inverting the dependency direction. Any other owner creates a cycle:
if CRM owned the resolver, hub would depend on CRM for user identity.

**Why this is one proposal and not eighteen.** The brief is explicit that shape
6 should be proposed once, for the identity problem, with the connections it
unblocks listed as its return. Proposing eighteen pairwise links would create
eighteen permanent dependencies to solve one problem, which is precisely the
over-connection the coupling discipline forbids.

**What the user does instead right now.** Exports to a spreadsheet and matches
on email or company name by eye. This is the specific behaviour the audit was
commissioned to find, and identity is its root cause rather than any single
missing join.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                           |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | every tenant running two or more apps that touch the same people or companies. For the reference tenant (HR, Flow, Finance, CRM) that is all four.                                                                                                                                                                                                              |
| Manual work today      | matching customers across CRM, Finance and Support by name or email in a spreadsheet; matching people across HR's two records and hub                                                                                                                                                                                                                           |
| Modelled monthly value | **Decisions enabled, not minutes saved.** The unlockable questions are: revenue net of cost to serve per customer, true customer lifetime value, whether a person is one person. None is answerable today at any price short of manual matching.                                                                                                                |
|                        | A modelled figure for the manual matching alone: an administrator reconciling customers across CRM and Finance monthly for ~120 invoices and a few hundred customers, at perhaps **4 hours per month**. That is the floor, not the value.                                                                                                                       |
| Build cost             | backend **15 to 25 days**, and that is a genuine range not false precision: a party or contact concept, references from the consuming domains, a matching and merge tool, and a migration per domain. Frontend 5 days for the merge and review surfaces. Maintenance ~2 days per year, and it becomes a permanent product concept somebody owns.                |
| Payback                | **not favourable in isolation and should not be sold as such.** It is justified by what it unblocks: `SUP->CRM-01`, `FIN->CRM-01`, `POS->CRM-01` and the customer-360 idea all depend on it to be correct rather than approximate, and `HR->FLOW-02` and the labour-cost pair depend on the HR half.                                                            |
| Confidence             | **Inferred from code** for the entire structural claim: every table, contract and join in `02-entity-graph.md` was read. **Modelled from assumption** for the 4 hours and for any tenant behaviour.                                                                                                                                                             |
| Cost of not doing it   | Every customer-centred and person-centred feature is built on email matching, which is right until it is silently wrong. Worse, the wrongness is invisible: a customer with two email addresses becomes two customers and no screen reports it. Cross-app reporting stays approximate forever, and each new app adds a ninth and tenth copy of the same person. |

## The data

Nothing moves. A reference is introduced.

The concept: a **party**, owned by hub, representing one real-world person or
organisation within one tenant. Each domain's existing record gains a nullable
`party_id`. Records keep their own ids and their own fields; the party is only
the assertion that these rows are the same real thing.

| Domain record     | Gains                                                                                                                        | Keeps                              |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| crm customer      | `party_id`                                                                                                                   | its own id, loyalty, tags, consent |
| pos customer      | already the same rows as crm (`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:33`) | nothing to do                      |
| support customer  | `party_id`                                                                                                                   | ticket counts                      |
| finance invoice   | `customer_id` first (`CRM->FIN-02`), then inherits via crm                                                                   | its `client_*` snapshot            |
| erp supplier      | `party_id`                                                                                                                   | purchasing terms                   |
| hub user          | `party_id`                                                                                                                   | credentials                        |
| hr employee       | `party_id`, once `HR->HR-01` has collapsed the two records                                                                   | employment data                    |
| recruit candidate | `party_id`                                                                                                                   | application data                   |

**A party is not a merge.** Two customer rows pointing at one party remain two
customer rows. That distinction is what makes the migration safe: asserting
sameness is reversible, merging is not.

## What exists today

- Exactly one identity resolver query in the whole backend, hub user to hr
  employee: `SELECT id FROM hr_employees WHERE business_id = ? AND user_id = ?`
  (`~/Development/onion/internal/modules/hr_core/sql/queries/employees.sql:47`).
- One genuine same-rows identity: POS customers read `crm_customers` directly,
  and the POS contract documents the resulting name collision as a hazard
  (`packages/voxie/api/pos/customer/v1/models.onk:3`).
- Three real foreign keys inside the hub/hr cluster
  (`internal/modules/hr_employees.user_id`, `internal/modules/hr_staff.user_id`, `internal/modules/hr_staff.employee_id`).
- **No merge, link, or matching mechanism anywhere.** Searched for identity,
  linking and merge concepts across the backend modules; the only cross-record
  resolution found is the query above.
- **Manual workaround:** spreadsheets.

## What is missing

| Wanted                        | Contract reality           | Backend reality          | What is missing                                                                                     |
| ----------------------------- | -------------------------- | ------------------------ | --------------------------------------------------------------------------------------------------- |
| "This company" across domains | eight independent entities | eight independent tables | a party concept and a reference from each                                                           |
| Knowing two rows are the same | nothing                    | nothing                  | a matching tool with human confirmation, and a place to record the assertion                        |
| Not creating a ninth copy     | nothing prevents it        | nothing prevents it      | a convention, enforced at review, that a new domain references a party rather than storing a person |

**Bucket: needs a new concept.**

## Precedent in the codebase

Two precedents, and they point the same way.

**The good one:** `pos.Order.customer_id` referencing `crm_customers` works,
and POS reading those rows cross-schema works
(`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:30`).
A domain deferring to another domain's record instead of copying it is already
proven here.

**The warning:** `internal/modules/hr_category_id` is a cross-domain reference that was declared
and never consumed (`02-entity-graph.md`). A `party_id` column added to eight
tables with no consumer would be that mistake eight times over. **Hence the
sequencing below: no `party_id` lands before something reads it.**

## How to build it

**Shape 6, and only here.** The tightest and most expensive shape, justified
once, for the problem it was named for.

The sequencing is the whole design, because the failure mode is a
widely-deployed unused column:

1. **`HR->HR-01` first.** Collapse HR's two person records. It is the same
   problem inside one domain, one team can do it, and it is the pilot that
   proves the migration approach.
2. **`CRM->FIN-02` next.** Invoice to customer. It is a single nullable field
   with an immediate consumer (`FIN->CRM-01`) and it demonstrates the value
   before any party concept exists. It is also the item whose cost grows daily.
3. **Then the party concept**, introduced with **one** consuming feature built
   in the same change: the customer view that lists a customer's invoices,
   orders and tickets together. One party reference per domain, added only when
   that feature reads it.
4. **Then the matching tool**, with human confirmation. Never automatic merging
   on email.
5. Remaining domains as their features need them. `erp supplier` and
   `recruit candidate` are last: neither has a waiting consumer.

**Why not shape 5, contract foreign keys pairwise.** That is the eighteen-edge
design. It also does not solve the problem, because a customer known to
Finance, Support and POS needs a common referent, not three bilateral links.

**Why not shape 2, a backend aggregate that joins on email.** Tempting, cheap,
and it produces a number that is right most of the time and silently wrong the
rest. An audit whose purpose is removing confidently-wrong numbers should not
propose one.

## Coupling cost

- **New dependency:** every participating domain depends on hub's party
  contract. That is **many edges into one node**, which is the correct shape for
  identity and the reason it must be hub: hub is already a universal dependency,
  so no new dependency direction is created.
- **Direction:** all inbound to hub as producer. No cycles: hub has one
  proposed outbound edge in the whole audit (`HUB->HR-01`, rejected).
- **Owner:** **hub** owns the party contract and the matching semantics. Each
  domain owns its own reference and whether it is set.
- **Degrade when hub is unavailable:** hub is not optional. No tenant runs
  without it, since `AppGuard` requires it for authentication
  (`packages/auth/src/guards/AppGuard.tsx`). So the usual degrade question does
  not apply. The one that does: **a record with no `party_id` must behave
  exactly as it does today.** Every consuming feature has to handle an unset
  party, because most records will have one unset for a long time. A screen that
  requires a party is a screen that breaks for existing data.

## Risks and correctness notes

- **Tenancy is the top risk.** A party is tenant-scoped. Matching people or
  companies across tenants, even accidentally through a shared email domain,
  would be a cross-tenant data leak of the most serious kind. Every match
  query must be scoped by `business_id` and that scoping should be enforced in
  one place, not per call site.
- **Permissions widen structurally.** Once a party links a CRM customer to a
  Support customer and a Finance invoice, a screen showing "everything about
  this party" can aggregate data from domains the viewer may not be entitled to
  see. **The party must not become a permission bypass.** Each domain's data
  stays behind that domain's own authorisation, and the aggregate view shows
  only the sections the caller may read, degrading section by section. This is
  the single most important correctness note in the audit: get it wrong and the
  identity work becomes a horizontal privilege escalation.
- **Merging is destructive; asserting is not.** Keep them separate. Provide
  unlink. Never auto-merge.
- **False positives.** Two people share an email at small companies; two
  companies share a name. Matching must be advisory with human confirmation.
- **Currency, rounding, timezone, historical rates:** not directly applicable.
  The related concern is that aggregating money across domains crosses a
  currency boundary with no single authority (`02-entity-graph.md`), so any
  party-level financial total must state its currency and refuse to sum across
  differing ones.

## Effort

**L, and the largest item in the audit.**

| Part                | Estimate                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 15 to 25 days across the sequence above, of which the party concept and the first consumer are perhaps 8                  |
| Frontend            | 5 days: the combined customer view, the matching and confirmation surface                                                 |
| Ongoing maintenance | ~2 days per year, plus a standing review obligation: every new domain that stores a person must reference a party instead |

## Backend section (round two)

Confirmed: exactly one identity-resolution query exists product-wide
(`internal/modules/hr_core/sql/queries/employees.sql:47`, hub user to hr
employee). POS-to-CRM contact matching in customer_sync remains contact-based
matching, not identity resolution
(`internal/platform/customer_sync/customer_sync.go:47`), consistent with round one's
reading. Nothing since round one changes the picture.

_Round two (2026-08-22): confirmed unchanged._
