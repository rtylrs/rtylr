// Generates round-two issue bodies from the verification findings.
// Fails rather than emitting a body when Benefit, Why or ROI is absent:
// an issue without all three is not filed (brief, Phase 9).
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const OUT = join(process.cwd(), 'scripts', 'cross-app', 'round2-issue-bodies');

const issues = [
  // ---------------- onion ----------------
  {
    repo: 'onion',
    title:
      'Terminating an employee revokes nothing: add reconciliation job and status-driven revocation',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:hr_core', 'type:gap', 'severity:critical'],
    benefit:
      'The day after this ships, ending someones employment removes their app access automatically instead of leaving their login and every app entitlement active until a human remembers.',
    why: 'Gating and correctness (taxonomy 6) plus trigger and automation (4). Round one filed this as rtylr#834; round two re-verified the backend and it still holds: termination writes two status columns and stops (`internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:425`, `internal/modules/staff/sql/queries/staff.sql:26`), while no module anywhere reads a terminated status to revoke `user_app_permissions` or edit any apps array. The asymmetry is now sharper: team invite provisioning already writes across the boundary in the opposite direction, INSERTing an hr_employees row cross-schema (`internal/modules/team/sql/queries/team.sql:192-201`). Hire couples entitlement to employment; termination does not. This pairing over hub-side manual offboarding because the entitlement layer is what every surface reads.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | every tenant with staff turnover; reference tenant: ~8 departures/yr |\n| Manual work today | admin must remember to strip apps per leaver; missed cases are silent |\n| Modelled monthly value | 8 departures/yr x 15 min deprovisioning = 2h/mo reclaimed; one avoided incident of ex-employee access is far larger (unquantified) |\n| Build cost | 3 developer days (job + hook + tests) |\n| Payback | under 1 month |\n| Confidence | Inferred from code (counts modelled from the reference tenant) |\n| Cost of not doing it | unrevoked access persists indefinitely per missed case; security exposure with no compensating control |',
    exists:
      'Status write: employee_handlers.go:425 sets `terminated`; staff.sql:26 terminates hr_staff. No consumer of terminated status exists (`grep -rn terminated internal/modules` shows only writers). Entitlement storage: user_app_permissions (shared schema), per-user apps array served by hubprofile (`internal/platform/hubprofile/profile.go:14`). Provisioning precedent writing cross-schema: team.sql:192-201.',
    build:
      'Shape B4b then B3 of docs/cross-app/17-backend-plumbing.md (rtylr round-two branch). First the reconciliation job: a jobs entry finding users whose linked employee rows are terminated while permissions remain active, reporting before revoking. Then the synchronous half: on employee status change to terminated, deactivate matching user_app_permissions rows in the same request path via an explicit provider interface owned by team/user (not a direct module import). Idempotent, scoped by business_id, multi-business users lose access only to the terminating tenant.',
    coupling:
      'Dependency: team/user own entitlement writes; hr emits intent. Direction hr -> hub-side modules via interface; owner: hr_core. Tenant predicate: business_id on both employee lookup and permission rows; indexes exist (owner-leading on user_app_permissions). Degrade: if HR is disabled for a tenant, job finds nothing and changes nothing; behaviour identical to today.',
    links:
      'Dossier: tryrtylr/rtylr branch docs/cross-app-round2, docs/cross-app/05-connections/HR-HUB-01-revoke-access-on-termination.md. Frontend issue: rtylr#834. Related: onion#762 (token revocation on auth events).',
  },
  {
    repo: 'onion',
    title: 'Expose the live project labour-cost port through Flow projects workspaces RPC',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:workspaces', 'type:gap', 'severity:high'],
    benefit:
      'A project manager opens a Flow project and sees what its labour has cost so far, with rates never leaving HR.',
    why: 'Derivation (1) and rollup (7). Every input already exists and is joined server-side by the labourcost port landed in #743: SQL inside hr_payroll joins hr_time_entries to hr_employees.hourly_rate (`internal/modules/hr_payroll/sql/queries/labour_cost.sql`), workspaces consumes it through SetLabourCostProvider with a noop fallback (`internal/modules/workspaces/internal/application/service.go:113-129`, wired `cmd/api/main.go:466`). Nothing exposes it: no RPC calls Service.GetProjectLabourCost, so the capability answers no one. This pairing (expose through workspaces) beats a new hr.cost service because the port, permission posture and noop fallback all exist; only the wire binding is missing.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | services tenants running Flow + HR (reference tenant: 40-person, 6 projects) |\n| Manual work today | PMs export hours to a spreadsheet and multiply by remembered rates |\n| Modelled monthly value | 6 projects x 2 cost checks/mo x 20 min = 4h reclaimed per PM team; overruns caught weeks earlier |\n| Build cost | 1.5 developer days backend (RPC + contract binding), 2 days frontend panel |\n| Payback | about 1 month |\n| Confidence | Modelled from assumption on check frequency; code facts measured |',
    exists:
      'Port: internal/platform/labourcost/contract.go (ProjectCost: CostCents, CostedMinutes, UncostedMinutes, UncostedEmployeeCount). Implementation: hr_payroll repository.go:450. Consumer plumbing live but unconsumed. Contract drafted on voxie branch docs/cross-app-round2 (flow/project/v1 service + models). No .onk today: contracts.csv has no labour message.',
    build:
      'Bind the drafted getProjectLabourCost RPC on flow.project.v1 to Service.GetProjectLabourCost in the workspaces API layer; resolve business id via tenant.ResolveBusinessID; keep billable-only scope documented. Permission-gate on a Flow project-level visibility rule following presence.gos reduced-projection reasoning. Frontend panel consumes with rate-coverage labelling.',
    coupling:
      'Direction flow -> hr via existing port; owner hr_payroll. Tenant predicate business_id on both sides of the join (labour_cost.sql WHERE e.business_id = ? AND emp.business_id = ?); serves off idx_time_entries_owner-style keys plus project filter, verify against docs/tenant-index-matrix.md and add business-leading index if EXPLAIN disagrees. No transaction. Degrade: noop fallback returns zero-value struct; UI must render unknown, not zero.',
    links:
      'Dossier: docs/cross-app/05-connections/HR-FLOW-01-project-labour-cost.md (round-two backend section). Issues: rtylr#837, rtylr#862 partially. Related onion PR #743.',
  },
  {
    repo: 'onion',
    title: 'Nothing marks time entries invoiced: give invoices a port to write billable_invoice_id',
    labels: [
      'audit:cross-app-round2-2026-08-22',
      'area:invoices',
      'area:hr_time',
      'type:gap',
      'severity:high',
    ],
    benefit:
      'Once hours are billed, they stop appearing as unbilled, so Finance can trust the number instead of re-checking old invoices.',
    why: 'Single source of truth (3) and gating correctness (6). The unbilled-billable query filters `AND billable_invoice_id IS NULL` (`internal/modules/hr_time/sql/queries/time_entries.sql:52`) and the column exists (`schemas/hr_time/schema.sql:22`), but a repo-wide sweep finds no writer at any layer: the predicate is permanently true and the unbilled number can never decrease. Round one proposed the filters (landed as voxie#46/onion#783); this is the half neither pass saw. Pairing invoices -> hr_time rather than a generic export keeps ownership: hr_time owns its table, invoices call an explicit interface.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | any tenant billing time (reference tenant bills ~120 invoices/mo) |\n| Manual work today | billing staff cross-check exported batches against issued invoices by hand |\n| Modelled monthly value | 4h/mo reconciliation avoided; prevents double-billing incidents (each costly) |\n| Build cost | 1 day backend (interface + write path), included frontend effect |\n| Payback | under 1 month |\n| Confidence | Inferred from code; frequency modelled |',
    exists:
      'Column and read filter verified; write path absent everywhere including the invoices and finance_reports modules. Domain struct carries BillableInvoiceID (`hr_time/internal/domain/types.go:24`). The export flow already demonstrates the claim-and-release idempotency pattern to copy (`hr_payroll/internal/application/service.go:377-409`).',
    build:
      'Application-facing provider interface (shape B3, 17-backend-plumbing.md): hr_time exposes MarkEntriesInvoiced(ctx, businessID, ids, invoiceID); invoices module receives it at construction like insights receives FlowReader (`main.go:464` precedent). Called in the same transaction as invoice creation or with claim/release if deferred. Add query group to docs/sqlc-query-ownership.md register.',
    coupling:
      'Direction invoices -> hr_time via interface; owner hr_time. Tenant predicate business_id in UPDATE WHERE. Index: existing pk-based update suffices. Transaction: single-module atomic write. Degrade: invoices module works unchanged when hr_time disabled for tenant (provider nil-guarded); unbilled counts simply stay as-is.',
    links:
      'Dossier: docs/cross-app/05-connections/FLOW-FIN-01-unbilled-billable-hours.md round-two section; gap A1/A2 in docs/cross-app/14-layer-reconciliation.md. Issues rtylr#836, rtylr#855 (filters landed).',
  },
  {
    repo: 'onion',
    title: 'Webhook registrations have no delivery engine: registered URLs never receive events',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:webhooks', 'type:gap', 'severity:medium'],
    benefit:
      'Tenants who register a webhook start receiving events, instead of configuring integrations that silently never fire.',
    why: 'Trigger and automation (4). The webhooks module is CRUD over registrations exposed through the hubdeveloper port (`main.go:732`), but a repo-wide sweep finds exactly one outbound POST site, Paddles client talking to Paddle (`internal/platform/paddle/client.go`); nothing delivers tenant webhook events (`backend_events.csv`, emission probe). Dash sells a developer surface whose subscriptions are inert. Filed in onion because filling the gap is backend platform work; it also gates any future shape-3 design that assumed webhooks exist.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | those using api_keys/webhooks developer features |\n| Manual work today | tenants poll or hand-sync; or do not integrate at all |\n| Modelled monthly value | unquantified enablement; small now, grows with API adoption |\n| Build cost | 5 to 8 developer days (worker-driven delivery, retry/backoff, signing, dedup) |\n| Payback | strategic; negative short-term on direct ROI alone |\n| Confidence | Modelled from assumption |',
    exists:
      'Registration CRUD: webhooks/internal/application/service.go:46-100. Provider port: platform/hubdeveloper/developer.go:34. Delivery: absent (search across repo). Note onion#755 covers Paddle inbound security, not this outbound engine.',
    build:
      'Worker-driven dispatcher reading due registrations, POSTing signed payloads with retries and dead-lettering, per-event-type filters; reuse emailqueue patterns (workerqueue). Decide event vocabulary after HR->HUB-01 and FIN->ERP-02 needs are known; do not invent events speculatively.',
    coupling:
      'Platform capability, no new domain edge until events exist. Owner: webhooks module + worker. Tenant predicate business_id scoping throughout. Degrade: absence is todays behaviour; additive.',
    links:
      'Docs: docs/cross-app/16-backend-matrix.md async backbone; 17-backend-plumbing.md shape B4 option 3.',
  },
  {
    repo: 'onion',
    title: 'Memory module queries carry no tenant predicate: isolation unverified on an AI surface',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:memory', 'type:gap', 'severity:high'],
    benefit:
      'Operators get proof that one businesss AI memory can never surface in another businesss assistant session.',
    why: 'Gating and correctness (6); tenancy is part of every backend edge. Of memorys sqlc queries, none references business_id or owner_id (`backend_queries.csv`: 18 of 18 queries without either; handler tokens likewise none-found-in-module). Spot-check of the API layer found tenancy tokens nowhere in `internal/modules/memory/internal/api/memory_service.go`. This may be safe if memory is keyed purely by user-scoped session ids, but nothing states that invariant, no test asserts it, and echo feeds it into AI contexts where a leak would be invisible. Filed for owners to confirm-or-fix rather than asserted as a leak: static analysis cannot see key derivation.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | all tenants using echo/Caret surfaces |\n| Manual work today | none possible: nobody can audit what cannot be expressed |\n| Modelled monthly value | risk removal; one cross-tenant memory leak would be a reportable incident |\n| Build cost | 0.5 days to prove isolation and pin it with a test; 2 days if a scope column must be added and backfilled |\n| Payback | immediate (risk) |\n| Confidence | Inferred from code; severity contingent on what the sweep could not see |',
    exists:
      'Queries: memory sqlc set, zero tenant tokens (backend_queries.csv). Handler: none-found-in-module. Consumers: echo wired with memoryModule twice in main.go:547.',
    build:
      'Either document the isolation invariant (session/user-keyed only) with a citation and add a test asserting cross-user unreadability, or add business_id scoping mirroring the v8 tables convention (NOT NULL DEFAULT-free) plus backfill. Either way record the disposition here and in docs/sqlc-query-ownership.md style notes.',
    coupling:
      'Intra-module hardening; no cross-domain edge. Owner: memory module maintainers. Index impact: adding business_id would want a leading composite; defer decision to fix path.',
    links: 'Coverage note: docs/cross-app/13-backend-inventory.md tenancy section.',
  },
  {
    repo: 'onion',
    title:
      'Three writers maintain crm_customers outside the customers module: consolidate through customer_sync',
    labels: [
      'audit:cross-app-round2-08-22'.replace('08', '2026-08'),
      'area:customers',
      'type:gap',
      'severity:medium',
    ],
    benefit:
      'Customer records behave the same no matter which surface created or updated them, so renames and merges stop splitting history.',
    why: 'Single source of truth (3). backend_tables.csv shows crm_customers written by customers, store_orders and platform/customer_sync, and read by eight modules. The sanctioned seam for external writers is customer_sync (POS/storefront upserts with contact matching, `platform/customer_sync/customer_sync.go:18`); store_orders bypasses it with a direct write. Two independent writer paths means validation, name-splitting and visit-count rules drift apart. Consolidation belongs in onion where the writes live.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | retail tenants using storefront ordering alongside POS/CRM |\n| Manual work today | duplicate customer cleanup by staff |\n| Modelled monthly value | 30 min/mo cleanup per retail tenant (modelled) |\n| Build cost | 1 day rerouting writes through customer_sync + tests |\n| Payback | about 1 month for retail tenants |\n| Confidence | Inferred from code |',
    exists:
      'Writers: customers (module CRUD), platform/customer_sync UpsertCRMCustomer, store_orders direct INSERT/UPDATE (backend_tables.csv crm_customers row). Reads: insights, kpis, loyalty, segments, feedback, pos side.',
    build:
      'Route store_orders customer persistence through customer_sync.UpsertCRMCustomer; keep customers module as canonical CRUD; add the query group to the ownership register. Verify same-business validation parity between paths first.',
    coupling:
      'Direction storefront-side module -> crm via platform seam; owner: customers module defines invariants, customer_sync implements matching. Tenant predicate business_id present in sync path; verify store_orders current path. Degrade: storefront order creation must not fail when CRM disabled; fall back to order-local guest record.',
    links:
      'Matrix: docs/cross-app/16-backend-matrix.md watchlist item 3. Context: packages/voxie/api/pos/customer/v1/models.onk collision note.',
  },
  {
    repo: 'onion',
    title: 'Serve the leave-request date range and pagination filters once voxie lands them',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:workspaces', 'type:gap', 'severity:medium'],
    benefit:
      'Planning screens can ask who is away between two dates, so work stops being assigned to people on holiday.',
    why: 'Context in place (5) and gating (6). ListLeaveRequestsRequest filters by employee and status only; asking overlap requires fetching all leave and filtering client-side, which no consumer should do. Drafted in voxie (branch docs/cross-app-round2, hr/leave/v1/models.onk: from_date/to_date with documented overlap semantics plus limit/offset). Backend follow-up lands here. Pairs with rtylr#838 (HR->FLOW-02) whose warning feature needs exactly this query.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | all Flow+HR tenants |\n| Manual work today | PMs check the leave calendar by memory before assigning |\n| Modelled monthly value | avoids ~2 misassignments/mo x 45 min rework = 1.5h (modelled) |\n| Build cost | 0.5 day backend after contract lands |\n| Payback | under 1 month |\n| Confidence | Inferred from code |',
    exists:
      'Contract: drafted on voxie branch. Storage: hr_leave_requests has start_date/end_date/status (schemas/hr/schema.sql:24 region). Query: listLeaveRequests lacks range predicates today. Tenancy note: hr_leave ownerID semantics differ from other HR modules (presence.go comment); pass caller-scoped ids accordingly.',
    build:
      'Add range predicates with the documented overlap semantics (start <= to AND end >= from), limit/offset pagination, index check against hr_leave leading columns; expose via existing RPC; unit test the boundary-overlap case named in the draft comment.',
    coupling:
      'Consumer flow reads hr via existing listLeaveRequests; owner hr_leave. Predicate: caller identity per module convention, not business_id blindly. Degrade: consumers hide the panel when HR off; empty set otherwise.',
    links:
      'Draft: voxie branch docs/cross-app-round2 api/hr/leave/v1/models.onk. Issues: rtylr#856 (contract), rtylr#838 (feature).',
  },

  // ---------------- voxie ----------------
  {
    repo: 'voxie',
    title: 'Contract: add date range and pagination to hr leave requests',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:hr', 'type:gap', 'severity:medium'],
    benefit: 'Callers can ask who is on approved leave between two dates in one paginated request.',
    why: 'Serves HR->FLOW-02 (leave gates assignment). ListLeaveRequestsRequest filters by employee_id and status only; overlap questions require unbounded client-side filtering. Cross-links: rtylr#856 (misfiled round-one contract issue), rtylr#838 (consumer feature).',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | Flow+HR tenants |\n| Manual work today | PMs consult leave calendar manually |\n| Modelled monthly value | ~1.5h avoided rework (modelled) |\n| Build cost | contract 0.25d; backend 0.5d; frontend part of #838 |\n| Payback | < 1 month after landing |\n| Confidence | Inferred from code |',
    exists:
      'Drafted on branch docs/cross-app-round2, api/hr/leave/v1/models.onk: from_date/to_date @nullable queries with documented overlap semantics; limit/offset added matching hr house style.',
    build:
      'Land contract here; onion serves (paired issue); packages pointer bump; rtylr consumes. Order: voxie, packages, onion, rtylr.',
    coupling:
      'No new domain edge; refines existing hr->flow read. Filters optional: absent means include-all, preserving current callers.',
  },
  {
    repo: 'voxie',
    title: 'Contract: give finance invoices a nullable CRM customer_id with snapshot semantics',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:finance', 'type:gap', 'severity:high'],
    benefit:
      'Revenue per customer becomes answerable: each invoice knows which customer it was issued to.',
    why: 'Identity resolution (8), single source of truth (3). Invoices store four free-text client_* strings only; nothing joins them to crm_customers, so a renamed or re-typed customer splits revenue history irrecoverably. Highest-leverage change product-wide (roadmap). Cross-links: rtylr#857, rtylr#840 (consumer), roadmap docs/cross-app/08-roadmap.md.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | every finance tenant |\n| Manual work today | spreadsheet matching of invoice clients to customers |\n| Modelled monthly value | 6h/mo reporting prep (reference tenant); enables credit decisions (#841) |\n| Build cost | contract 0.25d; onion column+validator+snapshot 2d; form picker 1.5d |\n| Payback | ~1 month |\n| Confidence | Inferred from code; value modelled |\n| Cost of not doing it | unmatched history accrues daily and no migration can repair it later |',
    exists:
      'Drafted on branch: InvoiceDetail.customer_id, CreateInvoiceRequest.customer_id (client_name/client_email relaxed to conditional), ListInvoicesRequest.customer_id filter; snapshot-preservation documented on each field.',
    build:
      'Voxie first (this PR); onion adds nullable column, same-business validation, server-side client_* snapshot when customer_id supplied; rtylr finance form gains customer picker. Never regenerate gen/ in the drafting PR.',
    coupling:
      'New declared edge crm -> finance (identity flows one way; balances return later via FIN->CRM-01). Owner: finance owns invoice, crm owns customer. Nullable: tenants without CRM behave identically.',
  },
  {
    repo: 'voxie',
    title: 'Contract: link pos menu items to erp products with unit_cost snapshot',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:pos', 'type:gap', 'severity:high'],
    benefit:
      'Menu items know what they cost, making per-item margin computable for the first time.',
    why: 'Derivation (1), single source of truth (3). MenuItem has stock and sale_price but no product reference and no cost; ERP holds both. Two stock numbers for one physical thing can disagree undetectably today. Cross-links: rtylr#858, rtylr#843.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | retail/restaurant tenants running ERP+POS |\n| Manual work today | margin computed offline from supplier sheets |\n| Modelled monthly value | pricing decisions enabled; shrinkage visible (unquantified) |\n| Build cost | contract 0.25d; onion link+validation 1d; pos UI 2d (other repo) |\n| Payback | 2 months (retail segment) |\n| Confidence | Inferred from code |',
    exists:
      'Drafted on branch: MenuItem.product_id @nullable, MenuItem.unit_cost @nullable with authoritative-value documentation. ERP->INS-01 margin work builds on this.',
    build:
      'Voxie; then onion menu module validates same-business product link and snapshots unit_cost on read; pos repo consumes. Stock-authority unification stays a separate ERP/POS decision, deliberately out of this draft.',
    coupling:
      'Declared edge erp -> pos. Nullable: non-stocked items unaffected. Server must reject foreign-business product ids.',
  },
  {
    repo: 'voxie',
    title:
      'Contract: make bill supplier_id nullable, require supplier_name snapshot, add purchase_order_id',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:finance', 'type:gap', 'severity:high'],
    benefit:
      'Bills record real suppliers instead of fabricated ids, and can reference the purchase orders they settle.',
    why: 'Correctness (6), double-entry elimination (2). supplier_id is @required @gt(0) yet the backend never joins erp_suppliers and the frontend slugifies typed names into the field (ERP->FIN-01). purchase_order_id enables settlement closing (FIN->ERP-02). Serves drafts 5+6 combined into one coherent bill change. Cross-links: rtylr#860, rtylr#859, rtylr#835, rtylr#851.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | every finance tenant |\n| Manual work today | vendor names retyped per bill; PO matching manual |\n| Modelled monthly value | 3h/mo AP clerk time (reference tenant); duplicate-payment risk reduced |\n| Build cost | contract 0.25d; onion 1.5d; finance UI 2d |\n| Payback | < 1 month |\n| Confidence | Inferred from code |\n| Cost of not doing it | fabricated foreign keys accumulate; payables reporting unreliable |',
    exists:
      'Drafted on branch: Bill.supplier_id int64? @nullable; supplier_name string @required (snapshot); Bill.purchase_order_id + CreateBillRequest fields; ListBillsRequest.supplier_id/purchase_order_id filters.',
    build:
      'Voxie; onion relaxes validator, stores snapshot, verifies foreign ids same-business; finance ships supplier picker (wave-1 item 2) before PO-settlement UI.',
    coupling:
      'Declared edges erp -> finance (identity) and finance -> erp (settlement reference, read by erp). Fact graph stays acyclic: settlement reference does not consume derived data.',
  },
  {
    repo: 'voxie',
    title: 'Contract: reference the CRM customer from support customer',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:support', 'type:gap', 'severity:medium'],
    benefit:
      'Support history attaches to real customers, so account views show ticket health without email guesswork.',
    why: 'Identity resolution (8), context in place (5). support_customers is an unrelated table joining to crm_customers only by loose email strings. Explicit nullable reference with confident-match-or-human-confirm semantics beats automatic merging. Cross-links: rtylr#861, rtylr#842.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | support users; operator surface still absent in rtylr |\n| Manual work today | agents search CRM by email |\n| Modelled monthly value | low today; rises when a support app ships |\n| Build cost | contract 0.1d; onion column+match UX 1d |\n| Payback | contingent on support surface shipping |\n| Confidence | Modelled from assumption |',
    exists:
      'Drafted on branch: Customer.customer_id int64? @nullable with match-policy doc comment.',
    build:
      'Voxie; onion adds column + set-on-confident-match logic; consumer lands with SUP->CRM-01. Keep email matching as suggestion, never auto-write.',
    coupling: 'Declared edge support -> crm. Nullable preserves current behaviour entirely.',
  },
  {
    repo: 'voxie',
    title: 'Contract: bind the live project labour-cost aggregate to flow.project.v1',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:flow', 'type:gap', 'severity:high'],
    benefit:
      'Flow projects can display what their labour has cost, computed server-side without exposing any pay rate.',
    why: 'Derivation (1), rollup (7). The aggregate exists end-to-end behind the labourcost port (#743) but no RPC binds it, so no consumer can reach it. Contract follows the live implementation field-for-field (billable-only scope, uncosted coverage counts) rather than inventing a parallel shape. Cross-links: rtylr#862 (partially superseded), rtylr#837, onion paired exposure issue.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | Flow+HR services tenants |\n| Manual work today | spreadsheet cost tracking per project |\n| Modelled monthly value | 4h/mo PM time; earlier overrun detection |\n| Build cost | contract 0.25d; onion RPC 1d; flow panel 2d |\n| Payback | ~1 month |\n| Confidence | Inferred from code |',
    exists:
      'Drafted on branch: GetProjectLabourCost RPC on ProjectService (@get /projects/{id}/labour-cost); ProjectLabourCost message mirrors port fields with partial-labelling doc comments. HR->INS-01 may consume the same RPC.',
    build:
      'Voxie; onion binds RPC in workspaces API over the port (permission-gated, tenant-resolved); rtylr renders with partial labels when uncosted_minutes > 0.',
    coupling:
      'Edge flow -> hr already exists via port; contract formalises it. Aggregate-only output preserves salary boundary (staff contract precedent).',
  },
  {
    repo: 'voxie',
    title: 'Contract: add per-product margin rows to the existing insights margins response',
    labels: ['audit:cross-app-round2-2026-08-22', 'area:insights', 'type:gap', 'severity:medium'],
    benefit:
      'Margin per product becomes visible next to the existing margin distribution, without pulling the ERP catalogue.',
    why: 'Rollup (7), derivation (1). getProductMargins exists and is implemented but returns distribution buckets only; ERP->INS-01 needs per-product rows carrying cost. Narrow additive extension beats round-ones standalone ProductMargin proposal since Insights already reads erp_products server-side. Cross-links: rtylr#853 (superseded in part), rtylr#853 consumer.',
    roi: '| Item | Value |\n| --- | --- |\n| Affected tenants | ERP+Insights tenants |\n| Manual work today | export join between sales report and price list |\n| Modelled monthly value | 2h/mo analyst time (retail reference) |\n| Build cost | contract 0.1d; onion query 0.75d; chart 1d |\n| Payback | ~2 months (retail segment) |\n| Confidence | Inferred from code |',
    exists:
      'Drafted on branch: ProductMargin message (cost_minor_units nullable when no cost on record) plus products field on GetProductMarginsResponse.',
    build:
      'Voxie; onion extends GetMarginDistribution path with per-product query (own query group, register entry); dash/insights charts consume.',
    coupling:
      'Edge erp -> insights already flows via raw table reads (undeclared today); this contract declares the projection formally. Nullable cost distinguishes unknown from zero.',
  },
];

