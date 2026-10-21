# FIN->ERP-02: bill settlement from finance to erp

- **Class:** CAN
- **Necessity:** High value
- **Producer:** finance.bill.v1 (`Bill.status`)
- **Consumer:** erp (purchase orders)
- **Why (taxonomy):** 2 double entry elimination, 4 trigger and automation
- **Shape:** 3, an event emitted by Finance
- **Severity:** medium

## Benefit

The day after this ships, paying a supplier's bill closes out the purchase order
it settles, so purchasing stops showing orders as outstanding after they have
been paid.

## Why this pairing

**Trigger.** Finance marks a bill paid. ERP still shows the purchase order as
open, because nothing tells it. Somebody closes the order by hand, or nobody
does and the outstanding-orders list becomes untrustworthy.

`erp.PurchaseOrderService.updatePurchaseOrderStatus` exists
(`packages/voxie/api/erp/purchase_order/v1/service.onk:11`) and bills carry a
`status` (`packages/voxie/api/finance/bill/v1/models.onk:18`). The pieces exist;
nothing connects them, and **the bill has no reference to the order it settles.**

**Why an event and not a direct write.** Rule 4 is decisive: purchase order
status is ERP's record, and Finance must not write it. Finance emits "this bill
was paid, and it settled purchase order N"; ERP decides what that means for the
order. ERP might close it, might partially close it, might ignore it for an order
paid across several bills.

**Why this ranks below `ERP->FIN-03`.** That one gives Finance visibility with
zero backend work. This one needs a contract field and an event. Visibility
first, automation second.

**What the user does instead right now.** Updates the purchase order status
manually, or lets the outstanding list drift.

## ROI

