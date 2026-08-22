## Benefit

Once hours are billed, they stop appearing as unbilled, so Finance can trust the number instead of re-checking old invoices.

## Why

Single source of truth (3) and gating correctness (6). The unbilled-billable query filters `AND billable_invoice_id IS NULL` (`internal/modules/hr_time/sql/queries/time_entries.sql:52`) and the column exists (`schemas/hr_time/schema.sql:22`), but a repo-wide sweep finds no writer at any layer: the predicate is permanently true and the unbilled number can never decrease. Round one proposed the filters (landed as voxie#46/onion#783); this is the half neither pass saw. Pairing invoices -> hr_time rather than a generic export keeps ownership: hr_time owns its table, invoices call an explicit interface.

## ROI

| Item                   | Value                                                                         |
| ---------------------- | ----------------------------------------------------------------------------- |
| Affected tenants       | any tenant billing time (reference tenant bills ~120 invoices/mo)             |
| Manual work today      | billing staff cross-check exported batches against issued invoices by hand    |
| Modelled monthly value | 4h/mo reconciliation avoided; prevents double-billing incidents (each costly) |
| Build cost             | 1 day backend (interface + write path), included frontend effect              |
| Payback                | under 1 month                                                                 |
| Confidence             | Inferred from code; frequency modelled                                        |

## What exists, what is missing

Column and read filter verified; write path absent everywhere including the invoices and finance_reports modules. Domain struct carries BillableInvoiceID (`hr_time/internal/domain/types.go:24`). The export flow already demonstrates the claim-and-release idempotency pattern to copy (`hr_payroll/internal/application/service.go:377-409`).

## How to build it

Application-facing provider interface (shape B3, 17-backend-plumbing.md): hr_time exposes MarkEntriesInvoiced(ctx, businessID, ids, invoiceID); invoices module receives it at construction like insights receives FlowReader (`main.go:464` precedent). Called in the same transaction as invoice creation or with claim/release if deferred. Add query group to docs/sqlc-query-ownership.md register.

## Coupling cost

Direction invoices -> hr_time via interface; owner hr_time. Tenant predicate business_id in UPDATE WHERE. Index: existing pk-based update suffices. Transaction: single-module atomic write. Degrade: invoices module works unchanged when hr_time disabled for tenant (provider nil-guarded); unbilled counts simply stay as-is.

Dossier: docs/cross-app/05-connections/FLOW-FIN-01-unbilled-billable-hours.md round-two section; gap A1/A2 in docs/cross-app/14-layer-reconciliation.md. Issues rtylr#836, rtylr#855 (filters landed).

Labelled audit:cross-app-round2-2026-08-22, area:invoices, area:hr_time, type:gap, severity:high.
