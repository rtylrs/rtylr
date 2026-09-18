# FIN->CRM-01: outstanding balance from finance to crm

- **Class:** CAN
- **Necessity:** High value
- **Producer:** finance.invoices.v1 (`InvoicesService.listInvoices`)
- **Consumer:** crm (customer page, quoting)
- **Why (taxonomy):** 5 context in place, 6 gating and correctness
- **Shape:** 1, a cross-domain read from the CRM app
- **Severity:** medium

## Benefit

The day after this ships, a CRM customer page shows what that customer owes and
whether anything is overdue, so nobody offers new terms to someone who has not
paid for the last thing.

## Why this pairing

**Context in place, and a soft gate.** CRM is where a salesperson decides what
to offer. Finance knows whether the customer pays. The salesperson cannot see it.

**This proposal is blocked until `CRM->FIN-02` ships.** Invoices have no
customer reference: they carry `client_name` and `client_email` as free strings
(`packages/voxie/api/finance/invoices/v1/models.onk:30`, `:62`). There is
nothing to filter by, so "this customer's invoices" is not a query that can be
written. That is the hardest dependency edge in the audit and the reason this
dossier is short: the interesting design work is in `CRM->FIN-02`.

**Why a signal and not a block.** Extending credit to a slow payer is a
commercial judgement that varies by tenant and by customer. A hard block would
be RTYLR making a business rule, which the configurability rules forbid. Show
the balance and the overdue count; let the human decide.

**What the user does instead right now.** Asks Finance, or does not check.

## ROI

| Field                  | Value                                                                                                                                                                 |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants using CRM with Finance, after `CRM->FIN-02`.                                                                                                                  |
| Manual work today      | ask a colleague, or open Finance and search by customer name                                                                                                          |
| Modelled monthly value | reference tenant: assume a salesperson checks payment status before ~20 quotes or renewals a month, at 4 minutes each when they bother: **80 minutes per month.**     |
|                        | The leakage side: credit extended to non-payers. Not quantified, because doing so needs the tenant's own bad-debt figures.                                            |
| Build cost             | backend **1 day** (the `customer_id` filter, which ships with `CRM->FIN-02`), frontend 1.5 days (balance panel on the customer page), maintenance ~0.25 days per year |
| Payback                | ~2 months on time, once unblocked. The build itself is cheap; the prerequisite is not.                                                                                |
| Confidence             | **Inferred from code** that no customer reference exists on invoices. **Modelled from assumption** for the 20 checks and 4 minutes.                                   |
| Cost of not doing it   | n/a, CAN not MUST.                                                                                                                                                    |

## The data

| Field                | From            | Cited                                                  |
| -------------------- | --------------- | ------------------------------------------------------ |
| invoice `status`     | finance invoice | `packages/voxie/api/finance/invoices/v1/models.onk:32` |
| invoice `total`      | finance invoice | `packages/voxie/api/finance/invoices/v1/models.onk:20` |
| `customer_id` filter | proposed        | `06-proposed-contracts.md` item 3                      |

```
outstanding(customer) = sum(total) over invoices where customer_id = C
                                              and status is unpaid
overdue(customer)     = the same, restricted to due_date < today
```

**"Unpaid" is a status string with no enumerated vocabulary in the contract**
(`status: string`, `packages/voxie/api/finance/invoices/v1/models.onk:32`). The
consumer must not hardcode a list of status values it believes mean unpaid: that
is a business rule masquerading as a constant, and a tenant renaming a status
would silently break the total. Either the contract enumerates the values, or
Finance exposes the aggregate itself (see below).

## What exists today

- `InvoicesService.listInvoices` (`packages/voxie/api/finance/invoices/v1/service.onk:5`)
  and `updateInvoiceStatus` (`:13`).
- `InvoicePaymentService.recordPayment`
  (`packages/voxie/api/finance/invoices/payments/v1/service.onk:5`).
- CRM's customer page.
- **No link between them.**
- **Manual workaround:** asking Finance.

## What is missing

| Wanted                   | Contract reality                    | Backend reality                   | What is missing                                         |
| ------------------------ | ----------------------------------- | --------------------------------- | ------------------------------------------------------- |
| This customer's invoices | invoices have no customer reference | `finance_invoices` stores strings | `CRM->FIN-02` first, then a `customer_id` filter        |
| What counts as unpaid    | `status: string`, unenumerated      | free string                       | either enumerated statuses or a Finance-owned aggregate |

**Bucket: needs a contract field**, and gated behind another proposal.

## Precedent in the codebase

CRM already reads a foreign domain: the app makes 2 support-domain requests
(`01-inventory.md`). So a foreign panel on a CRM screen is established. Note
that if both this and `SUP->CRM-01` ship, the customer page reads **two** foreign
domains, which is exactly the limit in rule 5. **A third panel on that page
requires a CRM-owned aggregate instead**, and that is the trigger condition to
watch, not a hypothetical.

## How to build it

**Shape 1, a cross-domain read**, with a caveat that may promote it to shape 2.

If the invoice `status` vocabulary cannot be enumerated in the contract, the
correct shape becomes **2**: Finance exposes "outstanding and overdue for this
customer" as an aggregate, so the definition of unpaid stays inside the domain
that owns it. That is the looser-shape-fails test being applied honestly rather
than defaulting to shape 1 because it is cheaper.

Sequence: `CRM->FIN-02` first. Then the filter. Then the panel.

## Coupling cost

- **New dependency:** crm depends on finance invoices. Edge `finance -> crm`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `crm <-> finance` is bidirectional. `12-coupling.md` resolves it:
  CRM tells Finance who a customer is, Finance tells CRM what they owe. Neither
  reads the other's derived output. **Order matters:** `CRM->FIN-02` ships first,
  and this one consumes the field it adds.
- **Owner:** **finance** owns invoices and the definition of outstanding. CRM
  owns the panel.
- **Degrade when Finance is disabled:** the panel is absent. The customer page
  is otherwise unchanged. Additive, never blocking.

## Risks and correctness notes

- **Permissions, and it is a real widening.** A salesperson would see customer
  financial data. Many businesses want exactly that; some deliberately separate
  sales from collections. This must be **role-gated and ideally
  tenant-configurable**, not shown to every CRM user by default.
- **Two-foreign-reads limit.** Discussed above. This is the page where the limit
  actually binds.
- **Currency.** Invoices each carry their own `currency`
  (`packages/voxie/api/finance/invoices/v1/models.onk:40`). A customer with
  invoices in two currencies has no single outstanding total. **Show per-currency
  subtotals and refuse to sum across them.** Summing would produce a confidently
  wrong number, which is the failure mode this audit exists to remove.
- **Rounding.** `float64` totals; sum in minor units where possible.
- **Timezone.** "Overdue" depends on today in the tenant's timezone, not the
  browser's.
- **Status vocabulary.** Covered above; it drives the shape choice.
- **Tenancy:** standard scoping.

## Effort

**S**, once unblocked.

| Part                | Estimate                                                            |
| ------------------- | ------------------------------------------------------------------- |
| Backend             | 1 day for the filter, or 2 if it becomes a Finance-owned aggregate  |
| Frontend            | 1.5 days: balance panel with per-currency subtotals and role gating |
| Ongoing maintenance | ~0.25 days per year                                                 |

## Backend section (round two)

Confirmed blocked exactly as stated: invoices carry no customer reference at
any layer, so nothing to filter by. Dependency on CRM->FIN-02 unchanged.

_Round two (2026-08-22): confirmed unchanged._