mkdirSync(OUT, { recursive: true });
const manifest = [];
for (const iss of issues) {
  for (const [key, sec] of [
    ['benefit', 'Benefit'],
    ['why', 'Why'],
    ['roi', 'ROI'],
    ['exists', 'What exists, what is missing'],
    ['build', 'How to build it'],
    ['coupling', 'Coupling cost'],
  ]) {
    if (!iss[key] || !iss[key].trim()) {
      console.error(`FAIL: ${iss.title} missing ${sec}`);
      process.exit(1);
    }
  }
  const body = [
    '## Benefit',
    iss.benefit,
    '',
    '## Why',
    iss.why,
    '',
    '## ROI',
    iss.roi,
    '',
    '## What exists, what is missing',
    iss.exists,
    '',
    '## How to build it',
    iss.build,
    '',
    '## Coupling cost',
    iss.coupling,
    '',
    iss.links,
    '',
    `Labelled ${iss.labels.join(', ')}.`,
  ].join('\n');
  const file = join(
    OUT,
    `${iss.repo}-${iss.title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .slice(0, 48)}.md`
  );
  writeFileSync(file, body);
  manifest.push({ repo: iss.repo, title: iss.title, labels: iss.labels, file });
}
writeFileSync(join(OUT, 'manifest.json'), JSON.stringify(manifest, null, 2));
console.log(`wrote ${manifest.length} bodies to ${OUT}`);
