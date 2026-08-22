# CRM->FIN-01: signed contract from crm to finance

- **Class:** CAN
- **Necessity:** Essential
- **Producer:** crm sales (`crm_contracts`, uncontracted)
- **Consumer:** finance (invoices and quotes)
- **Why (taxonomy):** 2 double entry elimination, 4 trigger and automation
- **Shape:** 1, a cross-domain read from the Finance app
- **Severity:** high

## Benefit

The day after this ships, a won deal can be turned into a draft invoice with its
plan, value and customer already filled in, instead of being retyped.

## Why this pairing

**Double entry elimination.** CRM records what was sold and for how much.
Finance bills it. The handover is a human retyping.

`crm_contracts` holds `lead_id`, `plan`, `mrr`, `seats` and `signed_at`
(`~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:62`).
A signed contract with a monthly value is exactly the input an invoice needs.

**The complication: CRM sales is uncontracted.** The `crm` contract domain has
capabilities for campaign, customer, dashboard, feedback, loyalty, segment,
settings and template, and **no sales capability at all.** The CRM app reaches
leads, contacts, touchpoints and contracts through legacy paths such as
`/crm/sales/contracts` (`apps/crm/src/shared/hooks/queries/useSales.ts:52`
onward), served by the `crm_sales` module against `crm_leads`, `crm_contacts`,
`crm_touchpoints` and `crm_contracts`.

So this connection cannot be built the right way without contracting CRM sales
first. That is the main cost, and it is why this proposal ranks below
`CRM->FIN-02` despite both touching the same pair.

**Why a draft and not an automatic invoice.** Issuing an invoice is a financial
act with tax, currency, payment terms and timing attached, none of which is on
the contract record. An automatic invoice would either guess those or produce a
document nobody reviewed. A prefilled draft keeps the accountant in the loop.

**What the user does instead right now.** Reads the deal in CRM and retypes it
into the invoice form, including the client details that
`CRM->FIN-02` addresses separately.

## ROI

| Field                  | Value                                                                                                                                                                              |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants using CRM sales with Finance.                                                                                                                                              |
| Manual work today      | retype plan, value and customer per won deal into an invoice                                                                                                                       |
| Modelled monthly value | reference tenant, assume 8 new contracts signed per month. Retyping a deal into an invoice costs perhaps 4 minutes: **32 minutes per month**. Modest.                              |
|                        | The error side matters more: a mistyped `mrr` invoices the wrong amount, and nothing reconciles the invoice back to the deal, so the discrepancy is never detected.                |
| Build cost             | backend **3 days**, most of it contracting a minimal `crm.sales.v1` surface rather than the read itself; frontend 1.5 days (prefill from contract); maintenance ~0.5 days per year |
| Payback                | ~6 months on time alone. Better if the recurring-invoice case is included, since `finance.recurring_invoice` already exists and an MRR contract is naturally a recurring invoice.  |
| Confidence             | **Inferred from code** for the contract table, its columns and the absent CRM sales contract domain. **Modelled from assumption** for deal volume and the 4 minutes.               |
| Cost of not doing it   | n/a, CAN not MUST. Deals and invoices stay unreconciled, so "did we bill everything we sold" is unanswerable.                                                                      |

## The data

| Field           | From                           | Cited                                                                                 |
| --------------- | ------------------------------ | ------------------------------------------------------------------------------------- |
| `plan`          | crm_contracts                  | `~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:67` |
| `mrr`           | crm_contracts                  | `~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:68` |
| `seats`         | crm_contracts                  | `~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:69` |
| `signed_at`     | crm_contracts                  | `~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:70` |
| customer        | via `crm_contacts.customer_id` | `~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:41` |
| invoice created | finance invoices               | `packages/voxie/api/finance/invoices/v1/service.onk:7`                                |

Note the customer path: a contract references a `lead_id`, and a **contact**
references both a lead and a `customer_id`. So contract to customer is a
two-hop join through contacts, and it is only as reliable as whether a contact
was created and linked. That is worth stating plainly rather than implying the
customer is directly available.

## What exists today

- `crm_contracts`, `crm_leads`, `crm_contacts`, `crm_touchpoints` in the
  backend, reached by the CRM app through legacy paths
  (`apps/crm/src/shared/hooks/queries/useSales.ts:52`).
- `InvoicesService.createInvoice`
  (`packages/voxie/api/finance/invoices/v1/service.onk:7`) and a
  `finance.recurring_invoice` capability.
- **No CRM sales contract domain.**
- **Manual workaround:** retyping.

## What is missing

