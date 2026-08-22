# SUP->CRM-01: ticket history from support to crm

- **Class:** IS
- **Necessity:** High value
- **Producer:** support.customer.v1 (`Customer`)
- **Consumer:** crm (customer page)
- **Why (taxonomy):** 5 context in place
- **Shape:** 1, a cross-domain read from the CRM app
- **Severity:** medium

## Benefit

The day after this ships, a CRM customer page shows whether that customer is
currently having a bad time with support, before anyone calls them to upsell.

## Why this pairing

**Context in place.** Support knows whether a customer is unhappy. CRM decides
what to say to them. The classic failure is an upsell call to someone with three
open tickets.

This is classed **IS** rather than CAN because the wire already exists: the CRM
app makes 2 support-domain requests today (`01-inventory.md`). What is missing is
the customer-level join, not the connection.

The producer side is unusually convenient. `support.Customer` already carries
`total_tickets` and `last_ticket_at` alongside `name`, `email` and `phone`
(`packages/voxie/api/support/customer/v1/models.onk:3`), and
`CustomerService.getCustomer` serves it
(`packages/voxie/api/support/customer/v1/service.onk:6`). Somebody already
aggregated the ticket count onto the customer record.

**The problem is which customer.** `support_customers` is a **separate table**
from `crm_customers`. Nothing joins them. `support.inbox` compounds it by storing
`customer_name` and `customer_email` as strings
(`packages/voxie/api/support/inbox/v1/models.onk:5`) rather than referencing even
the support customer.

So the only available join today is **email**, which is right until a customer
emails from a second address, at which point the CRM page silently shows zero
tickets for someone with five.

**Why CRM and not a combined view.** CRM is where the customer relationship is
managed. Support has no app in this repo at all, so there is no support surface
to put CRM data on.

**What the user does instead right now.** Nothing, or asks. There is no support
app to check.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                            |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants using CRM whose support inbox is in RTYLR. Because there is **no support app in this repo**, that population is currently small and may be zero: the support backend exists with no operator surface (`01-inventory.md`). This materially weakens the case and is stated here rather than buried.                                                        |
| Manual work today      | none possible: there is no support screen to check                                                                                                                                                                                                                                                                                                               |
| Modelled monthly value | conditional on support being used at all. If it is, avoiding a handful of tone-deaf sales contacts a month is worth more in relationship terms than in minutes, and cannot be honestly quantified from code. **No number is offered**, because inventing one for a feature whose user base may be zero would be exactly the fabrication the brief warns against. |
| Build cost             | backend **1 day** for the email-match version, **2 days** with a proper `customer_id` on the support customer (`06-proposed-contracts.md` item 7); frontend 1 day (panel); maintenance ~0.25 days per year                                                                                                                                                       |
| Payback                | unquantified, deliberately. Justified as cheap and as a stepping stone that exercises the identity link on a low-risk surface.                                                                                                                                                                                                                                   |
| Confidence             | **Inferred from code** for the fields, the separate tables and the CRM app's existing support reads. **Unvalidated** business value: the connection is technically possible and its demand is unverified.                                                                                                                                                        |
| Cost of not doing it   | n/a, IS/CAN not MUST.                                                                                                                                                                                                                                                                                                                                            |

## The data

| Field                             | From             | Cited                                                                                                    |
| --------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------- |
| `total_tickets`                   | support customer | `packages/voxie/api/support/customer/v1/models.onk:3`                                                    |
| `last_ticket_at`                  | support customer | same message                                                                                             |
| `email` (the only join key today) | both sides       | `packages/voxie/api/support/customer/v1/models.onk:3`, `packages/voxie/api/crm/customer/v1/models.onk:9` |
| `customer_id` (proposed)          | support customer | `06-proposed-contracts.md` item 7                                                                        |

No formula. The join is the substance, and the join is currently a string
comparison.

## What exists today

- `support.Customer` with counts, and `getCustomer` / `listCustomers`
  (`packages/voxie/api/support/customer/v1/service.onk:4`, `:6`).
- 5 `support_*` backend modules.
- 2 existing support requests from the CRM app.
- `support_customers` as its own table, unjoined to `crm_customers`.
- **No support app.**

## What is missing

| Wanted                        | Contract reality                       | Backend reality     | What is missing                             |
| ----------------------------- | -------------------------------------- | ------------------- | ------------------------------------------- |
| Tickets for this CRM customer | no reference between the two customers | separate tables     | a `customer_id` on the support customer     |
| A reliable join               | email on both sides                    | email on both sides | the above, because email is not an identity |

**Bucket: needs a contract field.**

## Precedent in the codebase

`pos.Order.customer_id` referencing `crm_customers` is the working precedent
(`packages/voxie/api/pos/orders/v1/models.onk:35`), and POS goes further by
reading `crm_customers` directly
(`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:33`).
Support could have done the same and did not. There is no evidence of a reason;
it looks like the two were simply built at different times.

## How to build it

**Shape 1, a cross-domain read**, with the `customer_id` field from item 7 so
the read has something reliable to key on.

**Why not ship the email-match version alone.** It is one day cheaper and it
produces a number that is silently wrong for any customer with more than one
address. A "0 open tickets" badge on an unhappy customer is worse than no badge.
If the email version does ship first, **the panel must say it matches on email**
so a reader knows what they are looking at.

**Why not wait for `IDENTITY-01`.** This is a good pilot for the identity work:
low stakes, small surface, and it exercises the nullable-reference-with-honest-
degradation pattern that the party concept will generalise.

## Coupling cost

- **New dependency:** crm depends on support customers. Edge `support -> crm`.
- **Direction:** consumer depends on producer. Correct. No cycle: nothing is
  proposed from crm into support.
- **Owner:** **support** owns tickets and the customer reference. CRM owns the
  panel.
- **Degrade when support is unlicensed or unused:** the panel hides itself. Given
  there is no support app, **the default state for most tenants is hidden**, and
  the panel must be built for that case first rather than treating it as the edge
  case.
- **Two-foreign-reads limit:** with `FIN->CRM-01`, the CRM customer page reaches
  two foreign domains. That is the cap. See that dossier.

## Risks and correctness notes

- **Permissions.** Ticket counts are mild, but support conversations are not.
  Read **counts and dates only**, never ticket bodies or subjects, into CRM. A
  support thread can contain anything a customer chose to write.
- **The email join is the main correctness risk.** Covered above. Do not present
  a matched-on-email figure as authoritative.
- **Tenancy.** Both tables carry `business_id`. An email match must be scoped by
  tenant or it becomes a cross-tenant leak, which for an email-keyed join is a
  genuinely plausible mistake.
- **Currency, rounding, historical rates:** not applicable.
- **Timezone.** `last_ticket_at` displayed as a relative time should use the
  tenant's timezone.

## Effort

**S.**

| Part                | Estimate                                                                     |
| ------------------- | ---------------------------------------------------------------------------- |
| Backend             | 1 to 2 days depending on whether the `customer_id` reference is included     |
| Frontend            | 1 day: panel with counts, last-contact date, and hidden-by-default behaviour |
| Ongoing maintenance | ~0.25 days per year                                                          |

## Backend section (round two)

Confirmed: support_customers remain their own table; no statement in the
support_customers query set references CRM tables.

_Round two (2026-08-22): confirmed unchanged._
