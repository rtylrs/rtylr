# ERP->POS-01: product stock and cost from erp to pos

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** erp.product.v1 (`Product`)
- **Consumer:** pos (menu items), and Insights via `ERP->INS-01`
- **Why (taxonomy):** 3 single source of truth, 1 derivation, 6 gating
- **Shape:** 5, a contract-level foreign key
- **Severity:** high

## Benefit

The day after this ships, a sellable item knows which stocked product it comes
from, so its margin is computable and its stock figure has one authority instead
of two that can disagree.

## Why this pairing

**Single source of truth, and the clearest duplicate number in the product.**

|                   | `pos.MenuItem`                                                              | `erp.Product`                                                                        |
| ----------------- | --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| stock             | `stock: float64` (`packages/voxie/api/pos/menu_item/v1/models.onk:28`)      | `stock_quantity: float64` (`packages/voxie/api/erp/product/v1/models.onk:15`)        |
| price             | `sale_price: float64` (`packages/voxie/api/pos/menu_item/v1/models.onk:11`) | `selling` side plus `cost_price` (`packages/voxie/api/erp/product/v1/models.onk:13`) |
| cost              | **none**                                                                    | `cost_price`                                                                         |
| link to the other | **none**                                                                    | none                                                                                 |
| sku               | `sku` (`packages/voxie/api/pos/menu_item/v1/models.onk:10`)                 | `sku` (`packages/voxie/api/erp/product/v1/models.onk:9`)                             |

Two independent stock numbers for the same physical thing, and **no cost on the
sellable side at all**, so "what is the margin on this item" cannot be answered.
Both carry a `sku`, which is the only thing that even hints they might be
related, and nothing enforces that the skus match.

**Why ERP is the authority.** ERP is where stock is received, counted, valued
and transferred: it has warehouses, stock takes, stock transfers, valuation and
purchase orders. POS is a selling surface. A selling surface should not be the
system of record for how much of something exists.

**Why the link and not a stock sync.** A sync keeps two numbers and adds a
mechanism for them to disagree in a new way. The link makes the duplication
visible, which is the necessary first step before deciding which number wins.

**What the user does instead right now.** Maintains stock in two places, or in
practice maintains it in one and lets the other drift. **Inferred, not
verified:** which one a tenant maintains depends on which app they use for
receiving.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Affected tenants       | tenants running ERP with POS, which is the retail and restaurant shape RTYLR is sold into. **Not** the reference tenant, which is a services business: the reference figures below are therefore modelled for a different, retail-shaped tenant and labelled as such.                                                                                        |
| Manual work today      | maintaining two stock figures; computing item margin by hand from supplier invoices                                                                                                                                                                                                                                                                          |
| Modelled monthly value | For a modelled retail tenant with 300 sellable items: reconciling POS stock against ERP stock is either never done or costs hours. Assume a modest **2 hours per month** of reconciliation, plus write-offs from selling items believed in stock.                                                                                                            |
|                        | The decision enabled is the bigger half: margin per item at today's supplier prices, which is the difference between a menu that makes money and one that does not. Not quantified here, because doing so would require pricing data this audit does not have.                                                                                               |
| Build cost             | backend **2 days** (nullable `product_id` and a `unit_cost` projection on the menu item, `06-proposed-contracts.md` item 4), frontend 1 day in the ERP/POS admin surfaces that exist in this repo, maintenance ~0.5 days per year. **Plus an unestimated migration** to decide which stock number wins, which is deliberately not folded into this proposal. |
| Payback                | favourable for retail tenants; not applicable to services tenants.                                                                                                                                                                                                                                                                                           |
| Confidence             | **Inferred from code** for both contracts, the two stock fields and the absent cost. **Modelled from assumption** for the 300 items and the 2 hours.                                                                                                                                                                                                         |
| Cost of not doing it   | Item margin is uncomputable, so `ERP->INS-01` (margin per product) can only ever price ERP products, not what is actually sold. Two stock numbers can disagree with no way to detect it, meaning a tenant can sell what it does not have and neither number is provably wrong.                                                                               |

## The data

| Field              | From          | Cited                                                              |
| ------------------ | ------------- | ------------------------------------------------------------------ |
| product `id`       | erp product   | `packages/voxie/api/erp/product/v1/models.onk:8` (`name` adjacent) |
| `cost_price`       | erp product   | `packages/voxie/api/erp/product/v1/models.onk:13`                  |
| `stock_quantity`   | erp product   | `packages/voxie/api/erp/product/v1/models.onk:15`                  |
| `product_id` (new) | pos menu item | proposed, `06-proposed-contracts.md` item 4                        |
| `unit_cost` (new)  | pos menu item | proposed, same                                                     |

Formula, once the link exists:

```
margin(item) = sale_price - unit_cost        (per unit, excluding tax)
```

`pos.MenuItem` already carries `tax_rate`
(`packages/voxie/api/pos/menu_item/v1/models.onk:26`), so the screen must be
explicit about whether the margin shown is pre- or post-tax. Silently mixing
those is a classic source of a wrong number that looks right.

## What exists today

- `erp.Product` with `cost_price`, `stock_quantity`, `sku`, and a full ERP
  surface around it: `getValuation`
  (`packages/voxie/api/erp/inventory/v1/service.onk:9`),
  `listReorderSuggestions` (`packages/voxie/api/erp/product/v1/service.onk:23`),
  `listExpiringItems` (`packages/voxie/api/erp/product/v1/service.onk:25`).
- `pos.MenuItem` with its own `stock` and `sale_price` and no cost.
- `pos.StockMovement.product_id` and `pos.BranchInventory.product_id`
  (`packages/voxie/api/pos/stock/v1/models.onk:8`, `:25`) **already reference
  ERP products.** So POS stock _movements_ know about ERP products while POS
  _menu items_ do not. That inconsistency is the strongest internal evidence
  that the missing link is an oversight rather than a design choice.
