# ERP->FIN-02: inventory valuation from erp to finance

- **Class:** CAN
- **Necessity:** High value
- **Producer:** erp.inventory.v1 (`InventoryService.getValuation`)
- **Consumer:** finance (balance sheet)
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 1, a cross-domain read from the Finance app
- **Severity:** medium

## Benefit

The day after this ships, the balance sheet shows the value of stock on hand
from the system that counts it, so inventory as an asset stops being a figure
somebody types in.

## Why this pairing

**Derivation and correctness of a statutory report.** Inventory is an asset on
the balance sheet. ERP knows what stock exists and what it cost. Finance draws
the balance sheet. Nothing connects them, so either the figure is manual or the
balance sheet omits inventory.

The producer is ready: `InventoryService.getValuation`
(`packages/voxie/api/erp/inventory/v1/service.onk:9`) with `total_value`,
`total_cost` and per-item `stock_value` already declared
(`packages/voxie/api/erp/inventory/v1/models.onk:21`, `:22`, `:13`). ERP also has
stock takes and a valuation surface around it.

**Why Finance reads rather than ERP posting.** Rule 4: ERP must not write
Finance's balance sheet. And the balance sheet is a presentation, not a stored
record, so there is nothing to write.

**What the user does instead right now.** Reads the ERP valuation screen and
types the number into the balance sheet, or leaves inventory out. **Inferred,
not verified:** the Finance app makes no ERP-domain requests at all
(`01-inventory.md` shows finance reading finance, crm, hub and insights), so the
figure cannot be arriving automatically.

## ROI

| Field                  | Value                                                                                                                                                                                          |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running ERP with Finance, so retail and restaurant shapes rather than the services reference tenant.                                                                                   |
| Manual work today      | copy the valuation across at each reporting date, or omit inventory                                                                                                                            |
| Modelled monthly value | one balance sheet per month, perhaps 20 minutes to source and cross-check the inventory figure: **20 minutes per month.**                                                                      |
|                        | Correctness matters more than time: a balance sheet with a stale or omitted inventory figure misstates assets, and for any tenant with financed stock that is a covenant and lending question. |
| Build cost             | backend **0 days**, the endpoint exists; frontend 1 day (read and render on the balance sheet with an as-at date); maintenance ~0.25 days per year                                             |
| Payback                | immediate; this is close to free.                                                                                                                                                              |
| Confidence             | **Inferred from code** that the valuation endpoint and fields exist and that Finance makes no ERP requests today. **Modelled from assumption** for the 20 minutes.                             |
| Cost of not doing it   | n/a, CAN not MUST, though a misstated balance sheet is arguably closer to MUST for a tenant that files accounts from it. Kept as CAN because the product does not claim to file accounts.      |

## The data

| Field         | From           | Cited                                                                |
| ------------- | -------------- | -------------------------------------------------------------------- |
| `total_value` | erp inventory  | `packages/voxie/api/erp/inventory/v1/models.onk:21`                  |
| `total_cost`  | erp inventory  | `packages/voxie/api/erp/inventory/v1/models.onk:22`                  |
| balance sheet | finance report | `packages/voxie/api/finance/report/v1/service.onk:5` (`getPLReport`) |

**Which figure is the asset?** `total_value` and `total_cost` are different
numbers and the contract's doc comments do not distinguish them beyond their
names. Inventory on a balance sheet is normally carried at cost (or the lower of
cost and net realisable value), so `total_cost` is the likely correct input and
`total_value` is likely a retail-value figure. **This must be confirmed with the
ERP owner before shipping**, and it is exactly the sort of ambiguity that
produces a plausible wrong number. Flagged rather than guessed.

## What exists today

- `getValuation`, `getInventoryReport`, `getLowStock`, `adjustInventory`
  (`packages/voxie/api/erp/inventory/v1/service.onk:5` onward).
