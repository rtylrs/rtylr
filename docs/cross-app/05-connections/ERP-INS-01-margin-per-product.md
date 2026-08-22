# ERP->INS-01: product cost from erp to insights

- **Class:** CAN
- **Necessity:** High value
- **Producer:** erp.product.v1 (`Product.cost_price`)
- **Consumer:** insights (product analytics)
- **Why (taxonomy):** 1 derivation, 7 rollup and reporting
- **Shape:** 2, a backend aggregate owned by Insights
- **Severity:** medium

## Benefit

The day after this ships, the product analytics screen shows margin rather than
just revenue, so a best seller that loses money stops being called a best seller.

## Why this pairing

**Rollup that is currently only half honest.** Insights already reports revenue
per product: `insights.BestSeller` and `insights.SlowMover` both carry a
`product_id` referencing ERP products
(`packages/voxie/api/insights/product/v1/models.onk:5`, `:16`). **Neither carries
cost.** So the analytics app ranks products by revenue and calls the top ones
best sellers, which is a statement about turnover masquerading as a statement
about value.

ERP holds `cost_price` on the product
(`packages/voxie/api/erp/product/v1/models.onk:13`). The join key already exists
in the Insights contract. Nothing uses it for cost.

**Why Insights owns the aggregate.** Insights owns the revenue side and is the
reporting domain; it should read ERP cost server-side and emit margin. The
alternative, the Insights app fetching the product catalogue and joining
client-side, means an unbounded catalogue read per report and duplicated join
logic in the browser.

**Why not have ERP produce margin.** ERP does not know revenue. It would have to
read Insights, which inverts the direction and makes a source domain depend on a
projection.

**What the user does instead right now.** Exports sales and costs and joins them
in a spreadsheet, or accepts a revenue ranking as if it were a margin ranking.

## ROI

| Field                  | Value                                                                                                                                                                                                     |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running Insights with ERP, so retail and restaurant shapes.                                                                                                                                       |
| Manual work today      | spreadsheet joins of sales against cost, or no margin view at all                                                                                                                                         |
| Modelled monthly value | modelled retail-shaped tenant: a monthly margin review built by hand costs perhaps **2 hours**. The decision enabled is larger: which products to delist or reprice. Not quantified without pricing data. |
| Build cost             | backend **3 days** (an Insights-owned aggregate joining its revenue rows to ERP cost, plus the contract in `06-proposed-contracts.md` item 8b), frontend 1 day, maintenance ~0.5 days per year            |
| Payback                | ~2 to 3 months on the modelled 2 hours.                                                                                                                                                                   |
| Confidence             | **Inferred from code** that Insights already references `product_id` and carries no cost, and that ERP has `cost_price`. **Modelled from assumption** for the 2 hours.                                    |
| Cost of not doing it   | n/a, CAN not MUST, but note that the product currently **presents a revenue ranking under the label "best seller"**, which is a mild form of the confidently-wrong-number problem this audit is about.    |

## The data

| Field          | From                 | Cited                                                 |
| -------------- | -------------------- | ----------------------------------------------------- |
| `product_id`   | insights best seller | `packages/voxie/api/insights/product/v1/models.onk:5` |
| revenue, units | insights             | same capability                                       |
| `cost_price`   | erp product          | `packages/voxie/api/erp/product/v1/models.onk:13`     |

```
margin(product, period) = revenue(product, period)
                        - units_sold(product, period) * cost_price(product)
```

**Two honesty requirements**, both encoded in the proposed contract:

- `cost` is **nullable**. A product with no cost price on record is different
  from a product that costs nothing, and collapsing the two understates cost and
  overstates margin.
- The cost used is the **current** cost price, not the cost at the time of sale.
  There is no cost history, so a margin on last quarter's sales valued at today's
  supplier price is an approximation. Label it.

## What exists today

- `insights.product` capability with `BestSeller` and `SlowMover`, both carrying
  `product_id` (`packages/voxie/api/insights/product/v1/models.onk:5`, `:16`).
- `erp.Product.cost_price` and `ProductService.listProducts`
  (`packages/voxie/api/erp/product/v1/service.onk:5`).
- Insights makes 28 requests, all within its own domain plus one hub call
  (`01-inventory.md`), so it reads **no** foreign business domain today.
