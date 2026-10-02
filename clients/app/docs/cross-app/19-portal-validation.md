# Portal feature and shared-data validation

This checklist describes the migrated Portal at the API, database and browser
boundaries. Passing the existing suites is not exhaustive feature certification.
The earlier cross-app proposals describe design intent; verify them against the
current contracts and modules before treating them as implemented integrations.

## Evidence and limits

| Area                   | Evidence                                                                                                                                                                 | Limit                                                                                               |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| ERP → POS catalog      | `pos_catalog_integration_test.go`: create, update price/stock/image, deactivate; POS-only member can read; purchase cost omitted; wrong business denied                  | Real MySQL and generated HTTP handlers; resolved auth context injected, not a browser login         |
| ERP → warm POS cache   | `crossAppProducts.test.tsx`: actual ERP mutation and POS query share a QueryClient and refresh immediately                                                               | API boundary mocked; same browser session only                                                      |
| POS checkout numbering | `integration_refund_money_test.go`: eight simultaneous register/public checkouts; unique order numbers                                                                   | Local MySQL fixture workload, not production load testing                                           |
| Console counts         | `stats` database transition tests                                                                                                                                        | Live aggregate reads replace unsupported maintenance triggers; large-database query cost unmeasured |
| Employee RBAC          | `make -C services/onion authorization-evidence`; documented direct-handler cases and license lifecycle tests                                                             | Operation ledger covers 104 records; permission catalog contains 127 keys                           |
| Browser behavior       | Fixture-backed Playwright on desktop and 390px; ERP create/edit/delete, CRM customer/loyalty, Finance, HR, Flow, Insights, Recruit, public Menu, Portal POS, Automations | API responses stubbed and external traffic blocked; not evidence of live integrations               |
| Native POS             | POS Vitest and TypeScript/build checks                                                                                                                                   | Web checks do not validate Tauri hardware, printers, cash drawers or native offline recovery        |

Reproduce database checks with `services/onion/scripts/test-db.sh` on the dedicated
local `onion_test` database. Do not reset a customer database. Browser specs use
`clients/app/e2e/fixtures.ts` to block unstubbed external traffic.

## Feature acceptance checklist

Every feature needs a successful read/write cycle where applicable, persisted
state after reload, an honest failure/retry path, tenant isolation, and the
minimum allowed member permission. Hidden navigation alone is insufficient.

| App                        | Features to sign off                                                                                                                                                                                             | Existing browser sweep                                                                            | Still needs end-to-end data/permission sign-off                                                                               |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| POS                        | Terminal/PIN, catalog/cart, checkout/payments, orders/refunds/voids/splits, held orders, sessions/cash, customers/loyalty, floors/tables, kitchen, modifiers/combos, gift cards, receipts, stock, staff, offline | Owner and POS-only member terminal/PIN/cart; cart survives Portal navigation; settings load/retry | Full real-API payment/refund → stock/Finance/Insights chain; hardware/offline; all scoped staff actions                       |
| ERP                        | Products/PIM/variants, suppliers, purchasing/receiving, inventory/adjustments/transfers/takes, price lists, reports/settings                                                                                     | Product create/edit/delete, suppliers, purchase-order states                                      | Receiving/transfer → POS location stock; delete/variant/tax/currency parity; Finance supplier/bill linkage                    |
| CRM                        | Customers, merge/import/export, segments, loyalty, campaigns/templates, sales pipeline/deals/activities, communication history                                                                                   | Customer forms and loyalty screens                                                                | POS customer identity/loyalty round trip; campaign authorization and external-provider sandbox delivery                       |
| Finance                    | Invoices/payments, recurring invoices, quotes, expenses/bills, categories, accounts/journal, bank feeds/reconciliation, reports/tax                                                                              | Invoice and expense forms, bills, profit/loss, tax screens                                        | POS payments/refunds → journals/reports; supplier bill reconciliation; payroll posting; provider sandbox feeds                |
| HR                         | Staff/departments/reporting, shifts/roster, attendance/timesheets/approvals, leave, payroll, documents, reviews/one-on-ones, benefits/equipment, reports/settings                                                | Staff/detail, shifts, payroll view, leave balances, timesheet approval screen                     | Team user ↔ employee ↔ POS operator identity; payroll → Finance; approval scopes and sensitive fields                         |
| Flow                       | Tasks/projects/board/list/detail, inbox, today/triage/archive, views, cycles/milestones/calendar, time/presence/directory, settings                                                                              | Project/task/inbox/detail and access guards; central settings navigation                          | Real API task lifecycle across roles/scopes, attachments/reminders and time reporting; preview routes remain previews         |
| Insights                   | Overview, revenue, customers, products, operations, financial, reports/settings                                                                                                                                  | Dashboard render/error checks                                                                     | Reconcile totals and date/currency filters with POS, ERP, CRM and Finance source records                                      |
| Recruit                    | Jobs/careers, candidates/status, application/uploads/consent, admin permissions                                                                                                                                  | Owner creates job; candidate status refetch; viewer restrictions; public application/errors       | Real upload/storage and candidate → HR handoff; sensitive/export scopes and retention behavior                                |
| Menu                       | Published menu, product/image/price/availability, cart/checkout and order status, management                                                                                                                     | Public rendering and application screens; backend Menu operations in DB suites                    | Real Menu order → POS/kitchen/stock, offline/unavailable product races and modifier parity                                    |
| Portal / Caret / utilities | Home, team/groups/permissions, settings, business switching, Caret history/streaming/tools, Extensions, Automations                                                                                              | Home/launcher/member restrictions; Automations CRUD and navigation fixtures                       | Business switching with warm caches; live Caret providers/persistence; extension install/uninstall and provider configuration |

## Staff location versus floor assignment

`pos_member_locations` assigns a member to a POS location and stores management
capabilities. Portal Team invitation/member detail uses the POS access service.
Floors are seating layouts within a location. No separate employee-to-floor
assignment was found. Do not present location assignment as floor assignment.
A floor assignment feature would need an explicit contract, persistence and
server-side scope rules before a UI control is added.

## Permission operation coverage to extend

The catalog keys below are absent from the 104-record operation evidence ledger.
Their absence is an evidence gap, not proof of an authorization bypass. Add real
handler cases for allowed access, missing permission, wrong app, unlicensed or
revoked access, another tenant, and relevant location/project scopes.

- CRM: `campaigns:manage`, `segments:manage`, `templates:manage`.
- ERP: `price_lists:manage`, `price_lists:view`, `settings:manage`, `settings:view`,
  `suppliers:create`, `suppliers:edit`, `suppliers:view`.
- Finance: `accounts:manage`, `accounts:view`, `bank_feeds:manage`, `quotes:manage`,
  `quotes:view`, `recurring_invoices:manage`, `recurring_invoices:view`.
- Hub: `ai:quota:manage`, `billing:manage`, `billing:view`,
  `business:settings:manage`, `business:settings:view`.
- POS: `gift_cards:manage`.

The existing validator checks only the recorded operation set. Its success must
not be reported as coverage of every canonical permission. Cross-app refresh
checks must include warm caches, business switching, and separate sessions;
sharing a database does not automatically keep every UI synchronized.