- **Manual workaround:** maintaining both.

## What is missing

| Wanted                                 | Contract reality                   | Backend reality                    | What is missing                                                               |
| -------------------------------------- | ---------------------------------- | ---------------------------------- | ----------------------------------------------------------------------------- |
| A sellable item that knows its product | `pos.MenuItem` has no `product_id` | two tables, no join                | a nullable `product_id`                                                       |
| Item margin                            | no cost on the menu item           | `cost_price` exists on the product | a `unit_cost` projection, or a read of the product                            |
| One stock number                       | two independent fields             | two independent columns            | **a decision**, then a migration. The contract field alone does not settle it |

**Bucket: needs a contract field**, with an explicit follow-on decision.

## Precedent in the codebase

`pos.StockMovement.product_id` (`packages/voxie/api/pos/stock/v1/models.onk:8`)
is the precedent, and it is a good one: POS already references ERP products in a
neighbouring message, so the pattern, the id space and the tenancy story are all
established. This proposal extends an existing reference rather than inventing a
relationship, which is the same argument that makes `CRM->FIN-02` credible.

## How to build it

**Shape 5, a contract-level foreign key.** The link is a durable product
concept: which stocked product a sellable item is sold from is a permanent fact
about the item, not a screen.

**Why not shape 1, a cross-domain read.** A read could display an ERP product
next to a menu item, but only if something already told it _which_ product. The
link is the missing piece; a read presupposes it.

**Why not shape 2, an aggregate.** Nothing to aggregate yet.

Sequence, and step 3 is the one that gets skipped if it is not written down:

1. Add nullable `product_id` and the `unit_cost` projection.
2. Provide a way to set the link, with a **reviewed** sku-match suggestion that
   a human confirms. An automatic sku backfill would assert relationships
   nobody checked.
3. **Report divergence**: items whose linked product's `stock_quantity`
   disagrees with the menu item's `stock`. This is what makes the duplication
   visible and is the input to step 4.
4. Decide which number is authoritative and migrate. Jointly owned by the ERP
   and POS owners. Out of scope for this dossier, and stated as such rather
   than implied to be solved.

## Coupling cost

- **New dependency:** pos depends on erp products. Edge `erp -> pos`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `pos -> erp` is marked `would-invert` in the matrix and is not
  proposed, so this pair has a single direction.
- **Owner:** **erp** owns the product and cost. **pos** owns the menu item and
  its reference.
- **Degrade when ERP is disabled:** `product_id` stays null, `unit_cost` is
  absent, and the menu item behaves exactly as today with its own stock and no
  margin. This is why both fields are nullable and why `stock` is left alone:
  a POS-only tenant is a real configuration.
- **Consumer surface caveat.** POS's operator surface lives in a **separate
  repo**. The contract change belongs to Voxie and the reporting can be built in
  the ERP app here, but the menu-item admin experience is not in this monorepo.
  `ERP->POS-02` was rejected for exactly this reason (`11-rejected.md`); this
  one survives because the _contract_ fix and the divergence report are both
  actionable here and because it is the prerequisite for `ERP->INS-01` being
  honest.

## Risks and correctness notes

- **Tenancy.** `product_id` must be validated as belonging to the caller's
  business on write. A write path taking a foreign id is the highest-risk
  detail, as in `CRM->FIN-02`.
- **Permissions.** Exposing `unit_cost` on a menu item shows purchase cost to
  whoever can see the menu. In a restaurant that may include floor staff using
  a terminal. **`unit_cost` should be omitted from the projection served to
  terminal roles** and included only for management, or the product leaks
  supplier pricing to shift workers.
- **Composites and services.** A menu item may be a recipe of several products,
  or a service with no product at all. A single nullable `product_id` cannot
  model a recipe, and pretending otherwise would produce a confidently wrong
  cost. The field is documented as "absent for items that are not stocked as a
  single product", and recipe-level costing is explicitly a separate concept.
- **Currency.** `cost_price` and `sale_price` are both `float64` with currency
  living on ERP settings (`packages/voxie/api/erp/settings/v1/models.onk:7`) and
  on the product (`packages/voxie/api/erp/product/v1/models.onk:28`). A margin
  computed across differing currencies is meaningless; the screen must refuse
  rather than subtract.
- **Rounding.** Both sides are floats. Margin displayed to two decimals from
  float inputs will occasionally show a cent that does not reconcile with a
  ledger. Acceptable for a display metric; state it.
- **Historical cost.** `cost_price` is current. A margin on a past sale valued
  at today's cost is not the margin that was earned. Label the figure as
  valued at current cost, the same discipline as `HR->FLOW-01`.
- **Timezone:** not applicable.

## Effort

**M.**

| Part                | Estimate                                                                                                              |
| ------------------- | --------------------------------------------------------------------------------------------------------------------- |
| Backend             | 2 days: nullable `product_id`, `unit_cost` projection with role-based omission, tenant validation, divergence report  |
| Frontend            | 1 day in this repo (ERP-side reporting and linking). The POS terminal experience is out of repo.                      |
| Ongoing maintenance | ~0.5 days per year, plus the deferred stock-authority migration which is unestimated and should be tracked separately |

## Backend section (round two)

Confirmed unchanged: MenuItem still declares neither a product reference nor a
cost field. Storage-layer context this round adds: `erp_products` is the most
shared business table in the product after `businesses` itself (14 modules
touch it, `backend_tables.csv`), so the join target is heavily contended and
the eventual stock-authority migration should be planned with that readership
in mind.

_Round two (2026-08-22): confirmed; shared-table pressure documented._
