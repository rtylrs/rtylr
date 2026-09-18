## Benefit

Margin per product becomes visible next to the existing margin distribution, without pulling the ERP catalogue.

## Why

Rollup (7), derivation (1). getProductMargins exists and is implemented but returns distribution buckets only; ERP->INS-01 needs per-product rows carrying cost. Narrow additive extension beats round-ones standalone ProductMargin proposal since Insights already reads erp_products server-side. Cross-links: rtylr#853 (superseded in part), rtylr#853 consumer.

## ROI

| Item                   | Value                                           |
| ---------------------- | ----------------------------------------------- |
| Affected tenants       | ERP+Insights tenants                            |
| Manual work today      | export join between sales report and price list |
| Modelled monthly value | 2h/mo analyst time (retail reference)           |
| Build cost             | contract 0.1d; onion query 0.75d; chart 1d      |
| Payback                | ~2 months (retail segment)                      |
| Confidence             | Inferred from code                              |

## What exists, what is missing

Drafted on branch: ProductMargin message (cost_minor_units nullable when no cost on record) plus products field on GetProductMarginsResponse.

## How to build it

Voxie; onion extends GetMarginDistribution path with per-product query (own query group, register entry); dash/insights charts consume.

## Coupling cost

Edge erp -> insights already flows via raw table reads (undeclared today); this contract declares the projection formally. Nullable cost distinguishes unknown from zero.

Labelled audit:cross-app-round2-2026-08-22, area:insights, type:gap, severity:medium.
