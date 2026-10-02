# FIN->ERP-01: bill and payment history from finance to erp

- **Class:** CAN
- **Necessity:** High value
- **Producer:** finance.bill.v1 (`BillService.listBills`)
- **Consumer:** erp (supplier page)
- **Why (taxonomy):** 5 context in place
- **Shape:** 1, a cross-domain read from the ERP app
- **Severity:** medium

## Benefit

The day after this ships, an ERP supplier page shows what that supplier has
billed and whether it has been paid, so a buyer negotiating terms knows the
payment history they are negotiating from.

## Why this pairing

**Context in place.** ERP is where purchasing decisions are made. Finance knows
whether the business pays this supplier on time. The buyer cannot see it.

This is the reverse of `ERP->FIN-01` and it moves a different fact: ERP tells
Finance **who** a supplier is, Finance tells ERP **what has been billed and
paid**. Neither reads the other's derived output, which is what keeps the
bidirectional pair acyclic in substance (`12-coupling.md`).

**Why ERP and not a Finance report.** Finance already has bill lists. The value
here is putting the history where the purchasing decision happens, not producing
another report.

**Why it must wait for `ERP->FIN-01`.** Today `supplier_id` on a bill is
fabricated by slugifying a typed name
(`apps/finance/src/features/bills/Bills.tsx:59`). Filtering bills by supplier
before that is fixed would return a supplier's bills only by accident. **This
proposal is worthless until the ids are real**, which is the cleanest sequencing
argument in the audit.

**What the user does instead right now.** Asks Finance, or negotiates without
the history.

## ROI

| Field                  | Value                                                                                                                                                                                      |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Affected tenants       | tenants running ERP with Finance, after `ERP->FIN-01`.                                                                                                                                     |
| Manual work today      | ask Finance for a supplier's payment history                                                                                                                                               |
| Modelled monthly value | modelled retail-shaped tenant: a buyer checks history before perhaps 10 supplier conversations a month, at 5 minutes each: **50 minutes per month.**                                       |
|                        | The decision enabled: whether we are a good customer of this supplier, which affects terms and priority.                                                                                   |
| Build cost             | backend **1 day** (a `supplier_id` filter on `ListBillsRequest`), frontend 1 day (panel on the supplier page), maintenance ~0.25 days per year                                             |
| Payback                | ~3 months on time, once unblocked.                                                                                                                                                         |
| Confidence             | **Inferred from code** that `ListBillsRequest` accepts only `page` and `page_size`, and that supplier ids are currently fabricated. **Modelled from assumption** for the 10 conversations. |
| Cost of not doing it   | n/a, CAN not MUST.                                                                                                                                                                         |

## The data

| Field         | From         | Cited                                              |
| ------------- | ------------ | -------------------------------------------------- |
| `supplier_id` | finance bill | `packages/voxie/api/finance/bill/v1/models.onk:10` |
| `amount`      | finance bill | `packages/voxie/api/finance/bill/v1/models.onk:12` |
| `status`      | finance bill | `packages/voxie/api/finance/bill/v1/models.onk:18` |
| `due_date`    | finance bill | `packages/voxie/api/finance/bill/v1/models.onk:16` |

```
history(supplier) = bills where supplier_id = S, with status and due_date
```

As with `FIN->CRM-01`, **`status` is an unenumerated string**
(`packages/voxie/api/finance/bill/v1/models.onk:18`), so the consumer must not
hardcode which values mean paid. If the vocabulary cannot be enumerated, the
shape becomes 2 and Finance exposes the summary.

## What exists today

- `BillService.listBills` (`packages/voxie/api/finance/bill/v1/service.onk:5`).
- `ListBillsRequest` with `page` and `page_size` only
  (`packages/voxie/api/finance/bill/v1/models.onk:65`).
- ERP's supplier surface (`SupplierService.listSuppliers`,
  `packages/voxie/api/erp/supplier/v1/service.onk:5`).
- **No supplier filter on bills**, and no bill data in ERP.
- **Manual workaround:** asking.

## What is missing

| Wanted                         | Contract reality                                                                                    | Backend reality              | What is missing                                                                                  |
| ------------------------------ | --------------------------------------------------------------------------------------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------ |
| This supplier's bills          | `ListBillsRequest` has no `supplier_id` filter (`packages/voxie/api/finance/bill/v1/models.onk:65`) | `finance_bills` is queryable | a `supplier_id` filter. Paging the whole ledger client-side to find one supplier is not a design |
| Real supplier ids to filter by | `supplier_id` exists but is fabricated by the UI                                                    | never joined                 | `ERP->FIN-01` first                                                                              |

**Bucket: needs a contract field.**

## Precedent in the codebase

The ERP app already makes 4 hub requests and uses OneKit clients throughout
(`01-inventory.md`), so adding a foreign read is routine. There is no precedent
for ERP reading Finance specifically; this would be the first, which is why the
ownership note below matters.

## How to build it

**Shape 1, a cross-domain read**, after `ERP->FIN-01`.

**Why not shape 2.** Only if the paid/unpaid vocabulary cannot be enumerated, as
above. Prefer the read; promote if the definition of paid has to stay inside
Finance.

Sequence: `ERP->FIN-01` (real supplier ids), then the filter, then the panel.

## Coupling cost

- **New dependency:** erp depends on finance bills. Edge `finance -> erp`, which
  is **the reverse of the `erp -> finance` edge** used by three other proposals.
- **Direction:** consumer depends on producer, correct for this fact.
- **Cycle:** this is one half of the `erp <-> finance` bidirectional pair.
  `12-coupling.md` shows the two directions carry different facts and no
  connection consumes the other side's output. Sequencing enforces it in
  practice: the ERP-to-Finance direction ships first.
- **Owner:** **finance** owns bills and payment status. ERP owns the panel.
- **Degrade when Finance is disabled:** the panel is absent from the supplier
  page. Purchasing is unaffected. Additive.

## Risks and correctness notes

- **Permissions, and this one deserves thought.** A buyer or warehouse user
  would see what the business owes and whether it pays late. That is financial
  data in an operations app. Role-gate it, and consider whether it should be
  visible to a warehouse operator at all: the useful audience is purchasing
  managers, not everyone with ERP access.
- **Currency.** Bills carry their own `currency`
  (`packages/voxie/api/finance/bill/v1/models.onk:14`). Per-currency subtotals;
  never sum across.
- **Status vocabulary.** Covered above; drives the shape.
- **Rounding.** `float64`.
- **Timezone.** "Overdue" is relative to today in the tenant's timezone.
- **Tenancy:** standard.
- **Historical rates:** not applicable.

## Effort

**S**, once unblocked.

| Part                | Estimate                                                                                                |
| ------------------- | ------------------------------------------------------------------------------------------------------- |
| Backend             | 1 day: `supplier_id` filter on `ListBillsRequest`, or a summary aggregate if statuses stay unenumerated |
| Frontend            | 1 day: panel on the ERP supplier page with role gating and per-currency subtotals                       |
| Ongoing maintenance | ~0.25 days per year                                                                                     |

## Backend section (round two)

Confirmed: `ListBillsRequest` accepts page and page_size only
(`packages/voxie/api/finance/bill/v1/models.onk`). Blocked-by quality note from
verification: until ERP->FIN-01 lands, filtering bills by supplier would filter
on fabricated ids, exactly as the dossier warns.

_Round two (2026-08-22): confirmed unchanged._
