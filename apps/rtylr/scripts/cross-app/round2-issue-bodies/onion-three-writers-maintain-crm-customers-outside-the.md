## Benefit

Customer records behave the same no matter which surface created or updated them, so renames and merges stop splitting history.

## Why

Single source of truth (3). backend_tables.csv shows crm_customers written by customers, store_orders and platform/customer_sync, and read by eight modules. The sanctioned seam for external writers is customer_sync (POS/storefront upserts with contact matching, `platform/customer_sync/customer_sync.go:18`); store_orders bypasses it with a direct write. Two independent writer paths means validation, name-splitting and visit-count rules drift apart. Consolidation belongs in onion where the writes live.

## ROI

| Item                   | Value                                                      |
| ---------------------- | ---------------------------------------------------------- |
| Affected tenants       | retail tenants using storefront ordering alongside POS/CRM |
| Manual work today      | duplicate customer cleanup by staff                        |
| Modelled monthly value | 30 min/mo cleanup per retail tenant (modelled)             |
| Build cost             | 1 day rerouting writes through customer_sync + tests       |
| Payback                | about 1 month for retail tenants                           |
| Confidence             | Inferred from code                                         |

## What exists, what is missing

Writers: customers (module CRUD), platform/customer_sync UpsertCRMCustomer, store_orders direct INSERT/UPDATE (backend_tables.csv crm_customers row). Reads: insights, kpis, loyalty, segments, feedback, pos side.

## How to build it

Route store_orders customer persistence through customer_sync.UpsertCRMCustomer; keep customers module as canonical CRUD; add the query group to the ownership register. Verify same-business validation parity between paths first.

## Coupling cost

Direction storefront-side module -> crm via platform seam; owner: customers module defines invariants, customer_sync implements matching. Tenant predicate business_id present in sync path; verify store_orders current path. Degrade: storefront order creation must not fail when CRM disabled; fall back to order-local guest record.

Matrix: docs/cross-app/16-backend-matrix.md watchlist item 3. Context: packages/voxie/api/pos/customer/v1/models.onk collision note.

Labelled audit:cross-app-round2-2026-08-22, area:customers, type:gap, severity:medium.
