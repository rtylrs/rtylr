# Persistence ownership

The first directory under `internal/modules` is a bounded context. Child packages are capabilities of that context, not additional owners of its tables. SQL sources, sqlc configuration, generated queries, repositories and transactions live with that owner.

| Context | Mutable state |
| --- | --- |
| Organization | Accounts, businesses, membership, authorization assignments, subscriptions, entitlements and settings |
| ERP | Products, categories, procurement, stock locations, stock levels, stock ledger, online product listings and order fulfillment |
| POS | Registers, floors, tables, orders, payments, refunds, terminals, register users and kitchen workflow |
| CRM | Customers, contacts, leads, loyalty, campaigns and segmentation |
| Finance | Accounts, journals, invoices, payments, expenses, bills and reconciliation |
| HR | Employees, organization structure, rosters, attendance, leave and payroll |
| Flow | Projects, tasks, workspaces, comments and planning |
| Booking | Services, resources, appointments, reminders and calendar concurrency |
| Store | Storefronts, checkout sessions, shipping, payment configuration and themes |
| Insights | Report settings and consumer projections |
| Caret | Chats, projects and conversation memory |

Other contexts and platform state retain their named schema owners. The complete table inventory is computed from the checked-in SQL sources by `scripts/cross-app/sweep-tables.mjs`.

## Integration boundaries

A consumer defines the small interface or callback it needs. Wiring belongs in the composition root; field mapping belongs in an adapter owned by the consumer. A producer exposes application operations, values and explicit errors. Consumers cannot import its repositories, generated SQL types or internal packages.

A synchronous owner read returns current data or an error. For example, organization reporting preferences return `sql.ErrNoRows` for an unknown or inactive business. ERP sales quotes expose only active products within the specified business and exclude costs and supplier information. Consumers persist the price they observed; a quote does not lock another context's product transaction.

A projection belongs to its consumer and is refreshed from published facts or owner APIs. A direct SQL join to an owner's table is not a projection. Published facts describe completed source operations; they cannot instruct another module to invalidate its cache.

Insights reports read through their owners. Finance reports, HR reports, ERP products, POS orders and CRM customers serve `Insights*` reads over their own tables, with result types in `internal/contracts/<owner>`. Insights combines the results itself: product reports join POS sales to ERP product details, and the top-customer list adds CRM contact details to POS purchase totals. Customer purchase figures count the customer recorded on each paid order, so a customer later deleted from CRM still counts for the periods they bought in. Cached reports are versioned by the latest change any owner reports.

POS commits the paid order and its durable fact in one POS transaction. ERP processes that fact in a separate ERP transaction, aggregates repeated products and deduplicates ledger movements. Delivery can be retried. Inventory visibility follows event delivery, so it is not synchronous with POS payment.

Online order confirmation belongs to ERP fulfillment because the status and stock availability form one synchronous invariant. It retains the existing Store routes. Storefront configuration and checkout remain Store capabilities; checkout reads active product quotes through the ERP boundary.

CRM owns the POS contact lookup route because it reads the same customer records and exposes only contact fields. There is no separate POS customer repository or domain copy. CRM mutations invalidate that lookup alongside its other customer caches.

## Transactions and migrations

Transactions may cover multiple capabilities of one owner. Never pass a SQL transaction to another bounded context. Booking schedule locks and POS order sequences belong to their respective contexts rather than using organization rows as lock targets.

`internal/platform/database/schema_registry.tsv` maps stable migration identities to their owned SQL sources. Moving a source does not change its applied identity. A historical migration may concatenate sources from different owners when splitting ownership of an existing schema; subsequent changes belong to the relevant owner. Fresh database replay must produce the intended schema, and production migrations must be applied separately from the service rollout.

`make architecture-check` inspects source boundaries, contracts, transaction calls, context cycles, table access and the migration registry. The only accepted deviations are exact, dated entries in `docs/architecture-exceptions.json` with an owner, reason, containment and exit plan. No blanket legacy exception applies.