| Field                  | Value                                                                                                                                                                                                             |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running ERP with Finance and purchasing on account.                                                                                                                                                       |
| Manual work today      | close the purchase order separately after paying the bill                                                                                                                                                         |
| Modelled monthly value | modelled retail-shaped tenant with 30 purchase orders a month: closing each by hand costs perhaps 1 minute, so **30 minutes per month**, and the orders that are never closed degrade `ERP->FIN-03`'s usefulness. |
| Build cost             | backend **3 days** (`purchase_order_id` on the bill, the event, ERP's consumer, partial-settlement handling), frontend 0.5 days (order picker on the bill form), maintenance ~0.5 days per year                   |
| Payback                | ~6 months on time alone. Better counted as what keeps `ERP->FIN-03` honest.                                                                                                                                       |
| Confidence             | **Inferred from code** that the status endpoint exists and no bill-to-order reference does. **Modelled from assumption** for order volume.                                                                        |
| Cost of not doing it   | n/a, CAN not MUST. But note the compounding: without it, `ERP->FIN-03`'s received-not-billed view has to be reconciled by hand forever.                                                                           |

## The data

| Field                     | From               | Cited                                                     |
| ------------------------- | ------------------ | --------------------------------------------------------- |
| `status`                  | finance bill       | `packages/voxie/api/finance/bill/v1/models.onk:18`        |
| `amount`                  | finance bill       | `packages/voxie/api/finance/bill/v1/models.onk:12`        |
| `purchase_order_id` (new) | finance bill       | proposed, `06-proposed-contracts.md` item 5               |
| order status updated      | erp purchase order | `packages/voxie/api/erp/purchase_order/v1/service.onk:11` |

No formula. One rule, and it is ERP's to make: what a settled bill implies for
an order's status.

## What exists today

- `updatePurchaseOrderStatus` and `approvePurchaseOrder`
  (`packages/voxie/api/erp/purchase_order/v1/service.onk:11`, `:13`).
- Bills with a `status`.
- A webhooks module and a notifications module in the backend
  (`~/Development/onion/internal/modules/webhooks`,
  `~/Development/onion/internal/modules/notifications`), so event plumbing exists
  in some form and should be surveyed before adding another mechanism.
- **No reference between a bill and an order.**
- **Manual workaround:** closing orders by hand.

## What is missing

| Wanted                         | Contract reality         | Backend reality  | What is missing                                                          |
| ------------------------------ | ------------------------ | ---------------- | ------------------------------------------------------------------------ |
| Which order a bill settles     | no reference on the bill | unrelated tables | a nullable `purchase_order_id`                                           |
| ERP told when it is paid       | no event                 | nothing          | an emission from Finance and a consumer in ERP                           |
| Orders paid over several bills | nothing models it        | nothing          | partial settlement handling, or an explicit decision to support only 1:1 |

**Bucket: needs a contract field**, plus an event.

## Precedent in the codebase

`HR->HUB-01` is the other event proposal, and the same reasoning applies: both
modules live in one Go binary, so a direct call is simpler and should be
considered first. The argument for the event is ownership: Finance should not
know ERP's status vocabulary. **Whichever mechanism is chosen, choose the same
one for both proposals**, rather than introducing two event styles.

The cautionary precedent is `internal/modules/hr_category_id`: do not add
`purchase_order_id` before the consumer exists, or it becomes another declared
link nobody reads (`02-entity-graph.md`).

## How to build it

**Shape 3, an event.**

**Why not shape 1, a read.** ERP would have to poll Finance for bill statuses,
which inverts the dependency (ERP depending on Finance for its own record's
state) and does work continuously to catch an occasional change.

**Why not shape 2.** No aggregation.

**Why not shape 5 alone.** The `purchase_order_id` field is necessary but not
sufficient: without a notification, ERP learns nothing until someone looks.

Sequence: field first, then the picker so the field gets populated, then the
event once there is data for it to be about. That order also means the field is
never live-but-unread for long.

## Coupling cost

- **New dependency:** erp depends on a Finance event. Edge `finance -> erp`,
  shared with `FIN->ERP-01`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** the `erp <-> finance` pair again; `12-coupling.md`.
- **Owner:** **finance** owns the event contract and the bill's reference.
  **erp** owns what the event means for an order.
- **Degrade when Finance is disabled:** no events arrive, `purchase_order_id`
  stays null, and purchase orders are closed manually as today. ERP never
  errors and never blocks. Additive.

## Risks and correctness notes

- **Partial settlement is the main correctness risk.** One order may be paid by
  several bills, or one bill may cover several orders. A single nullable
  `purchase_order_id` models 1:1 only. **Either restrict to 1:1 and say so in
  the field's doc comment, or model it properly.** Silently closing an order on
  the first partial payment would understate liabilities, which is the worse
  failure.
- **Idempotency and ordering.** Events can arrive twice or out of order. Closing
  an order must be idempotent, and a later "unpaid" correction must be able to
  reopen it.
- **Permissions.** The event carries a bill amount and status into ERP. Keep the
  payload minimal: order id and settlement state, not the full bill.
- **Currency.** A bill's currency and an order's total may differ. ERP should
  not attempt reconciliation of amounts across currencies; the event should
  carry settlement state rather than an amount to compare.
- **Tenancy.** The event must carry `business_id`, and ERP must scope by it.
- **Rounding.** Avoided by not comparing amounts.
- **Timezone, historical rates:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                               |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| Backend             | 3 days: nullable `purchase_order_id`, emission, ERP consumer, idempotency, partial-settlement decision |
| Frontend            | 0.5 days: order picker on the bill form                                                                |
| Ongoing maintenance | ~0.5 days per year. Higher than a read: events carry delivery, ordering and retry concerns.            |

## Backend section (round two)

Confirmed: Bill carries no purchase order reference; ERP owns
`updatePurchaseOrderStatus` as assumed
(`packages/voxie/api/erp/purchase_order/v1/service.onk`). Webhook context from
the events sweep strengthens the dossier's shape choice: the webhooks module is
registration-only CRUD with no delivery engine anywhere in the repo
(`backend_events.csv`), so an event-shaped implementation cannot assume tenant
webhooks exist; wsbus carries UI realtime topics only. The in-process port
pattern used by financejournal/financeaccounts is the codebase's working
precedent for exactly this direction of notification.

_Round two (2026-08-22): confirmed; substrate evidence updated._
