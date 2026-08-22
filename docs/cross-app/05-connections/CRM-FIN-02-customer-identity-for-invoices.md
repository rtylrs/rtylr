# CRM->FIN-02: customer identity from crm to finance

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** crm.customer.v1 (`Customer`)
- **Consumer:** finance (invoices)
- **Why (taxonomy):** 2 double entry elimination, 3 single source of truth,
  8 identity resolution
- **Shape:** 5, a contract-level foreign key
- **Severity:** high

## Benefit

The day after this ships, an invoice knows which customer it was issued to, so
"what has this customer been billed, and what do they still owe" becomes a
question the product can answer instead of a spreadsheet exercise.

## Why this pairing

**Single source of truth.** CRM owns customers. Finance bills them. Today
Finance stores the customer as four unrelated strings and never refers to the
CRM record.

`finance.Invoice` carries `client_name`, `client_email`, `client_address` and
`client_tax_id` (`packages/voxie/api/finance/invoices/v1/models.onk:30`,
`:60`, `:62`, `:64`), the first two `@required` on create
(`packages/voxie/api/finance/invoices/v1/models.onk:147`). There is **no
`customer_id` anywhere in the invoices contract**. Meanwhile `crm.Customer`
holds `id`, `first_name`, `email`, and already tracks `total_spent` and
`loyalty_points` (`packages/voxie/api/crm/customer/v1/models.onk:5`, `:7`,
`:9`, `:17`, `:16`).

The consequences compound:

- Revenue per customer is unanswerable, even though CRM has a `total_spent`
  field that suggests somebody wanted the answer.
- A customer renamed in CRM silently splits: old invoices carry the old
  string, new ones the new, and nothing relates them.
- Two people spelling a company differently create two customers as far as
  Finance is concerned.

**Why this is the identity keystone.** `FIN->CRM-01` (outstanding balance on
the customer page) is impossible without it: there is nothing to filter
invoices by. That is the clearest hard dependency in the roadmap.

**Why not solve it inside `IDENTITY-01`.** The full resolver is the right
long-term answer and it is a wave-4 item. This is the one edge worth doing
early because it is a single nullable field on one message, it unblocks another
proposal, and it does not require the party concept to be designed first.

**What the user does instead right now.** Types the client's name and email
into the invoice form, having already created that customer in CRM.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                         |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | any tenant using Finance with CRM. Also every Finance-only tenant, who gains nothing here and loses nothing: the field is nullable.                                                                                                                                                                                                                           |
| Manual work today      | retype name, email, address and tax id per invoice; reconcile customer revenue by hand                                                                                                                                                                                                                                                                        |
| Modelled monthly value | reference tenant issues ~120 invoices a month. Retyping four fields costs perhaps 45 seconds each: **120 x 0.75 = 90 minutes per month** of administrator time.                                                                                                                                                                                               |
|                        | The larger value is the decision enabled: revenue and outstanding balance per customer, which today requires exporting invoices and matching names by hand. That is a **modelled** claim about how tenants work; the absence of the field is a code fact.                                                                                                     |
| Build cost             | backend 2 days (nullable column, snapshot-on-create, conditional validator, `customer_id` filter on list), frontend 1.5 days (customer picker with free-text fallback), maintenance ~0.5 days per year                                                                                                                                                        |
| Payback                | roughly 2 to 3 months on time alone for the reference tenant, sooner if the per-customer reporting is valued at all.                                                                                                                                                                                                                                          |
| Confidence             | **Inferred from code** that no `customer_id` exists and the four strings are required. **Modelled from assumption** for invoice volume and the 45 seconds.                                                                                                                                                                                                    |
| Cost of not doing it   | Revenue per customer stays unanswerable. Customer records and invoice records drift apart permanently, and the longer this waits the more historical invoices carry strings that can never be confidently matched to a customer. **This is the one item in the audit whose cost grows with delay**, because it is accumulating unmatchable history every day. |

## The data

| Field                 | From            | Cited                                                   |
| --------------------- | --------------- | ------------------------------------------------------- |
| customer `id`         | crm customer    | `packages/voxie/api/crm/customer/v1/models.onk:5`       |
| `first_name`, `email` | crm customer    | `packages/voxie/api/crm/customer/v1/models.onk:7`, `:9` |
| `customer_id` (new)   | finance invoice | proposed, `06-proposed-contracts.md` item 3             |
| `client_*` (kept)     | finance invoice | `packages/voxie/api/finance/invoices/v1/models.onk:30`  |

No formula. The point is the link.

**The snapshot stays.** An invoice is a document that was sent. If a customer
is renamed, last year's invoice must still show the name it was issued under,
so the `client_*` fields are kept and are not re-derived from the customer
record. This is stated in the draft's doc comment because it is exactly the
kind of decision a later engineer would "clean up" by joining and formatting.

## What exists today

- `crm.CustomerService.getCustomer` and `listCustomers` are served
  (`packages/voxie/api/crm/customer/v1/service.onk:6`, `:12`).
- Invoices store the four strings, two required.
- `crm.Customer.total_spent` exists and nothing populates it from invoices
  (**inferred, not verified**: the field is declared at
  `packages/voxie/api/crm/customer/v1/models.onk:17`, and no invoice-to-customer
  path exists for it to be computed from, so any value it holds comes from POS
  orders or is unset).
