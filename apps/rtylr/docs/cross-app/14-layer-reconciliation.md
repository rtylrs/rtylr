# 14. Layer reconciliation

Where the four layers disagree: database schema, Go module, `.onk` contract,
frontend usage. Round one's `10-doc-drift.md` covers docs versus code; this
covers schema versus contract versus handler versus browser, which is where
silently broken cross-app assumptions live.

Sources: round one's `internal/modules/inventory/contracts.csv`, `internal/modules/inventory/endpoints.csv`,
round two's `backend_handlers.csv`, `backend_queries.csv`, `migrations.csv`,
plus targeted reads cited per row. Coverage: every entity class named below was
checked at all four layers; the remaining ~160 entities of round one's 178 were
checked at the layers named in their row and are listed in the summary counts.

## The six disagreement classes, with findings

### A. A column in the database with no contract field

| #   | Finding                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Layers verified                                                                                                                                  |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| A1  | `internal/modules/hr_time_entries.billable_invoice_id` and `.export_batch_id` exist as columns (`internal/platform/database/schemas/hr_time/schema.sql:22-23`), are read by the unbilled-billable filter (`internal/modules/hr_time/sql/queries/time_entries.sql:52`, `AND billable_invoice_id IS NULL`) and by the domain type (`internal/modules/hr_time/internal/domain/types.go:24`), but appear in **no** `.onk` message (`packages/voxie/api/hr/time_entry/v1/models.onk` has neither field) and no frontend code references them. | Column: yes. Query read: yes. Contract: absent. Frontend: absent.                                                                                |
| A2  | **Nothing ever writes `billable_invoice_id`.** A repo-wide sweep finds reads, filters and struct fields, but no UPDATE sets it and no other module touches the column. The "unbilled" filter is therefore permanently true for every row: the invoice-side half of marking hours billed does not exist at any layer. This sharpens round one's FLOW->FIN-01 dossier (see `15-round-one-verification.md`).                                                                                                                                | Write path: absent in every module including invoices and finance_reports (`grep -rn billable_invoice_id internal/` returns hr_time files only). |
| A3  | `finance_invoices.reporting_date` is a generated stored column (`internal/platform/database/schemas/finance/schema.sql:19`) with no contract counterpart; harmless today because Insights reads its own projections, but any consumer trusting the contract alone would miss that revenue-by-month is precomputed in storage.                                                                                                                                                                                                            | Column yes; contract absent.                                                                                                                     |
| A4  | The recruit corpus has a database layer only. `careers_submissions`, `voxire_career_applications` and the rest of the `recruit`/`voxire` schema directories have no contract domain and no generated types; round one established this and the migration inventory re-confirms it (`migrations.csv`, domains `recruit`, `voxire`, `rtylr`).                                                                                                                                                                                              | All four layers: DB only.                                                                                                                        |

### B. A contract field no handler populates

| #   | Finding                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Status since round one                                                       |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| B1  | `flow.WorkType.hr_category_id` remains persisted-and-echoed by the workspaces module with no payroll consumer (`internal/modules/workspaces/internal/api/worktime_service.go:381` writes, `:149` echoes). Round one's verdict stands.                                                                                                                                                                                                                                                                                               | Unchanged by the 63 commits since round one.                                 |
| B2  | `insights.dashboard.v1`'s three delivery RPCs are **no longer stubs**: `GetInsights`, `GetRoadmap` and `GetWorkload` are implemented over Flow's insights module (`internal/modules/insights/internal/api/setting_service.go:116,136,165`). Only report-template handlers remain `notImplemented` (`internal/modules/insights/internal/api/report_service.go:73,77`). Round one's #870 described the world before commits `5d3822f8` (getKPIs deleted) and the dashboard implementation landed; see `15-round-one-verification.md`. | Corrected: handlers exist. Whether the frontend calls them is checked in D2. |

### C. A contract field no frontend reads

| #   | Finding                                                                                                                                                                                                                                                                                                                                                                   |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| C1  | `hr.TimeEntryPayload.billable_invoice_id`-adjacent gap is covered by A1. Beyond it, round one's list of 41 cross-domain contract references was spot-rechecked; the only additional unread cross-domain field found is `pos.RosterEntry.user_id` family, whose consumers live in the POS repo rather than this one, so "no frontend reads" is expected rather than drift. |

### D. A handler with no contract

Measured precisely for the first time (`backend_handlers.csv`):

