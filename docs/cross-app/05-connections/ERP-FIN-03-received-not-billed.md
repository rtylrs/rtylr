# ERP->FIN-03: purchase order commitments from erp to finance

- **Class:** CAN
- **Necessity:** High value
- **Producer:** erp.purchase_order.v1 (`PurchaseOrderService.listPurchaseOrders`)
- **Consumer:** finance (payables, cash view)
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 1, a cross-domain read from the Finance app
- **Severity:** medium

## Benefit

The day after this ships, Finance can see what has been ordered and received but
not yet billed, so committed spend appears in the payables picture before an
invoice arrives.

## Why this pairing

**Derivation.** ERP raises purchase orders and records receipt. Finance records
bills. Between "we received the goods" and "the supplier invoiced us" there is a
liability nobody sees, and at period end that is an accrual nobody books.

`erp.PurchaseOrder` carries `status`, `quantity_received`, `received_at`,
`total` and `supplier_id`
(`packages/voxie/api/erp/purchase_order/v1/models.onk:19`, `:8`, `:24`, `:22`,
`:16`), and `listPurchaseOrders` filters by `status`
(`packages/voxie/api/erp/purchase_order/v1/models.onk:42`). Everything needed to
find received-and-unbilled orders is served.

**Why Finance and not ERP.** ERP knows about receipt, not about billing. Only
Finance can say whether a bill exists. So the comparison belongs in Finance.

**Why this and not the full cash-forecast seed.** The brief's seed combined
recurring invoices, purchase orders and payroll into a 60-day cash view. That is
three foreign reads on one screen and breaches rule 5. This proposal takes the
one leg that is cheap and already served; the fuller forecast would need a
Finance-owned aggregate and is noted in `08-roadmap.md` rather than proposed here.

**What the user does instead right now.** Asks the person who does the ordering,
or discovers the invoice when it arrives.

## ROI

| Field                  | Value                                                                                                                                                                          |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Affected tenants       | tenants running ERP with Finance and purchasing on account.                                                                                                                    |
| Manual work today      | chase ERP for outstanding orders at period end, or accrue nothing                                                                                                              |
| Modelled monthly value | modelled retail-shaped tenant: assume 30 purchase orders a month and a period-end accrual exercise of **1 hour**, plus better payables visibility in between.                  |
|                        | The decision enabled: knowing committed spend before it is invoiced, which changes whether you place the next order.                                                           |
| Build cost             | backend **0 days**, the endpoint and its status filter exist; frontend 1.5 days (a received-not-billed view and the comparison against bills); maintenance ~0.25 days per year |
| Payback                | fast. Frontend-only against served endpoints.                                                                                                                                  |
| Confidence             | **Inferred from code** for the purchase order fields and the status filter. **Modelled from assumption** for order volume and the hour.                                        |
| Cost of not doing it   | n/a, CAN not MUST. Committed spend stays invisible until invoiced.                                                                                                             |

## The data

| Field               | From               | Cited                                                    |
| ------------------- | ------------------ | -------------------------------------------------------- |
| `status`            | erp purchase order | `packages/voxie/api/erp/purchase_order/v1/models.onk:19` |
| `quantity_received` | erp purchase order | `packages/voxie/api/erp/purchase_order/v1/models.onk:8`  |
| `total`             | erp purchase order | `packages/voxie/api/erp/purchase_order/v1/models.onk:22` |
| `received_at`       | erp purchase order | `packages/voxie/api/erp/purchase_order/v1/models.onk:24` |
| bills               | finance bills      | `packages/voxie/api/finance/bill/v1/models.onk:10`       |

```
received_not_billed = orders where received_at is set
                        and no bill references that order
```

**The subtrahend is the problem, again.** A bill has no `purchase_order_id`
today, so "no bill references that order" cannot be evaluated. Until
`FIN->ERP-02` adds that field, the honest version of this connection is a
**received-orders list that the user reconciles**, not an automatic
received-not-billed figure. Presenting the latter without the link would be
inventing a number. **This proposal therefore ships the visibility and names
`FIN->ERP-02` as what makes it automatic.**