- **Manual workaround:** retyping.

## What is missing

| Wanted                             | Contract reality                                 | Backend reality                       | What is missing                                                  |
| ---------------------------------- | ------------------------------------------------ | ------------------------------------- | ---------------------------------------------------------------- |
| An invoice that knows its customer | four strings, no id                              | `finance_invoices` stores the strings | a nullable `customer_id` on `Invoice` and `CreateInvoiceRequest` |
| Invoices for one customer          | no filter possible                               | n/a                                   | `customer_id` on `ListInvoicesRequest`                           |
| Issue an invoice without retyping  | `client_name` and `client_email` are `@required` | validator enforces both               | a conditional rule: a `customer_id` **or** the strings           |

**Bucket: needs a contract field.** Draft in `06-proposed-contracts.md` item 3.

## Precedent in the codebase

The good precedent is `pos.Order.customer_id`
(`packages/voxie/api/pos/orders/v1/models.onk:35`), which is a real join to
`crm_customers` and works: POS reads those rows directly
(`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:33`).
So "a domain outside CRM referencing a CRM customer by id" is already an
established, functioning pattern in this product. This proposal extends a
pattern rather than inventing one, which materially raises the odds it gets
consumed, unlike `internal/modules/hr_category_id`.

The bad precedent is `finance.Bill.supplier_id`, a required FK to a domain
Finance never reads, which the UI satisfies by fabrication
(`apps/finance/src/features/bills/Bills.tsx:59`). **That is why this field is
nullable and why the picker must have a free-text fallback**: a required link
that the product cannot always supply gets faked.

## How to build it

**Shape 5, a contract-level foreign key.** Justified because the link is a
durable product concept, not a screen: an invoice's customer is a permanent
property of the invoice.

**Why not shape 1, a cross-domain read.** A read shows a customer next to an
invoice for as long as the screen is open. It cannot make the invoice _be_ for
that customer, so it cannot answer revenue per customer or survive a rename.
The looser shape genuinely does not satisfy the case.

**Why not shape 6, the shared resolver.** That is `IDENTITY-01` and it is the
right end state. It is also a wave-4 structural project. Waiting for it means
accumulating more unmatchable invoice history in the meantime, and this field
is compatible with the resolver arriving later: `customer_id` becomes a
reference the resolver understands.

Order of work: contract and validator first, then the picker. The
`customer_id` filter on `ListInvoicesRequest` ships in the same change so
`FIN->CRM-01` is unblocked immediately.

## Coupling cost

- **New dependency:** finance depends on crm customers. Edge `crm -> finance`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `crm <-> finance` is bidirectional in this audit
  (`FIN->CRM-01` points back). Resolved in `12-coupling.md`: CRM tells Finance
  who a customer is, Finance tells CRM what they owe, and neither reads the
  other's derived output. **This connection must ship before `FIN->CRM-01`**,
  which depends on the filter it adds.
- **Owner:** **crm** owns the customer contract. **finance** owns the invoice
  field and the snapshot semantics.
- **Degrade when CRM is disabled:** the picker is absent and the invoice form
  behaves exactly as it does today, with the four free-text fields. This is why
  the fields are kept rather than replaced: a Finance-only tenant is a real
  configuration and must not see a broken form. Entitlement is checked through
  `packages/config/src/appGuards.tsx`.

## Risks and correctness notes

- **Permissions.** A Finance user who can issue invoices would gain read access
  to the customer list: names, emails, phones. That is a genuine widening.
  Invoicing requires knowing who you are invoicing, so it is defensible, but
  the picker should read a **minimal** customer projection, not the full
  `crm.Customer`, which also carries `loyalty_points`, `total_spent`, `tags`,
  `notes` and consent flags (`packages/voxie/api/crm/customer/v1/models.onk:16`
  onward). Handing Finance the whole record leaks CRM commentary about the
  customer into a different team's app.
- **Tenancy.** `customer_id` must be validated as belonging to the caller's
  business on write, or an invoice could reference another tenant's customer.
  This is the highest-risk detail: it is a write path taking a foreign id.
- **Migration.** Nullable, no backfill. Matching historical invoices to
  customers by email is tempting and should be a separate reviewed migration
  with an unmatched report, not a silent side effect.
- **Currency and rounding.** Untouched. Invoices keep their own `currency`.
- **Historical rate changes:** not applicable, but the analogous concern is
  handled: renaming a customer does not rewrite issued invoices.
- **Timezone:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 2 days: nullable column, snapshot on create, conditional validator, tenant validation of the foreign id, list filter      |
| Frontend            | 1.5 days: customer picker with free-text fallback, customer column on the invoice list                                    |
| Ongoing maintenance | ~0.5 days per year, rising if the snapshot-versus-live question gets relitigated. The doc comment exists to prevent that. |

## Backend section (round two)

Verified at the storage layer for the first time: `finance_invoices` declares
exactly the four client\_\* columns and no customer column
(`internal/platform/database/schemas/finance/schema.sql`, finance_invoices
block). The contract matches the schema field for field. Nothing in the 63
commits since round one touches this.

_Round two (2026-08-22): confirmed unchanged across schema, contract and code._
