## Benefit

Revenue per customer becomes answerable: each invoice knows which customer it was issued to.

## Why

Identity resolution (8), single source of truth (3). Invoices store four free-text client\_\* strings only; nothing joins them to crm_customers, so a renamed or re-typed customer splits revenue history irrecoverably. Highest-leverage change product-wide (roadmap). Cross-links: rtylr#857, rtylr#840 (consumer), roadmap docs/cross-app/08-roadmap.md.

## ROI

| Item                   | Value                                                                    |
| ---------------------- | ------------------------------------------------------------------------ |
| Affected tenants       | every finance tenant                                                     |
| Manual work today      | spreadsheet matching of invoice clients to customers                     |
| Modelled monthly value | 6h/mo reporting prep (reference tenant); enables credit decisions (#841) |
| Build cost             | contract 0.25d; onion column+validator+snapshot 2d; form picker 1.5d     |
| Payback                | ~1 month                                                                 |
| Confidence             | Inferred from code; value modelled                                       |
| Cost of not doing it   | unmatched history accrues daily and no migration can repair it later     |

## What exists, what is missing

Drafted on branch: InvoiceDetail.customer_id, CreateInvoiceRequest.customer_id (client_name/client_email relaxed to conditional), ListInvoicesRequest.customer_id filter; snapshot-preservation documented on each field.

## How to build it

Voxie first (this PR); onion adds nullable column, same-business validation, server-side client\_\* snapshot when customer_id supplied; rtylr finance form gains customer picker. Never regenerate gen/ in the drafting PR.

## Coupling cost

New declared edge crm -> finance (identity flows one way; balances return later via FIN->CRM-01). Owner: finance owns invoice, crm owns customer. Nullable: tenants without CRM behave identically.

Labelled audit:cross-app-round2-2026-08-22, area:finance, type:gap, severity:high.
