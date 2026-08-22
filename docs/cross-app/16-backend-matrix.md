# 16. Backend matrix: what already flows between modules

A module-to-module picture of what crosses on the backend today, by mechanism,
with every crossing marked declared or undeclared. This is the mirror of
`03-sharing-matrix.md`: that document asks what could flow between apps; this
one records what already flows between Go modules whether anyone declared it or
not.

Sources, all generated: `backend_tables.csv` (shared tables),
`backend_imports.csv` (import edges), `platform_packages.csv` (the sanctioned
seam), `backend_events.csv` (the async backbone).

## The four mechanisms, and how much each carries

| Mechanism                                               | Volume today                                                                                            | Declared?                                                       |
| ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Shared table (two modules' SQL on one)                  | 24 domain pairs share at least one table; 76 of 296 tables are multi-module                             | Mostly **undeclared**: no registry records them                 |
| Module imports module                                   | 19 edges (see below)                                                                                    | Declared only in the sense that Go forces it; policy says avoid |
| Platform package (`internal/modules/internal/platform`) | 41 packages; the universal ones carry every module                                                      | **Declared**: this is onion's sanctioned seam                   |
| Event / job / webhook                                   | wsbus: UI realtime only. Jobs: 6, all intra-module maintenance. Webhooks: registration with no delivery | Declared where it exists; the integration vocabulary is absent  |

## Shared-table crossings by domain pair

Aggregating `backend_tables.csv` to contract-domain granularity:

| Domain pair          | Tables shared | Examples and notes                                                                                                                      |
| -------------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| console <-> hub      | 8             | addons, business_addons, businesses, manual_access_grants: console's admin projections over hub-owned state                             |
| erp <-> pos          | 5             | erp_products, erp_categories, erp_product_images, stock_levels/locations/ledger: the ERP->POS-01 coupling already flows through storage |
| insights <-> pos     | 5             | pos*order_items, crm_customers, erp*\*: Insights is a cross-domain reader by design                                                     |
| hr <-> insights      | 5             | hr_employees, hr_shifts(+assignments), hr_leave_requests: same reader pattern                                                           |
| finance <-> insights | 5             | finance_invoices/expenses/categories plus hr_shifts                                                                                     |
| finance <-> hr       | 3             | hr_payroll_records (payroll module), hr_shifts, hr_staff                                                                                |
| crm <-> pos          | 2             | crm_customers written by store_orders and platform/customer_sync; businesses                                                            |
| erp <-> insights     | 2             | erp_products, erp_categories                                                                                                            |
| flow <-> hr          | 2             | **hr_time_entries** (workspaces writes, hr_payroll reads via labourcost) and hr_employees                                               |
| flow <-> hub         | 2             | users, hr_employees                                                                                                                     |
| hr <-> hub           | 2             | hr_employees (team writes), hr_locations                                                                                                |
| 12 further pairs     | 1 each        | mostly `businesses` or `users` reach-through                                                                                            |

Classification. **Declared** (sanctioned by a written register or port):
flow<->hr on `internal/modules/hr_time_entries` (onion CLAUDE.md documents ownership;
sqlc-query-ownership.md covers the team/user employee provisioning);
stock_core's ledger writes for pos_stock_movements
(`docs/sqlc-query-ownership.md:34`). Everything else in this table is
**undeclared coupling**: real, load-bearing, and recorded nowhere until this
inventory.

## Import graph between modules

19 direct module-to-module import edges exist (`backend_imports.csv`):

- **hr_core is imported by 11 modules** (staff, payroll, time, shifts, leave,
  attendance, equipment, overtime, setup, performance, workspaces). HR's core
  person record is the product's de facto identity anchor for staff-side data.
- **workspaces imports hr_core and hr_leave** directly (presence join,
  work-day writes): declared by constructor injection at `cmd/api/main.go`.
- **insights imports flow_insights** (`cmd/api/main.go:464` passes the module as a
  `FlowReader`): a clean, narrow provider shape.
- **customers imports campaigns and segments**, pos_orders imports
  pos_settings, hr_setup imports hr_core, business imports user, config
  imports custom_domains, voxhire imports careers.
- **echo imports nothing directly but receives 54 module references as
  ToolExecutor interfaces** at construction (`cmd/api/main.go:547`): the AI assistant
  is a universal consumer through a sanctioned interface, one tool surface per
  module (`internal/modules/*/tool_executor.go`, 61 files).
- **Platform-to-module imports: zero.** The policy boundary holds everywhere.

The brief's rough figures (customers 83, workspaces 73 inbound) do not
reproduce under import-statement parsing; see DECISIONS.md R2-7 for the method
difference. Measured inbound leaders: hr_core 11, everything else 0 to 2.

## Platform seam inventory

All 41 packages verdicted in `platform_packages.csv`. Summary:

- **Universal infrastructure** (>90 importing modules): logger, auth, cache,
  tenant, onekiterrors. Not coupling findings; they are the operating system.
- **Domain-to-domain ports** (declared, live): labourcost (hr -> flow),
  financejournal (finance -> hr), financeaccounts (accounts -> journal),
  hubapp, hubprofile, hubdeveloper, stock_core (erp+pos), customer_sync
  (pos/storefront -> crm), segmentrules.
- **Dead or near-dead**: runmode (zero importers despite documentation),
  models (one).
- **The gap**: no port exists for any customer-context read (finance cannot ask
  crm anything), no supplier/product context port, and no employment-status
  event. Every proposal in this audit either uses an existing port or names the
  new port it needs; none reaches across raw.

## Async backbone

| Mechanism | Publishers                                                                                                                            | Consumers                                            | Verdict for cross-domain use                                     |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------------- |
| wsbus     | workspaces, pos_orders, pos_kds publish per-user/per-workspace/per-business realtime notifications                                    | browser clients via `packages/api/src/realtime.ts`   | Transport live, vocabulary UI-only; not an integration bus today |
| jobs      | 6 registered jobs, all intra-module maintenance (accrual, overtime detection, reminders, recurring invoice generation, CV processing) | worker scheduler                                     | No job propagates state between domains                          |
| webhooks  | CRUD only, exposed to dash via hubdeveloper port                                                                                      | none: no delivery engine exists anywhere in the repo | Registration without emission                                    |

Consequence for proposals: any round-two item shaped as "an event" must either
fill the webhook delivery gap, put a domain event on wsbus (new vocabulary),
or follow the financejournal port pattern for synchronous notification. That
choice is made once per proposal in `17-backend-plumbing.md`.

## Undeclared-coupling watchlist

The rows most likely to surprise a developer, ranked by blast radius times
obscurity:

1. **`erp_products` (14 modules).** POS menu, storefront products and Insights
   all bind to it with no contract-level reference. ERP->POS-01 formalizes the
   smallest piece of this.
2. **`team` writes `internal/modules/hr_employees` on invite** (`internal/modules/team/sql/queries/team.sql:192-201`). Hub seat
   provisioning creates HR records; nothing symmetric exists on termination
   (HR->HUB-01).
3. **`store_orders` writes `crm_customers`** directly alongside
   platform/customer_sync: two writers outside the customers module, so
   customer-record invariants live in three places.
4. **Insights reads five domains' tables raw.** By design, but every schema
   change in hr/finance/erp/pos silently touches Insights' queries.
5. **`pos_orders`/`pos_kds` write into each other's operational tables**
   (order items, session refs) with only the KDS command-boundary note in the
   ownership register covering part of it.