- **Manual workaround:** spreadsheets.

## What is missing

| Wanted                               | Contract reality                                   | Backend reality                             | What is missing                                                                                   |
| ------------------------------------ | -------------------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| Margin per product                   | `BestSeller` has `product_id` and revenue, no cost | ERP has cost; Insights has revenue; no join | an Insights-owned aggregate that reads ERP cost                                                   |
| Honest gaps                          | no nullable cost concept                           | products may have no cost price             | a nullable `cost` so missing data is visible                                                      |
| Sold-item margin, not product margin | `pos.MenuItem` has no cost or product link         | separate tables                             | `ERP->POS-01`. Without it this reports on ERP products, not on what was actually sold through POS |

**Bucket: needs a query or endpoint.**

## Precedent in the codebase

The relevant precedent is the direction problem: `apps/finance` reads
`/insights/financial/cash-flow`, making a source domain depend on a projection
(`10-doc-drift.md`). This proposal is the correct inverse, Insights reading a
source fact. Keeping that direction consistent is why the matrix marks every
`insights -> X` cell as `insights-is-terminal`.

## How to build it

**Shape 2, a backend aggregate owned by the consumer.**

**Why not shape 1, a cross-domain read.** The Insights app would have to fetch
the product catalogue to get costs, which is an unbounded read per report and
puts the join in the browser. Rule 5's spirit applies even at one foreign
domain when the read is a whole table.

**Why Insights and not ERP owns it.** Discussed above: ERP does not know revenue.

Sequence: contract, aggregate, screen. `ERP->POS-01` is a soft prerequisite for
the number to describe what was actually sold rather than what is stocked; ship
without it, but say what the figure covers.

## Coupling cost

- **New dependency:** insights depends on erp products. Edge `erp -> insights`.
- **Direction:** consumer depends on producer. Correct. **This is Insights'
  first dependency on a business domain other than its own**, which is worth
  noting: it establishes a pattern, so the aggregate should be built as the
  model for how Insights reaches other domains.
- **Cycle:** none. `insights -> erp` is not proposed and never should be.
- **Owner:** **erp** owns cost. **insights** owns the margin definition and the
  aggregate.
- **Degrade when ERP is disabled:** the margin column is absent and the revenue
  ranking remains, **relabelled** so it does not claim to be a margin. Insights
  never errors. The relabelling matters: silently keeping the "best seller"
  framing when cost is unavailable is the current state, and shipping this
  should fix the label whether or not cost is present.

## Risks and correctness notes

- **Permissions.** Product cost is commercially sensitive: it reveals supplier
  pricing. Insights users are usually management, so this is a modest widening,
  but the margin figure should be role-gated the same way other financial
  analytics are.
- **Missing cost is the main correctness risk.** Covered above with the nullable
  `cost`. A report that shows 100% margin because cost was null would be worse
  than no report.
- **Current versus historical cost.** Covered above; label it.
- **Currency.** Products carry their own `currency`
  (`packages/voxie/api/erp/product/v1/models.onk:28`) and ERP settings carry a
  required one (`packages/voxie/api/erp/settings/v1/models.onk:7`). Revenue and
  cost must be in the same currency or the subtraction is meaningless. Refuse
  rather than convert: there is no conversion authority (`02-entity-graph.md`).
- **What was sold versus what is stocked.** Without `ERP->POS-01` this describes
  ERP products. For a restaurant selling menu items, that is not the same set.
  State the scope on the screen.
- **Rounding.** `float64` throughout; round once at display.
- **Tenancy:** standard.

## Effort

**M.**

| Part                | Estimate                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Backend             | 3 days: `ProductMargin` contract, Insights-owned aggregate reading ERP cost, nullable-cost and currency-mismatch handling |
| Frontend            | 1 day: margin column, missing-cost indication, scope and current-cost labelling                                           |
| Ongoing maintenance | ~0.5 days per year                                                                                                        |

## Backend section (round two)

Confirmed, plus one cheapening nuance: Insights' own SQL already reads
`erp_products` directly (`backend_tables.csv`, erp_products row lists insights
among readers). An Insights-owned margin aggregate therefore needs no new
cross-domain mechanism, only a query in its existing query set. Draft 8b stands
as written.

_Round two (2026-08-22): confirmed; build cost reduced._
