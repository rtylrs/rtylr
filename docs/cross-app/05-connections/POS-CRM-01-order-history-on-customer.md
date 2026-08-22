# POS->CRM-01: order history from pos to crm

- **Class:** IS
- **Necessity:** High value
- **Producer:** pos.orders.v1 (`Order`)
- **Consumer:** crm (customer page)
- **Why (taxonomy):** 5 context in place, 1 derivation
- **Shape:** 2, a backend aggregate owned by POS
- **Severity:** medium

## Benefit

The day after this ships, a CRM customer page shows what that customer has
actually bought and when they last came in, so loyalty and campaign decisions are
made against behaviour rather than guesswork.

## Why this pairing

**Context in place, on the product's one genuinely working cross-domain link.**

POS orders already reference CRM customers for real: `pos.Order.customer_id`
(`packages/voxie/api/pos/orders/v1/models.onk:35`), and the POS customers module
reads `crm_customers` directly with a comment saying so
(`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:30`).
The POS contract even documents the resulting name collision as a hazard
(`packages/voxie/api/pos/customer/v1/models.onk:3`). This is classed **IS**
because the identity is already shared; what is missing is a way to ask the
question.

CRM also already has fields that look like they were meant for this:
`total_spent`, `visit_count` and `last_visit_at`
(`packages/voxie/api/crm/customer/v1/models.onk:17`, `:18`, `:19`). **Inferred,
not verified:** something may populate these from POS today, or they may be
unset; the audit found no write path from orders to the customer record, so any
values present come from elsewhere. Either way they are summary fields with no
drill-down.

**The blocker is the query, not the link.** `ListOrdersRequest` requires
`session_id` as `@required @gt(0)`
(`packages/voxie/api/pos/orders/v1/models.onk:145`). A session is a till session.
So orders can only be listed **within one till session**, and "this customer's
orders across all time" is not a question the contract can express, despite the
customer reference existing on every order.

**Why POS owns the aggregate.** A customer's order history spans sessions,
locations and terminals. Summing it is POS's business, and returning a customer
summary avoids handing CRM the raw order stream including line items, staff ids
and table numbers.

**What the user does instead right now.** Nothing, or asks someone with POS
access.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                            |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running CRM with POS, so retail and hospitality. Not the services reference tenant.                                                                                                                                                                      |
| Manual work today      | ask a manager to look up a customer in POS                                                                                                                                                                                                                       |
| Modelled monthly value | modelled hospitality tenant: loyalty and campaign decisions currently made without purchase history. A modest **1 hour per month** of lookups, plus better-targeted campaigns, which is a decision-quality gain rather than a time saving and is not quantified. |
| Build cost             | backend **2 days** (a customer-scoped order summary aggregate), frontend 1 day (panel on the CRM customer page), maintenance ~0.5 days per year                                                                                                                  |
| Payback                | ~4 months on lookups alone; better if campaign targeting improves at all.                                                                                                                                                                                        |
| Confidence             | **Inferred from code** that the customer reference is real, that POS reads `crm_customers`, and that `ListOrdersRequest` is session-scoped. **Modelled from assumption** for the hour.                                                                           |
| Cost of not doing it   | n/a, IS/CAN not MUST. Though note that CRM shows `total_spent` and `visit_count` with no way to see what they are made of, which is a summary a user cannot verify.                                                                                              |

## The data

| Field                                         | From         | Cited                                                            |
| --------------------------------------------- | ------------ | ---------------------------------------------------------------- |
| `customer_id`                                 | pos order    | `packages/voxie/api/pos/orders/v1/models.onk:35`                 |
| order totals and dates                        | pos order    | `packages/voxie/api/pos/orders/v1/models.onk:30` onward          |
| `total_spent`, `visit_count`, `last_visit_at` | crm customer | `packages/voxie/api/crm/customer/v1/models.onk:17`, `:18`, `:19` |

```
history(customer) = orders where customer_id = C,
                    summarised as count, total, last order date
```

**If CRM's `total_spent` is populated from somewhere, this aggregate will
disagree with it at some point.** That is a single-source-of-truth question the
implementation must answer explicitly: either the aggregate becomes the authority
and `total_spent` is derived from it, or the panel shows both and labels which is
which. Shipping two numbers for "what has this customer spent" without deciding
would add to the problem this audit is documenting.

## What exists today

- `pos.Order.customer_id`, a real join.
- POS customers reading `crm_customers` cross-schema.
- `ListOrdersRequest` with `session_id` required, plus `status`, `date_from`,
  `date_to`, `search`, `page`, `page_size`
  (`packages/voxie/api/pos/orders/v1/models.onk:144` onward). **Everything
  except a customer filter, and the session requirement blocks it anyway.**