| Wanted                              | Contract reality                        | Backend reality                         | What is missing                                                                      |
| ----------------------------------- | --------------------------------------- | --------------------------------------- | ------------------------------------------------------------------------------------ |
| Read a signed contract from Finance | **no `crm.sales.v1` capability exists** | `crm_sales` module serves legacy routes | a contracted read of a signed contract                                               |
| Know which customer it is for       | `crm_contacts.customer_id` exists       | two-hop join via lead                   | either the join server-side or an explicit `customer_id` on the contract             |
| Not re-invoice the same deal        | nothing                                 | nothing                                 | a record on the contract that an invoice was raised, or idempotency at the call site |

**Bucket: needs a contract field**, in the strong sense that a whole capability
is missing.

## Precedent in the codebase

The precedent to avoid: Flow calls `/hr/staff/detail`, an ad hoc REST endpoint
outside the contract, which no onion module serves
(the profile call site (rewritten upstream), and Flow's own schema file
says so at `the deleted ad hoc schema file (removed with the dead call, #875 fix)`). A cross-domain read to
an uncontracted path is how that happened. **Finance must not read
`/crm/sales/contracts` directly**, however tempting, because a legacy path has
no owner, no type safety and no guarantee it survives.

## How to build it

**Shape 1, a cross-domain read**, after contracting the producer.

1. Contract a minimal `crm.sales.v1` covering the signed contract, including a
   resolved `customer_id` so the consumer does not reimplement the two-hop join.
2. Finance's invoice create screen accepts a contract id, reads it, prefills
   amount and description.
3. Record that an invoice was raised, so the action is idempotent.

**Why not shape 3, an event on deal-won.** Rejected for the same reason as
`REC->HR-01`: an automatic financial document nobody reviewed. Also, "won" is a
`stage` string on the lead with no fixed vocabulary
(`~/Development/onion/internal/platform/database/schemas/crm/v2_sales_pipeline.sql:12`
defaults it to `'cold'`), so the product cannot reliably detect the trigger
without imposing a stage name on every tenant, which the configurability rules
forbid.

**Why not shape 2.** No aggregation needed.

## Coupling cost

- **New dependency:** finance depends on a new CRM sales contract. Edge
  `crm -> finance`, the same edge as `CRM->FIN-02`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `crm <-> finance` is bidirectional; see `12-coupling.md`.
- **Owner:** the new **crm** sales contract is owned by the CRM team. Finance
  owns the prefill.
- **Degrade when CRM is disabled:** the "from a deal" entry point is absent and
  the invoice form works exactly as today. Additive, never required.

## Risks and correctness notes

- **Permissions.** Deal values and pipeline data would become readable by
  Finance users. Defensible, since they are billing it, but the read should
  return **one signed contract by id**, not the pipeline. A list endpoint would
  hand Finance the whole sales funnel.
- **Tenancy.** Standard scoping; the contract id must be validated against the
  caller's business.
- **Currency.** `mrr` is `DECIMAL(10,2)` with **no currency column** on
  `crm_contracts`. Invoices carry their own `currency`
  (`packages/voxie/api/finance/invoices/v1/models.onk:40`). Prefilling an amount
  across that boundary assumes a currency the contract never stated. The
  invoice's currency must be chosen by the user or defaulted from Finance's own
  settings, and **not** silently inherited.
- **Rounding.** `DECIMAL(10,2)` to `float64` on the invoice side. Fine at these
  magnitudes; worth noting the representation change.
- **Idempotency.** Without the raised-invoice record, clicking twice bills
  twice. That is a money bug, not a UX wrinkle.
- **Timezone.** `signed_at` is a `DATE`; invoice dates are separate. Do not
  default the invoice date to the signature date without saying so.
- **Historical rates:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                                    |
| ------------------- | ----------------------------------------------------------------------------------------------------------- |
| Backend             | 3 days: minimal `crm.sales.v1` contract with a resolved customer, the invoice-raised marker, tenant scoping |
| Frontend            | 1.5 days: prefill path on invoice create, converted-state on the CRM contract row                           |
| Ongoing maintenance | ~0.5 days per year, plus keeping the new contract in step with the legacy sales routes until those retire   |

## Backend section (round two)

Verified against voxie `847a34c8` and onion `ab6ff7ff`. Unchanged: voxie's api
tree carries no `crm/sales/` directory; the crm_sales module serves 17 legacy
chi routes (`backend_handlers.csv`, crm_sales); the CRM app hand-writes its
lead/contact/contract payload types at
`apps/crm/src/shared/hooks/queries/useSales.ts:106,245,272,347`.

_Round two (2026-08-22): confirmed unchanged._
