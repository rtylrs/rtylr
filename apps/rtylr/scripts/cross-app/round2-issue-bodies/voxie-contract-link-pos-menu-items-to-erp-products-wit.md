## Benefit

Menu items know what they cost, making per-item margin computable for the first time.

## Why

Derivation (1), single source of truth (3). MenuItem has stock and sale_price but no product reference and no cost; ERP holds both. Two stock numbers for one physical thing can disagree undetectably today. Cross-links: rtylr#858, rtylr#843.

## ROI

| Item                   | Value                                                            |
| ---------------------- | ---------------------------------------------------------------- |
| Affected tenants       | retail/restaurant tenants running ERP+POS                        |
| Manual work today      | margin computed offline from supplier sheets                     |
| Modelled monthly value | pricing decisions enabled; shrinkage visible (unquantified)      |
| Build cost             | contract 0.25d; onion link+validation 1d; pos UI 2d (other repo) |
| Payback                | 2 months (retail segment)                                        |
| Confidence             | Inferred from code                                               |

## What exists, what is missing

Drafted on branch: MenuItem.product_id @nullable, MenuItem.unit_cost @nullable with authoritative-value documentation. ERP->INS-01 margin work builds on this.

## How to build it

Voxie; then onion menu module validates same-business product link and snapshots unit_cost on read; pos repo consumes. Stock-authority unification stays a separate ERP/POS decision, deliberately out of this draft.

## Coupling cost

Declared edge erp -> pos. Nullable: non-stocked items unaffected. Server must reject foreign-business product ids.

Labelled audit:cross-app-round2-2026-08-22, area:pos, type:gap, severity:high.