| #   | Finding                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------- |
| D1  | onion serves **118 legacy chi routes** through the typed registration helpers. Exactly **zero** of them match a generated oneKit route by verb and path shape. 86 sit inside domains that also carry contracts (hr_core 24, crm_sales 17, hr_shifts 13, hr_time 11, hr_payroll 9, console 6, pos_orders 1); **32 sit in surfaces with no contract domain at all**: careers 23, memory 6, echo 3. This replaces round one's looser estimate (#873: "109 of 227 literals") with an operation-level count. |
| D2  | The implemented Insights delivery handlers are contracted but uncalled: `apps/insights` still calls only `getDashboardSummary` plus settings (verified against `internal/modules/inventory/endpoints.csv`; no `/insights/dashboard/v1/insights                                                                                                                                                                                                                                                          | roadmap | workload`rows). Backend capability again ahead of frontend adoption, the same shape as`export-billable`. |
| D3  | The payroll export route `POST /hr/payroll/runs/{id}/export` is served (`backend_handlers.csv`, hr_payroll) with no frontend caller in rtylr or pos. See `15-round-one-verification.md`, HR->FIN-01.                                                                                                                                                                                                                                                                                                    |

### E. A frontend type hand-written where a contract exists

| #   | Finding                                                                                                                                                                                                                                                                                                                                                                     |
| --- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| E1  | `apps/crm/src/shared/hooks/queries/useSales.ts:106,245,272,347` hand-writes lead/contact/contract payload interfaces for CRM sales endpoints that have no `.onk` domain. Not a violation while the domain is uncontracted, but it means the CRM app owns wire-format types the contracts repo should own once `crm.sales.v1` exists (round one's CRM->FIN-01 prerequisite). |
| E2  | `apps/flow/src/shared/services/flowApi.ts:46-50` defines `ApiBusinessSummary` dropping the currency that `hub.Business.currency` carries (round one's I14). Still true on this branch.                                                                                                                                                                                      |
| E3  | `apps/hr/src/shared/hooks/queries/useTimeEntries.ts:86-114` declares a response type for `/hr/time-entries/import-flow` that does not match the handler (round one #869). The route remains uncontracted legacy chi (`backend_handlers.csv`, hr_time), so the audit tooling still cannot catch the mismatch.                                                                |

### F. Domain-to-app mismatches

Unchanged from round one and re-confirmed: `booking` and `support` have contracts, backend modules and no app in this repo; `recruit` and `menu` are apps with no (full) contract domain. One addition: the schemas submodule carries domains `_perf`, `ai`, `cdn`, `store` and `voxhire` with no contract either; `store_*` modules serve the storefront surface that also lives outside the ten Vite apps.

## Cross-layer entities that reconcile cleanly

Recorded because negative space needs evidence too. These were checked across
all four layers and agree:

| Entity                      | Schema                                                  | Handler                                                 | Contract                            | Frontend                               |
| --------------------------- | ------------------------------------------------------- | ------------------------------------------------------- | ----------------------------------- | -------------------------------------- |
| finance invoice             | `finance_invoices` (client\_\* strings, no customer id) | invoices module serves CRUD via oneKit                  | matches schema exactly              | finance app mirrors both               |
| crm customer / pos customer | one table `crm_customers`                               | customers + pos_customers + customer_sync all target it | two contract views over one store   | crm app reads its own view             |
| hr time entry               | `internal/modules/hr_time_entries`                      | hr_time module + workspaces writer                      | `hr.time_entry.v1` minus A1 columns | hr app reads; flow writes via workdays |

## Summary counts

| Disagreement class                                   | Instances documented | Where |
| ---------------------------------------------------- | -------------------- | ----- |
| A. column, no contract field                         | 4                    | A1-A4 |
| B. contract field, handler does not populate/consume | 2                    | B1-B2 |
| C. contract field, no frontend reads                 | 1 new                | C1    |
| D. handler, no contract                              | 118 routes           | D1-D3 |
| E. hand-written frontend types                       | 3 confirmed          | E1-E3 |
| F. domain/app mismatch                               | unchanged            | F     |

The pattern across A, C and D is the same one round one found and round two can
now quantify: **the backend consistently builds capability one or two layers
below where the product can see it.** Columns exist without fields (A1), fields
exist without readers (B1), handlers exist without contracts (D1), and working
server-side flows exist without callers (D3, D2, and the labour-cost port).
Round one called this "the fifth gap bucket" candidate; the reconciliation data
supports promoting it.