- CRM summary fields.
- `pos.LoyaltyTransaction.customer_id` and `pos.GiftCard.customer_id`
  (`packages/voxie/api/pos/loyalty/v1/models.onk:24`,
  `packages/voxie/api/pos/gift_card/v1/models.onk:11`), so three POS entities
  reference CRM customers.
- **Manual workaround:** asking.

## What is missing

| Wanted                              | Contract reality                                                                              | Backend reality                                        | What is missing                              |
| ----------------------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------ | -------------------------------------------- |
| A customer's orders across sessions | `ListOrdersRequest` requires `session_id` (`packages/voxie/api/pos/orders/v1/models.onk:145`) | orders carry `customer_id`                             | a customer-scoped query or summary aggregate |
| One authority for lifetime spend    | CRM has `total_spent`; POS has the orders                                                     | no write path found from orders to the customer record | a decision about which is authoritative      |

**Bucket: needs a query or endpoint.**

## Precedent in the codebase

This is the best precedent in the audit rather than needing one: POS referencing
and reading CRM customers is the product's one fully working cross-domain
identity (`02-entity-graph.md`). It is the model `CRM->FIN-02` and
`SUP->CRM-01` are asked to imitate. That also raises the odds this connection
gets consumed once built, since the link is already load-bearing rather than
declarative.

## How to build it

**Shape 2, a backend aggregate owned by POS.**

**Why not shape 1, a cross-domain read.** Two reasons, and the first is
sufficient: the contract cannot express the query, because orders are
session-scoped. Even with a customer filter added, the raw order stream carries
line items, staff ids, table numbers and location ids, most of which CRM has no
business receiving.

**Why not extend `ListOrdersRequest` with a customer filter instead.** That is
the looser change and it was considered. It fails on the second reason above:
CRM would then be reading full order records. A summary is both narrower and
cheaper to render. If a tenant later wants order-level drill-down in CRM, that is
the point to revisit.

**Why not shape 3.** Nothing is triggered.

## Coupling cost

- **New dependency:** crm depends on a POS aggregate. Edge `pos -> crm`.
- **Direction:** consumer depends on producer. Correct. No cycle: `crm -> pos`
  is `no-consumer-surface` in the matrix, since POS has no operator app here.
- **Owner:** **pos** owns orders and the summary definition. CRM owns the panel.
- **Degrade when POS is not licensed:** the panel hides itself. For any tenant
  not using POS, which includes the reference tenant, that is the normal state.
- **Two-foreign-reads limit, and this is where it binds.** The CRM customer page
  is the most contested screen in the audit: `FIN->CRM-01` (invoices),
  `SUP->CRM-01` (tickets) and this one (orders) all want a panel there. **Three
  foreign domains breaches rule 5.** `08-roadmap.md` resolves it: at most two
  ship as direct reads, and if a third is wanted the page needs a single
  CRM-owned customer-summary aggregate that fans out server-side. That is the
  concrete trigger condition, and it is why the roadmap does not put all three
  in the same wave.

## Risks and correctness notes

- **Permissions.** Purchase history is customer behaviour data. CRM users
  generally should see it. **Line items should not cross**: what someone bought
  in detail is more sensitive than how much and how often, particularly in
  pharmacy or similar retail. The aggregate returns counts and totals.
- **The competing-total risk.** Covered above. Decide, do not ship two.
- **Currency.** POS orders may be in one currency per location; a customer
  visiting two locations in different currencies has no single lifetime total.
  Per-currency subtotals, never a cross-currency sum.
- **Tenancy.** Orders and customers are both tenant-scoped; the aggregate must
  scope by `business_id`, and the customer id passed in must be validated
  against the caller's tenant.
- **Rounding.** Sum in minor units where the contract allows; `float64` totals
  otherwise, rounded once.
- **Timezone.** "Last visit" and any period grouping in the tenant's timezone,
  not the browser's; a hospitality tenant's trading day may not end at midnight.
- **Historical rates:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                  |
| ------------------- | ----------------------------------------------------------------------------------------- |
| Backend             | 2 days: customer-scoped order summary aggregate, per-currency handling, tenant validation |
| Frontend            | 1 day: panel on the CRM customer page, hidden when POS is absent                          |
| Ongoing maintenance | ~0.5 days per year                                                                        |

## Backend section (round two)

Confirmed: `ListOrdersRequest.session_id` remains `@required @gt(0)`
(`packages/voxie/api/pos/orders/v1/models.onk:145`); a customer-scoped order
query still cannot be expressed.

_Round two (2026-08-22): confirmed unchanged._
