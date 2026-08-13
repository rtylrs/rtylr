# SQLC Query Ownership

SQLC packages are deliberately module-local. A byte-identical statement in two
packages is not automatically dead code: each package owns its repository API,
error mapping, cache invalidation, and (for write paths) transaction boundary.
Moving one module's generated method into another module would make a domain
service depend on another module's infrastructure package and can split an
otherwise atomic write into two connections.

This register records the disposition of the duplicate active query groups
audited in issue #203. A future change may share a query only through an
explicit application-facing provider interface; do not import another module's
`internal/infrastructure` package merely to reuse generated SQL.

Issue #203 describes 21 groups, but its enumerated duplicate-group list
contains the 20 groups below; this register covers every explicitly listed
group. Any newly discovered group must be added here with the same disposition
before its SQLC definition is moved or removed.

| Group | Canonical domain owner | Disposition |
| --- | --- | --- |
| Addon enable/disable (`addons`, `console`) | Addons | Intentional console administrative projection. Console keeps grant administration and its own audit/error boundary; the addon module owns product-facing enablement. |
| Bank balance sum (`bank_accounts`, `finance_reports`) | Bank accounts | Intentional finance-report projection. Reports compose this sum with accounts, receivables, and payables without depending on the bank-account repository. |
| Active seat/member count (`console`, `subscriptions`, `team`) | Subscriptions | Intentional read projections. Console renders administration state and team performs seat mutations in its transaction; neither may import subscription infrastructure. |
| Delete user (`console`, `user`) | User | Intentional console administrative capability. Console must preserve its administrator authorization and audit boundary; user owns self-service lifecycle. |
| Business owner lookup (`crm_sales`, `pos_stock_movements`, `user`) | Business | Intentional tenant-resolution reads. Each call participates in a different domain operation and maps missing/inactive businesses differently. |
| Customer count (`customers`, `kpis`) | Customers | Intentional KPI projection. The KPI aggregate reads several domains in one repository without coupling to customer infrastructure. |
| Active customer count (`customers`, `segments`) | Customers | Intentional segmentation projection. Segment recalculation owns its query/transaction boundary while customer owns CRUD. |
| Loyalty transaction count (`customers`, `loyalty`) | Loyalty | Intentional customer-detail projection. The customer module uses the count while merging customer records; loyalty owns the ledger API. |
| Active staff count (`kpis`, `staff`) | Staff | Intentional KPI projection. KPI composition must remain independent of staff infrastructure. |
| Low-stock product count (`kpis`, `products`) | Products | Intentional KPI projection. KPI composition keeps the exact inventory predicate local and does not expose product repository internals. |
| User ID by email (`oauth`, `workspaces`) | User | Intentional authentication/workspace projections. OAuth deliberately returns only an ID for credential flow; workspace needs its own user lookup shape. |
| Order-item status update (`pos_kds`, `pos_orders`) | POS orders | Intentional KDS command boundary. POS orders performs owner verification; KDS applies kitchen workflow changes in its own repository API. |
| Inventory transaction with reference (`purchase_orders`, `stock_takes`) | Inventory ledger | Intentional transactional writes. Receiving and stock-take completion insert the same ledger shape inside different multi-step transactions. |
| Login-history insert (`security`, `user`) | Security | Intentional auth-path write. Security owns login-history caching and session behavior; user records authentication events during its own flows. |
| Owner apps (`subscriptions`, `team`) | Subscriptions | Intentional team invite projection. Team resolves plan apps while provisioning a member and must retain its local failure/fallback behavior. |
| Subscription seat-count update (`subscriptions`, `team`) | Subscriptions | Intentional team transaction participant. Team updates the count together with membership and seat-period rows; subscriptions owns standalone billing changes. |
| Close active seat period (`subscriptions`, `team`) | Subscriptions | Intentional team transaction participant. Team closes and opens periods atomically with membership changes. |
| Open seat period (`subscriptions`, `team`) | Subscriptions | Intentional team transaction participant. Team creates the replacement period in the same transaction as the seat-count mutation. |
| User business ID (`team`, `user`) | User | Intentional tenant-resolution projection. Team maps missing users to team-domain errors; user owns profile/token behavior and caching. |
| Ensure employee for user (`team`, `user`) | User | Intentional onboarding transaction participant. Team provisioning and self-registration each retain their own error handling and transaction ownership. |

Campaign and segment SQLC duplication is deliberately excluded from this
register because it is tracked separately by issue #189.