- `erp.stock_take` capability for physical counts.
- Finance's balance sheet screen (`/balance-sheet` route, `apps/finance`).
- **No ERP requests from the Finance app.**

## What is missing

| Wanted                               | Contract reality                                        | Backend reality | What is missing                                               |
| ------------------------------------ | ------------------------------------------------------- | --------------- | ------------------------------------------------------------- |
| Inventory value on the balance sheet | `getValuation` exists with totals                       | ERP computes it | one read from the Finance balance-sheet screen                |
| Value as at a reporting date         | **no date parameter observed on the valuation request** | current stock   | valuation is "now", not "as at 31 March". See the risk below. |

**Bucket: available today**, with a significant caveat about as-at dating.

## Precedent in the codebase

`apps/dash` reads `erp` once and `pos` nine times (`01-inventory.md`), so a
non-ERP app reading ERP is established. `ERP->FIN-01` (suppliers) creates the
same edge, so if both ship the Finance app has one ERP dependency rather than
two independent ones.

## How to build it

**Shape 1, a cross-domain read.** The balance sheet reads the valuation and
renders it as the inventory line, labelled with the moment it was taken.

**Why shape 1 suffices.** One foreign domain, one call, no permission
asymmetry worth an aggregate: a Finance user preparing a balance sheet may
certainly see total stock value.

**Why not shape 2.** It would be justified if the balance sheet needed several
foreign domains. Today it needs one.

## Coupling cost

- **New dependency:** finance depends on erp inventory. Edge `erp -> finance`,
  shared with `ERP->FIN-01` and `ERP->FIN-03`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `erp <-> finance` is bidirectional; `12-coupling.md` resolves it.
- **Owner:** **erp** owns valuation and its definition. Finance owns the
  presentation.
- **Degrade when ERP is disabled:** the inventory line is absent from the
  balance sheet, which is correct: a tenant with no ERP has no tracked stock.
  It must not render zero, because zero inventory is a meaningful and false
  assertion on a balance sheet. This is the clearest case in the audit where
  "hide, do not show zero" is a correctness requirement rather than a nicety.

## Risks and correctness notes

- **As-at dating is the main risk.** A balance sheet is drawn at a date. If the
  valuation endpoint returns only current stock, then a balance sheet for last
  month shows this month's inventory, which is wrong in a way that looks right.
  Either the endpoint gains an as-at parameter, or **the figure must be labelled
  as current and the report must not be presented as a dated balance sheet
  line.** Do not silently place a live figure in a dated column.
- **Cost versus value.** Covered above. Confirm before shipping.
- **Currency.** ERP settings carry a `@required` currency
  (`packages/voxie/api/erp/settings/v1/models.onk:7`) and products carry their
  own (`packages/voxie/api/erp/product/v1/models.onk:28`). If a tenant's ERP
  currency and Finance currency differ there is no conversion authority in the
  product (`02-entity-graph.md`). Show the currency and refuse to place a
  foreign-currency figure into the report unconverted.
- **Rounding.** `float64` on both sides; round once at display.
- **Permissions.** Total stock value to a Finance user is a modest widening and
  a defensible one. No per-item cost need cross for this line.
- **Tenancy:** standard.
- **Historical rates:** the as-at problem is the analogous concern.

## Effort

**S.**

| Part                | Estimate                                                                                                     |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| Backend             | 0 days as scoped. Add an as-at parameter if the dated version is wanted, which would be a larger ERP change. |
| Frontend            | 1 day: read, render, label the valuation moment and currency                                                 |
| Ongoing maintenance | ~0.25 days per year                                                                                          |

## Backend section (round two)

Strengthened: `GetValuation` is implemented server-side beyond the contract,
through the full stack
(`internal/modules/inventory/internal/api/inventory_service.go:57`,
`internal/modules/inventory/internal/application/service.go:82`). This dossier needs exactly what it said: one read
from the balance-sheet screen.

_Round two (2026-08-22): confirmed; backend implementation verified._
