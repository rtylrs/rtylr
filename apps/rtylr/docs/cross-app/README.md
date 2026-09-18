# Cross-app data sharing audit

An audit of every piece of data that is shared, could be shared, or must be
shared between RTYLR's ten apps and twelve backend contract domains, and a
design for connecting the ones worth connecting.

Produced 2026-08-21 against `origin/main` at `ab006407`. Static analysis only:
**no dev server was started and no request was issued**, because the apps default
to the production API. Every runtime claim is a contract-and-code claim, and
`BLOCKED.md` says what that costs.

## The short version

|                                           |                                            |
| ----------------------------------------- | ------------------------------------------ |
| Contract records swept                    | 7,388 across 299 `.onk` files, 12 domains  |
| Frontend requests swept                   | 743 call sites across 4 client idioms      |
| Action handlers swept                     | 1,241                                      |
| Entities identified                       | 178, in 11 domains                         |
| Backend modules mapped                    | 133                                        |
| Ordered domain pairs given a verdict      | **132 of 132**                             |
| Candidates considered                     | 40                                         |
| **Proposals**                             | **21**, each with a dossier                |
| **Rejected**                              | **18**, each naming the bar rule it failed |
| Already working, kept as precedent        | 1                                          |
| New domain dependencies the proposals add | 14                                         |
| Total modelled effort                     | ~84 to 98 developer days                   |

## The five findings to act on first

1. **Terminating an employee revokes no access anywhere.** It sets one column
   (`~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:426`)
   and stops. The ex-employee keeps their login and every app entitlement until a
   human remembers to remove them by hand. This is the only security exposure in
   the audit. `HR->HUB-01`.
2. **Finance fabricates a foreign key.** `finance.Bill.supplier_id` is
   `@required @gt(0)`, the backend never joins `erp_suppliers`, and the create
   form manufactures the id by slugifying the typed vendor name
   (`apps/finance/src/features/bills/Bills.tsx:59`). `ERP->FIN-01`.
3. **An invoice does not know its customer.** It stores four free-text
   `client_*` strings (`packages/voxie/api/finance/invoices/v1/models.onk:30`).
   Revenue per customer is unanswerable, and every invoice issued in the
   meantime adds history no later migration can reliably match. The single
   highest-leverage change in the product. `CRM->FIN-02`.
4. **HR stores a person twice, with two pay rates.** `hr_staff` and
   `hr_employees` both carry `salary`, `hourly_rate` and `overtime_rate`, bridged
   by a nullable column. Leave keys on one id, payroll on the other. There are
   also two parallel leave systems, one of which hardcodes its leave types.
   `HR->HR-01`.
5. **A cross-domain field that exists and nothing reads.**
   `flow.WorkType.hr_category_id` is documented as mapping Flow work types onto
   HR payroll categories "so one entry serves both project reporting and
   payroll". It is persisted, echoed, and read by nothing
   (`packages/voxie/api/flow/time/v1/models.onk:20`). This is the base rate for
   any proposal that adds a contract field ahead of its consumer, and it shaped
   every recommendation here.

## How to read this

Start with whichever question you have:

| Question                                                           | Document                                                     |
| ------------------------------------------------------------------ | ------------------------------------------------------------ |
| What was swept, how completely, and what the sweep could not reach | [01-inventory.md](01-inventory.md)                           |
| Who owns which noun, and can two records be joined                 | [02-entity-graph.md](02-entity-graph.md)                     |
| Does anything flow between domain A and domain B                   | [03-sharing-matrix.md](03-sharing-matrix.md)                 |
| Is this blocked by a missing field, endpoint, or concept           | [04-gaps.md](04-gaps.md)                                     |
| What exactly should we build, and why that pairing                 | [05-connections/README.md](05-connections/README.md)         |
| What contract changes would be needed                              | [06-proposed-contracts.md](06-proposed-contracts.md)         |
| How should a cross-app connection be built in this codebase        | [07-plumbing.md](07-plumbing.md)                             |
| What should we do first                                            | [08-roadmap.md](08-roadmap.md)                               |
| Which issues were filed                                            | [09-issues.md](09-issues.md)                                 |
| Where do the docs disagree with the code                           | [10-doc-drift.md](10-doc-drift.md)                           |
| What was considered and deliberately not proposed                  | [11-rejected.md](11-rejected.md)                             |
| What dependencies would this create, and is the graph acyclic      | [12-coupling.md](12-coupling.md)                             |
| What repositories exist and which this audit covered               | [00-repos.md](00-repos.md)                                   |
| The backend swept as a system (round two)                          | [13-backend-inventory.md](13-backend-inventory.md)           |
| Where schema, contract, handler and browser disagree (round two)   | [14-layer-reconciliation.md](14-layer-reconciliation.md)     |
| Round-one claims re-verified against Go and SQL (round two)        | [15-round-one-verification.md](15-round-one-verification.md) |
| What already flows between backend modules today (round two)       | [16-backend-matrix.md](16-backend-matrix.md)                 |
| Backend plumbing shapes for cross-module work (round two)          | [17-backend-plumbing.md](17-backend-plumbing.md)             |
| Which forks were taken during the audit                            | [DECISIONS.md](DECISIONS.md)                                 |
| What could not be finished                                         | [BLOCKED.md](BLOCKED.md)                                     |
| Flow specifically, in more depth than the matrix allows            | [flow-deep-dive.md](flow-deep-dive.md)                       |

## Reproducing it

Everything mechanical is a committed script. The CSVs under `inventory/` are
generated, not hand-written.

```bash
# The five inventory artifacts. Order matters: later scripts read earlier CSVs.
node scripts/cross-app/sweep-contracts.mjs
node scripts/cross-app/sweep-endpoints.mjs
node scripts/cross-app/sweep-entities.mjs
node scripts/cross-app/sweep-actions.mjs
node scripts/cross-app/sweep-backend-modules.mjs   # needs ~/Development/onion

# Documents rendered from docs/cross-app/connections.json.
node scripts/cross-app/render-matrix.mjs           # writes into 03-sharing-matrix.md
node scripts/cross-app/render-connection-index.mjs # writes 05-connections/README.md and 11-rejected.md

# The two assertions. Both exit non-zero on failure.
node scripts/cross-app/check-acyclic.mjs           # dependency graph is acyclic
node scripts/cross-app/verify-citations.mjs        # every file:line citation resolves
```

`connections.json` is the single source of truth for the connection set. The
matrix, the dossier index, the rejected list and the dependency graph are all
derived from it, so they cannot drift apart.

## Evidence standard

- Every factual claim carries a `file:line` citation, and a committed script
  checks that all 467 of them resolve.
- The citation checker verifies a path exists and a line is in range. **It does
  not verify the line says what the prose claims**, so cited lines were also
  spot-read by hand while writing.
- "The endpoint exists" and "the field exists" are treated as different claims
  throughout, because the difference produced three of the five headline
  findings.
- Anything not verified is marked **"inferred, not verified"** with what would
  verify it.
- Every ROI figure is **modelled**, against the reference tenant defined in
  `08-roadmap.md`, with the arithmetic shown. Nothing in this audit is measured.
- Coverage is stated rather than rounded up: `01-inventory.md` lists the 13
  dynamic paths, the excluded test files and the 67 unmatched request paths.

## What this audit did not do

- **No code was changed.** No app, package, contract or backend file was edited.
  Proposed `.onk` is drafted inside `06-proposed-contracts.md` as fenced blocks;
  nothing under `packages/voxie` or `~/Development/voxie` was touched.
- **No requests were made**, so no production data was read or written.
- **No connection was built.** This is a design document set.