## What exists today

- `listPurchaseOrders` with a status filter, `getPurchaseOrder`,
  `updatePurchaseOrderStatus`, `approvePurchaseOrder`
  (`packages/voxie/api/erp/purchase_order/v1/service.onk:5` onward).
- `BillService.listBills` (`packages/voxie/api/finance/bill/v1/service.onk:5`).
- No reference between a bill and an order.
- **Manual workaround:** asking.

## What is missing

| Wanted                             | Contract reality                        | Backend reality  | What is missing                     |
| ---------------------------------- | --------------------------------------- | ---------------- | ----------------------------------- |
| Received orders visible in Finance | `listPurchaseOrders` with status filter | served           | one read                            |
| Automatic received-not-billed      | bills have no order reference           | unrelated tables | `FIN->ERP-02`'s `purchase_order_id` |

**Bucket: available today** for the visibility; the automatic version is gated
on another proposal.

## Precedent in the codebase

Same as `ERP->FIN-02`: `apps/dash` already reads ERP and POS
(`01-inventory.md`), and this shares the `erp -> finance` edge with `ERP->FIN-01`
and `ERP->FIN-02`. Three connections on one edge is cheaper than three edges,
and it is the reason `08-roadmap.md` groups them.

## How to build it

**Shape 1, a cross-domain read.** Finance reads received purchase orders and
lists them against bills for the user to reconcile.

**Why not shape 2.** Would be justified once this joins the cash forecast with
recurring invoices and payroll, because that screen would then read three
foreign domains. At that point one Finance-owned aggregate replaces three reads.
**That is the trigger to watch**, and it is a real prospect rather than a
hypothetical, since `HR->FIN-01` already puts payroll on Finance's screens.

**Why not shape 3.** Nothing is triggered.

## Coupling cost

- **New dependency:** finance depends on erp purchase orders. Edge
  `erp -> finance`, shared.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `erp <-> finance` bidirectional; `12-coupling.md`.
- **Owner:** **erp** owns purchase orders. Finance owns the view.
- **Degrade when ERP is disabled:** the view is absent. Payables is otherwise
  unchanged. Additive.
- **Two-foreign-reads limit:** if the Finance payables screen shows suppliers
  (`ERP->FIN-01`) and received orders, both are erp, counting as one foreign
  domain. Adding payroll (`HR->FIN-01`) to the same screen would make two, which
  is still within the limit but is the point at which to stop.

## Risks and correctness notes

- **Permissions.** Purchase order totals and suppliers become visible to
  Finance. Appropriate: they will pay them.
- **Currency.** Purchase orders and bills each carry money; the order total's
  currency is not obviously declared on the order message. Do not aggregate
  across currencies; show per-currency subtotals.
- **Rounding.** `float64` both sides.
- **Partial receipts.** `quantity_received` is a quantity, not a flag, so an
  order can be partly received. **A partly received order's liability is not its
  full `total`**, and treating it as such overstates payables. Either restrict
  to fully received orders in the first version, or pro-rate and say so.
- **Timezone.** `received_at` versus a period boundary; state the timezone.
- **Tenancy:** standard.
- **Historical rates:** not applicable.

## Effort

**S.**

| Part                | Estimate                                                                         |
| ------------------- | -------------------------------------------------------------------------------- |
| Backend             | 0 days as scoped                                                                 |
| Frontend            | 1.5 days: received-orders view, partial-receipt handling, per-currency subtotals |
| Ongoing maintenance | ~0.25 days per year                                                              |

## Backend section (round two)

Confirmed: `quantity_received` (`packages/voxie/api/erp/purchase_order/v1/models.onk:8`) and `received_at` (`packages/voxie/api/erp/purchase_order/v1/models.onk:24`)
are on the purchase order contract as cited. No change since round one.

_Round two (2026-08-22): confirmed unchanged._
