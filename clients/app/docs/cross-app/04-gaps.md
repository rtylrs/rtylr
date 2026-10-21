# 04. Gap classification

For every CAN and MUST proposal: what is wanted, what the contract actually
says, what the backend actually does, and therefore what is missing. The table
format follows `docs/flow-time-tracking.md`, which is this repo's worked
example of the difference between an endpoint existing and a field existing.

The four buckets, and how the 21 proposals fall into them:

| Bucket                        | Count | Meaning                                                           |
| ----------------------------- | ----- | ----------------------------------------------------------------- |
| **Available today**           | 4     | Frontend work only. Every field and endpoint already exists.      |
| **Needs a query or endpoint** | 6     | The data is in the database; no RPC exposes it in a usable shape. |
| **Needs a contract field**    | 9     | The concept exists; the field or filter does not.                 |
| **Needs a new concept**       | 2     | Neither side has it.                                              |

## Available today: frontend work only

| Id          | Wanted                                                                            | Contract reality                                                                                                                                                                                                                                                          | Backend reality                                                                                                                                   | What is missing                                                                                                                                                           |
| ----------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ERP->FIN-01 | Pick a supplier from ERP when creating a bill, and show its name on the bill list | `SupplierService.listSuppliers` exists (`packages/voxie/api/erp/supplier/v1/service.onk:5`); `finance.Bill.supplier_id` is `@required @gt(0)` (`packages/voxie/api/finance/bill/v1/models.onk:29`)                                                                        | `erp_suppliers` is a real table; the `bills` module never joins it                                                                                | **Nothing on the wire.** The Finance app must call `listSuppliers` and send a real id instead of slugifying a typed name (`apps/finance/src/features/bills/Bills.tsx:59`) |
| ERP->FIN-02 | Inventory value on the balance sheet                                              | `InventoryService.getValuation` exists (`packages/voxie/api/erp/inventory/v1/service.onk:9`); `total_value` and `total_cost` are declared (`packages/voxie/api/erp/inventory/v1/models.onk:21`)                                                                           | ERP computes and returns it                                                                                                                       | One read from the Finance balance-sheet screen                                                                                                                            |
| ERP->FIN-03 | What has been ordered and received but not yet billed                             | `PurchaseOrderService.listPurchaseOrders` (`packages/voxie/api/erp/purchase_order/v1/service.onk:5`), with `status`, `quantity_received`, `received_at`, `total` (`packages/voxie/api/erp/purchase_order/v1/models.onk:19`, `:8`, `:24`, `:22`)                           | ERP owns and serves it                                                                                                                            | One filtered read, plus the comparison against bills, which Finance already has                                                                                           |
| HR->FIN-01  | Payroll run total posted to the journal as a labour expense                       | `PayrollService.listPayroll` returns `period_start`, `period_end`, `net_pay` (`packages/voxie/api/hr/payroll/v1/models.onk:7`, `:17`); `JournalService.createJournalEntry` takes `date`, `description`, `lines[]` (`packages/voxie/api/finance/journal/v1/models.onk:25`) | Both sides are served. Payroll already computes the money (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:221`) | The action that reads one and writes the other. Finance writes its own journal, so no ownership rule is broken                                                            |

## Needs a query or endpoint

The data exists in the database. No RPC returns it in a shape a consumer can
use, usually because the aggregation or the permission boundary has to live on
the server.

| Id          | Wanted                                                    | Contract reality                                                                                                                                                                                                                                                                                                                                    | Backend reality                                                                                                                                                                                                                                                                     | What is missing                                                                                                                                                       |
| ----------- | --------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HR->FLOW-01 | Labour cost of a project                                  | No cost-per-project message exists anywhere                                                                                                                                                                                                                                                                                                         | Every input is present: `internal/modules/hr_time_entries` carries `project_id` and `duration_minutes`, `internal/modules/hr_employees` carries `hourly_rate` (`~/Development/onion/internal/platform/database/schemas/hr_core/schema.sql:91`), and payroll already multiplies them | **A new HR-owned aggregate RPC returning cost per project, never rates per person.** This is a permission boundary, not a convenience: see the RBAC note below        |
| HR->INS-01  | Labour cost per period for margin metrics                 | `insights` declares no labour-cost message                                                                                                                                                                                                                                                                                                          | Same inputs as above                                                                                                                                                                                                                                                                | The same aggregate, consumed by Insights. Deliberately the same endpoint, so the number cannot differ between Flow and Insights                                       |
| ERP->INS-01 | Margin per product, not just revenue per product          | `insights.BestSeller` carries `product_id` and revenue (`packages/voxie/api/insights/product/v1/models.onk:5`) but no cost                                                                                                                                                                                                                          | `erp_products.cost_price` exists (`packages/voxie/api/erp/product/v1/models.onk:13`)                                                                                                                                                                                                | An Insights-owned aggregate that joins its own revenue rows to ERP cost. Reading `listProducts` from the Insights app would be an unbounded catalogue read per report |
| REC->HR-01  | Turn a hired candidate into an employee                   | Recruit has **no contract domain at all**; `careers_submissions` has `candidate_name`, `email`, `phone`, `resume_url`, and a status funnel ending in `hired` (`~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:80`). `EmployeeService.createEmployee` exists (`packages/voxie/api/hr/employee/v1/service.onk:9`) | The careers module serves plain routes and imports no generated contract (`~/Development/onion/internal/modules/careers/module.go`)                                                                                                                                                 | A contracted read of one submission, so HR can prefill an employee form from it. The funnel already has a `hired` state that means nothing                            |
| POS->CRM-01 | Order history and lifetime spend on the CRM customer page | `pos.Order.customer_id` is a real join to `crm_customers` (`packages/voxie/api/pos/orders/v1/models.onk:35`), but `ListOrdersRequest` requires `session_id` as `@required @gt(0)` (`packages/voxie/api/pos/orders/v1/models.onk:145`)                                                                                                               | POS customers read `crm_customers` directly (`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:33`)                                                                                                                                         | **A customer-scoped order query.** Orders are listable only within one till session today, so a customer's history across sessions cannot be asked for at all         |
| HR->HUB-01  | Terminating an employee revokes app access                | `hub.App` has `is_enabled` (`packages/voxie/api/hub/app/v1/models.onk:8`); `hr.EmployeePayload.status` includes `terminated` (`packages/voxie/api/hr/employee/v1/models.onk:19`)                                                                                                                                                                    | Termination sets one column and nothing else (`~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:426`)                                                                                                                                     | A server-side consequence. Nothing in `apps/**` can fix this: the revocation must happen where the status changes                                                     |

### RBAC note, severity critical if done wrong

`HR->FLOW-01` and `HR->INS-01` are the two proposals that could leak salary
data. `hr.Staff` exposes `salary`, `hourly_rate` and `overtime_rate`
(`packages/voxie/api/hr/staff/v1/models.onk:17`). If Flow satisfied its labour
cost need by reading employee records and multiplying client-side, **any Flow
workspace member who can open the network tab would see every colleague's pay
rate.** That is why both are specified as shape 2 (a backend aggregate owned by
the producing domain) returning only a total, and why bar rule 5 is not the
binding constraint here: permissions are.

The audit found **no** existing path by which a Flow member can read HR salary
data. The 11 Flow-to-HR requests are staff detail and time endpoints. This is a
risk to avoid creating, not a bug to report.

## Needs a contract field

The concept exists on both sides. A field or, more often, a **query filter** is
missing. Draft `.onk` for all of these is in `06-proposed-contracts.md`.

| Id           | Wanted                                                                | Contract reality                                                                                                                                                                                                                                                              | Backend reality                                                                                          | What is missing                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------ | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| FLOW->FIN-01 | Which logged hours are billable and unbilled, by project              | `hr.TimeEntryPayload` already carries `project_id`, `task_id`, `is_billable` and `status` (`packages/voxie/api/hr/time_entry/v1/models.onk:12`, `:14`, `:28`, `:30`)                                                                                                          | `internal/modules/hr_time_entries` holds the rows, and Flow writes into them with `source = flow_import` | `ListTimeEntriesRequest` filters only by `employee_id` and a **single** `date` (`packages/voxie/api/hr/time_entry/v1/models.onk:71`). No date range, no `project_id`, no `is_billable`. There is also an uncontracted `/hr/time-entries/export-billable` route that returns exactly this shape and **no frontend calls it** (`~/Development/onion/internal/modules/hr_time/internal/routes/handlers/export_billable_handler.go:15`) |
| HR->FLOW-02  | Who is on approved leave during this cycle                            | `internal/modules/hr_leave_requests` carries `employee_id`, `start_date`, `end_date`, `status` (`~/Development/onion/internal/platform/database/schemas/hr_leave/schema.sql:24`) and `LeaveService.listLeaveRequests` exists (`packages/voxie/api/hr/leave/v1/service.onk:6`) | Served and configurable: leave types are rows, not an enum                                               | `ListLeaveRequestsRequest` filters by `employee_id` and `status` only (`packages/voxie/api/hr/leave/v1/models.onk:87`). Asking "who is away between these dates" requires a date range                                                                                                                                                                                                                                              |
| CRM->FIN-01  | Draft an invoice from a signed contract                               | CRM **sales has no contract domain**: `crm_contracts` (`lead_id`, `plan`, `mrr`, `seats`, `signed_at`) exists only as SQL and legacy routes                                                                                                                                   | `crm_sales` module serves `/crm/sales/contracts`                                                         | A contracted read of a signed contract. Today the CRM app reaches it through uncontracted paths (`apps/crm/src/shared/hooks/queries/useSales.ts:52`)                                                                                                                                                                                                                                                                                |
| CRM->FIN-02  | An invoice that knows which customer it is for                        | `finance.Invoice` stores `client_name`, `client_email`, `client_address`, `client_tax_id`, the first two `@required` (`packages/voxie/api/finance/invoices/v1/models.onk:146`)                                                                                                | `finance_invoices` stores the strings                                                                    | A nullable `customer_id` on the invoice, and the denormalised name kept for historical accuracy                                                                                                                                                                                                                                                                                                                                     |
| FIN->CRM-01  | Outstanding balance on the customer page                              | `InvoicesService.listInvoices` exists with `status`                                                                                                                                                                                                                           | Served                                                                                                   | Nothing to filter by: without `CRM->FIN-02` there is no customer id on an invoice, so this proposal **depends on that one**. It is the clearest dependency edge in the roadmap                                                                                                                                                                                                                                                      |
| SUP->CRM-01  | Open tickets on the CRM customer page                                 | `support.Customer` already carries `total_tickets` and `last_ticket_at` (`packages/voxie/api/support/customer/v1/models.onk:3`), and `CustomerService.getCustomer` exists (`packages/voxie/api/support/customer/v1/service.onk:6`)                                            | `support_customers` is a separate table from `crm_customers`                                             | A link. Matching on email works until an email differs, so the honest version needs a customer reference on the support customer                                                                                                                                                                                                                                                                                                    |
| ERP->POS-01  | One stock number and one cost per sellable item                       | `pos.MenuItem` has its own `stock` and `sale_price` and **no cost and no product reference** (`packages/voxie/api/pos/menu_item/v1/models.onk:28`, `:11`); `erp.Product` has `stock_quantity` and `cost_price` (`packages/voxie/api/erp/product/v1/models.onk:15`, `:13`)     | Two tables, two numbers, no join                                                                         | A `product_id` on the menu item. Until then item margin is uncomputable and the two stock figures can disagree without either being detectably wrong                                                                                                                                                                                                                                                                                |
| FIN->ERP-01  | Bill and payment history on the ERP supplier page                     | `BillService.listBills` exists but `ListBillsRequest` accepts only `page` and `page_size` (`packages/voxie/api/finance/bill/v1/models.onk:65`)                                                                                                                                | `finance_bills` is queryable, and never joined to `erp_suppliers`                                        | A `supplier_id` filter. Paging the whole bill ledger client-side to find one supplier's bills is not a design, it is an accident waiting for a large tenant                                                                                                                                                                                                                                                                         |
| FIN->ERP-02  | A purchase order that stops looking outstanding once its bill is paid | `finance.Bill` has `status` and `supplier_id` but no purchase order reference; `PurchaseOrderService.updatePurchaseOrderStatus` exists (`packages/voxie/api/erp/purchase_order/v1/service.onk:11`)                                                                            | Both served, unrelated                                                                                   | A `purchase_order_id` on the bill. Note the direction: ERP owns the status change, so Finance must emit rather than write (bar rule 4)                                                                                                                                                                                                                                                                                              |

## Needs a new concept

| Id          | Wanted                                                       | Contract reality                                                                     | Backend reality                                                                                                                                                                                        | What is missing                                                                                                                                                                                                                                                                           |
| ----------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| IDENTITY-01 | "This person" and "this company" mean one thing product-wide | Eight separate identity entities; 18 of 28 pairs have no join (`02-entity-graph.md`) | One resolver query exists, hub user to hr employee (`~/Development/onion/internal/modules/hr_core/sql/queries/employees.sql:47`)                                                                       | A party or contact concept that the domains reference, rather than each holding its own copy                                                                                                                                                                                              |
| HR->HR-01   | One HR person record with one pay rate                       | `hr.Staff` and `hr.EmployeePayload` are separate messages, both carrying pay fields  | `internal/modules/hr_staff` and `internal/modules/hr_employees` are separate tables, both with `salary`, `hourly_rate`, `overtime_rate`, bridged by a nullable `internal/modules/hr_staff.employee_id` | A decision about which is canonical, then a migration. Also two parallel leave systems: `internal/modules/hr_time_off_requests` keys on `staff_id` with a **hardcoded ENUM** of leave types, while `internal/modules/hr_leave_requests` keys on `employee_id` with configurable type rows |

## Two systemic findings the per-proposal tables kept running into

### 13 HR list endpoints cannot be paginated by their callers

Measured across the whole corpus: 17 `List*Response` messages declare `limit`
and `offset`, but only 4 of their matching `List*Request` messages accept them.
The 13 mismatches are all in `hr`:

`ListAttendanceRequest`, `ListPayrollRunsRequest`, `ListAssetsRequest`,
`ListEquipmentRequestsRequest`, `ListLeaveRequestsRequest`,
`ListLeaveTypesRequest`, `ListLeaveBalancesRequest`,
`ListChecklistTemplatesRequest`, `ListOvertimeRequestsRequest`,
`ListReviewCyclesRequest`, `ListGoalsRequest`, `ListTimeEntriesRequest`,
`ListTimesheetsRequest`.

The response tells the caller which page it received and the caller has no way
to ask for a different one. This violates this repo's own rule that every list
query must be paginated (root `CLAUDE.md`, performance rules), and it matters
for this audit specifically: `HR->FLOW-02` and `FLOW->FIN-01` both want to read
HR lists for a whole team or a whole project, which today means an unbounded
read. **Every contract draft in `06-proposed-contracts.md` that touches one of
these requests adds `limit` and `offset` alongside the new filter**, because
adding a filter to an unpaginated list endpoint makes the unbounded read
easier to trigger rather than harder.

### A third of the contract work is filters, not fields

Of the nine "needs a contract field" items, three need only a **query
parameter** over data that is already stored and already returned in the right
shape: a date range on time entries (`FLOW->FIN-01`), a date range on leave
requests (`HR->FLOW-02`), and a supplier filter on bills (`FIN->ERP-01`). Those
are the cheapest contract changes in the audit, and two of them are what stands
between the product and the labour-cost question in the brief.

The other six need a stored fact: `customer_id` on an invoice, `customer_id` on
a support customer, `product_id` on a menu item, `purchase_order_id` on a bill,
a contracted CRM sales surface, and the relaxation of
`finance.Bill.supplier_id`. Those carry migrations, and `06-proposed-contracts.md`
states each one.

### Pagination style is inconsistent across domains

A smaller observation that any cross-domain consumer runs into immediately:
`hr` paginates with `limit`/`offset`
(`packages/voxie/api/hr/employee/v1/models.onk:115`), while `finance`, `erp` and
`pos` paginate with `page`/`page_size`
(`packages/voxie/api/finance/bill/v1/models.onk:66`,
`packages/voxie/api/erp/purchase_order/v1/models.onk:44`,
`packages/voxie/api/pos/orders/v1/models.onk:153`). A shared cross-domain hook
therefore cannot have one pagination helper. This is not worth a migration on
its own, but it is worth knowing before anyone proposes shape 4 (a shared query
hook in `packages/`) for a paginated cross-domain list.

## Round two: the fifth bucket, already existing on the backend

The brief predicted this bucket would be larger than anyone thinks. It is.
Measured against onion at `ab6ff7ff` and voxie `847a34c8` (both one day after
round one shipped), seven capabilities are complete or nearly complete behind
the API surface with no consumer seeing them. Ranked by how much work remains
to turn each into something a customer can touch:

| #   | Capability                                              | Backend state                                                                                                                | Contract | Frontend | Remaining work                                                            |
| --- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | -------- | -------- | ------------------------------------------------------------------------- |
| 1   | Project labour cost (`HR->FLOW-01`)                     | Live platform port: SQL joins hours to rates inside HR; Flow's service holds it with a noop fallback (`cmd/api/main.go:466`) | none     | none     | One RPC on Flow's project surface plus a panel                            |
| 2   | Payroll export to journal (`HR->FIN-01`)                | Complete: route, idempotent claim, Finance-owned poster port, tenant account mapping                                         | none     | none     | One button in HR plus an optional posted marker                           |
| 3   | Delivery dashboards for Insights (flow-deep-dive O5-O7) | Implemented over a narrow FlowReader port (`internal/modules/insights/internal/api/setting_service.go:116-165`)              | exists   | none     | Hooks and screens in apps/insights                                        |
| 4   | Billable-hours export (`export-billable`)               | Live since before round one; uncalled                                                                                        | none     | none     | Decide adopt-or-delete (round one recommended delete once filters landed) |
| 5   | Time-entry range/project/billability filters (#855)     | Landed upstream after round one: voxie #46, onion #783                                                                       | done     | partial  | Consume the new filters in Finance's unbilled view                        |
| 6   | Work-schedule reads (flow-deep-dive #868)               | Fixed end to end: `listWorkSchedules` served and consumed                                                                    | done     | done     | Nothing                                                                   |
| 7   | Inventory valuation (`ERP->FIN-02`)                     | Implemented server-side (`internal/modules/inventory/internal/api/inventory_service.go:57`)                                  | exists   | none     | The single balance-sheet read                                             |

Two structural causes, both verifiable:

1. **Contract lag.** The labour-cost port and the payroll export have no `.onk`
   definition at all, so `pnpm contracts:audit-requests` cannot see them and no
   frontend generator can bind to them. Backend capability outruns contracts
   because contracts are a separate repo with a separate review loop.
2. **No consumer in the same change.** The codebase's own history
   (`internal/modules/hr_category_id`, `export-billable`) shows capability without a consumer in
   the same PR stays unconsumed. Items 1 and 2 above are the newest instances.

### Backend column for the round-one gap tables

Every round-one row was re-verified against handlers and SQL rather than the
`.onk` alone; verdicts live in `15-round-one-verification.md`. Rows whose
bucket changed:

- FLOW->FIN-01: moves from **needs a contract field** to **available today**
  (filters landed upstream). A new storage prerequisite replaces it: nothing
  writes `billable_invoice_id`, so "unbilled" cannot change value yet
  (`14-layer-reconciliation.md` A2).
- HR->FLOW-01 / HR->INS-01: move from **needs a query or endpoint** to the
  fifth bucket above.
- HR->FIN-01: stays **available today**, but shrinks: even the action endpoint
  exists.

All other rows confirmed unchanged at every layer.
