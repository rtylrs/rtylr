# 03. The sharing matrix

Every ordered pair of backend contract domains, with a verdict. The matrix is
generated from `connections.json` by `scripts/cross-app/render-matrix.mjs`, so
no pair can be quietly skipped: the script fails if any cell is unexplained or
any connection goes unrendered.

The matrix is by **backend domain, not by app** (`DECISIONS.md` #5). Data
ownership lives in domains; apps are windows onto them. The app-to-domain
mismatch, including the two domains with no app and the two apps with no
domain, is reported in `01-inventory.md`.

**12 x 12 = 144 cells. 12 are self-pairs. 132 ordered cross-domain pairs each
have a verdict.** 28 cells carry at least one connection id. The rest carry a
reason code, and the reason codes are not a shrug: `no-producer-data` means the
producing domain's entity list in `02-entity-graph.md` was checked and holds
nothing the consumer needs.

## Reading a cell

- A bare id (`ERP->FIN-01`) is a **proposal**, and has a dossier in
  `05-connections/`.
- `(rejected)` means the candidate was considered and failed the bar in
  section 7 of the brief. The rule it failed and the reasoning are in
  `11-rejected.md`. **18 of 40 candidates were rejected.**
- `(existing)` means it already works and is documented as precedent.
- An italic code is a pair with no connection, and the reason.

<!-- BEGIN GENERATED MATRIX -->

Rendered from `connections.json`: 12x12 = 144 ordered pairs, every one accounted for.

| producer \ consumer | booking                | common  | console              | crm                     | erp                         | finance                                     | flow                        | hr                                                | hub                    | insights                | pos                                    | support                |
| ------------------- | ---------------------- | ------- | -------------------- | ----------------------- | --------------------------- | ------------------------------------------- | --------------------------- | ------------------------------------------------- | ---------------------- | ----------------------- | -------------------------------------- | ---------------------- |
| **booking**         | -                      | _infra_ | _console_            | BOOK->CRM-01 (rejected) | _no-producer-data_          | _no-producer-data_                          | _no-producer-data_          | _would-invert_                                    | _no-producer-data_     | _no-producer-data_      | _no-consumer-surface_                  | _no-consumer-surface_  |
| **common**          | _infra_                | -       | _infra_              | _infra_                 | _infra_                     | _infra_                                     | _infra_                     | _infra_                                           | _infra_                | _infra_                 | _infra_                                | _infra_                |
| **console**         | _console_              | _infra_ | -                    | _console_               | _console_                   | _console_                                   | _console_                   | _console_                                         | _console_              | _console_               | _console_                              | _console_              |
| **crm**             | _no-consumer-surface_  | _infra_ | _console_            | -                       | _would-invert_              | CRM->FIN-01<br/>CRM->FIN-02                 | _no-producer-data_          | _would-invert_                                    | _no-producer-data_     | CRM->INS-01 (rejected)  | _no-consumer-surface_                  | _no-consumer-surface_  |
| **erp**             | _no-consumer-surface_  | _infra_ | _console_            | ERP->CRM-01 (rejected)  | -                           | ERP->FIN-01<br/>ERP->FIN-02<br/>ERP->FIN-03 | _no-producer-data_          | _no-producer-data_                                | _no-producer-data_     | ERP->INS-01             | ERP->POS-01<br/>ERP->POS-02 (rejected) | _no-consumer-surface_  |
| **finance**         | _no-consumer-surface_  | _infra_ | _console_            | FIN->CRM-01             | FIN->ERP-01<br/>FIN->ERP-02 | -                                           | FIN->FLOW-01 (rejected)     | _would-invert_                                    | FIN->HUB-01 (rejected) | FIN->INS-01 (rejected)  | _no-consumer-surface_                  | _no-consumer-surface_  |
| **flow**            | _no-consumer-surface_  | _infra_ | _console_            | _no-producer-data_      | _no-producer-data_          | FLOW->FIN-01                                | -                           | FLOW->HR-01 (rejected)<br/>FLOW->HR-02 (rejected) | _no-producer-data_     | FLOW->INS-01 (rejected) | _no-consumer-surface_                  | _no-consumer-surface_  |
| **hr**              | HR->BOOK-01 (rejected) | _infra_ | _console_            | HR->CRM-01 (rejected)   | _no-producer-data_          | HR->FIN-01                                  | HR->FLOW-01<br/>HR->FLOW-02 | -                                                 | HR->HUB-01             | HR->INS-01              | HR->POS-01 (rejected)                  | _no-consumer-surface_  |
| **hub**             | _hub-already-shared_   | _infra_ | _hub-already-shared_ | _hub-already-shared_    | _hub-already-shared_        | _hub-already-shared_                        | _hub-already-shared_        | HUB->HR-01 (rejected)                             | -                      | _hub-already-shared_    | _hub-already-shared_                   | _hub-already-shared_   |
| **insights**        | _insights-is-terminal_ | _infra_ | _console_            | _insights-is-terminal_  | _insights-is-terminal_      | _insights-is-terminal_                      | _insights-is-terminal_      | _insights-is-terminal_                            | _insights-is-terminal_ | -                       | _insights-is-terminal_                 | _insights-is-terminal_ |
| **pos**             | _no-consumer-surface_  | _infra_ | _console_            | POS->CRM-01             | _would-invert_              | POS->FIN-01 (rejected)                      | _no-producer-data_          | _would-invert_                                    | _no-producer-data_     | POS->INS-01 (rejected)  | -                                      | _no-consumer-surface_  |
| **support**         | _no-consumer-surface_  | _infra_ | _console_            | SUP->CRM-01             | _no-producer-data_          | _no-producer-data_                          | _no-producer-data_          | _no-producer-data_                                | _no-producer-data_     | SUP->INS-01 (rejected)  | _no-consumer-surface_                  | -                      |

### Product-wide connections (every domain is a consumer)

| Id          | Flows                                                            | Class | Status   |
| ----------- | ---------------------------------------------------------------- | ----- | -------- |
| IDENTITY-01 | one resolver so a person or company means one thing product-wide | MUST  | proposed |
| HUB->ALL-01 | tenant, user, team, app entitlement and config                   | IS    | existing |

### Intra-domain findings (not crossings, but blocking ones)

| Id          | Domain | Flows                                                                       | Class | Status   |
| ----------- | ------ | --------------------------------------------------------------------------- | ----- | -------- |
| HR->HR-01   | hr     | collapse hr_staff and hr_employees into one person record with one pay rate | MUST  | proposed |
| ERP->ERP-01 | erp    | reorder suggestions already exist; surface them where purchasing happens    | IS    | rejected |

### Connections outside the domain matrix

These involve an app with no contract domain, so they have no cell. See `01-inventory.md` on the app-to-domain mismatch.

| Id         | Producer | Consumer | Flows                                                                  | Class | Status   |
| ---------- | -------- | -------- | ---------------------------------------------------------------------- | ----- | -------- |
| REC->HR-01 | recruit  | hr       | hired candidate becomes an employee record without retyping the person | CAN   | proposed |

### Reason codes for pairs with no connection

| Code                   | Meaning                                                                                                                                                                                                                                      |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `self`                 | Same domain; not a crossing.                                                                                                                                                                                                                 |
| `infra`                | common holds auth and pagination wrappers, not business data. Every app already depends on it as infrastructure; there is no business fact to share in either direction.                                                                     |
| `console`              | console is RTYLR's own cross-tenant admin surface. A tenant domain reading it would cross a tenancy boundary, and console reading a tenant domain is already its whole job. Deliberately not connected.                                      |
| `no-consumer-surface`  | The consuming domain has no operator surface in this repo (booking and support have contracts and backend modules but no app; pos serves the separate terminal repo). A connection whose consumer cannot be built here fails bar rule 3.     |
| `no-producer-data`     | The producing domain holds no fact this consumer needs. Verified against the producer's entity list in 02-entity-graph.md.                                                                                                                   |
| `insights-is-terminal` | insights is a read-only projection layer. It consumes; nothing should consume it, because a second domain reading insights would depend on a derived number instead of its source. Finance already violates this once (see 10-doc-drift.md). |
| `would-invert`         | The natural direction of this pair is the reverse edge, which is listed. Adding this direction too would create a cycle and breaks rule 2.                                                                                                   |
| `hub-already-shared`   | hub is the tenant, user, team and entitlement domain that every app already reads; that crossing is built and working (HUB->ALL-01). Its outbound cells are marked here rather than re-proposed.                                             |

### Cell census

| Outcome                                       | Cells   |
| --------------------------------------------- | ------- |
| carries at least one connection id            | 28      |
| `infra`                                       | 22      |
| `no-producer-data`                            | 21      |
| `console`                                     | 19      |
| `no-consumer-surface`                         | 18      |
| `self`                                        | 12      |
| `hub-already-shared`                          | 9       |
| `insights-is-terminal`                        | 9       |
| `would-invert`                                | 6       |
| **total cells**                               | **144** |
| of which self-pairs (not crossings)           | 12      |
| **ordered cross-domain pairs with a verdict** | **132** |

<!-- END GENERATED MATRIX -->

## The three classes

### IS shared: data crosses today

| Id           | Crossing            | How it crosses                                                                                                                                                                    | Necessity    |
| ------------ | ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| HUB->ALL-01  | hub to every domain | real shared contract; every app reads `/hub/user/v1/users/me` and team/config endpoints                                                                                           | Essential    |
| POS->CRM-01  | pos to crm          | **same rows**: the POS customers module reads `crm_customers` directly (`~/Development/onion/internal/modules/pos_customers/internal/infrastructure/repository.go:30`)            | High value   |
| SUP->CRM-01  | support to crm      | the CRM app already makes 2 support-domain requests; the data reaching the screen is the gap, not the wire                                                                        | High value   |
| FIN->INS-01  | finance to insights | crosses in the **wrong direction**: `apps/finance` reads `/insights/financial/cash-flow`, so Finance depends on a derived projection of its own data. Logged in `10-doc-drift.md` | High value   |
| FLOW->INS-01 | flow to insights    | partly built via `insights.RoadmapProject.project_id` and Flow's own `flow_insights` module                                                                                       | Nice to have |
| POS->INS-01  | pos to insights     | dash already reads 9 pos endpoints                                                                                                                                                | Nice to have |
| CRM->INS-01  | crm to insights     | both `crm.dashboard` and the insights customers area exist                                                                                                                        | Nice to have |
| ERP->ERP-01  | intra-erp           | reorder suggestions are wired end to end (`apps/erp/src/features/reorder/useReorderSuggestions.ts:21`)                                                                            | Nice to have |

Manual re-typing also counts as IS shared, and it is the richest source of
findings because it proves demand without anyone having built anything. The
three confirmed instances:

| What a human retypes                | Into                      | Where the other half already lives | Cited                                                                                   |
| ----------------------------------- | ------------------------- | ---------------------------------- | --------------------------------------------------------------------------------------- |
| supplier / vendor name              | Finance bill create form  | `erp_suppliers`                    | `apps/finance/src/features/bills/Bills.tsx:59`                                          |
| client name, email, address, tax id | Finance invoice           | `crm_customers`                    | `packages/voxie/api/finance/invoices/v1/models.onk:146`                                 |
| customer name and email             | Support inbox             | `crm_customers`                    | `packages/voxie/api/support/inbox/v1/models.onk:5`                                      |
| the whole person                    | HR employee, after hiring | `careers_submissions`              | `~/Development/onion/internal/platform/database/schemas/recruit/recruit_v1_base.sql:80` |

### CAN be shared: both sides exist today

Sorted by necessity, then by how little has to change.

| Id           | Crossing                              | Gap bucket             | Necessity  |
| ------------ | ------------------------------------- | ---------------------- | ---------- |
| FLOW->FIN-01 | billable hours to Finance             | available today        | Essential  |
| CRM->FIN-01  | signed contract to draft invoice      | needs a contract field | Essential  |
| REC->HR-01   | hired candidate to employee           | needs an endpoint      | Essential  |
| HR->FIN-01   | payroll total to journal              | needs an endpoint      | Essential  |
| HR->FLOW-01  | pay rate to project labour cost       | needs an endpoint      | High value |
| HR->FLOW-02  | approved leave to assignment warning  | needs an endpoint      | High value |
| FIN->CRM-01  | outstanding balance on customer page  | needs a contract field | High value |
| FIN->ERP-01  | bill history on supplier page         | available today        | High value |
| ERP->FIN-02  | inventory valuation to balance sheet  | available today        | High value |
| ERP->FIN-03  | received-but-unbilled purchase orders | available today        | High value |
| FIN->ERP-02  | bill paid closes the purchase order   | needs a contract field | High value |
| ERP->INS-01  | product cost to margin by product     | needs an endpoint      | High value |
| HR->INS-01   | labour cost per period                | needs an endpoint      | High value |

### MUST be shared: the product is currently wrong without it

These are the priority. Each is a case where two domains can disagree, an
action is not blocked that should be, or a number nobody can compute gates
money or compliance.

| Id          | What is wrong today                                                                                                                                                                                                                                                                                                    | Severity     |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| HR->HUB-01  | Terminating an employee only sets `hr_employees.status='terminated'` (`~/Development/onion/internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:426`). No app access is revoked anywhere. An ex-employee keeps their login and every app entitlement until a human remembers to remove them by hand. | **critical** |
| HR->HR-01   | Two person tables (`hr_staff`, `hr_employees`) each carry `salary`, `hourly_rate` and `overtime_rate`, bridged by a nullable column. Two answers to "what do we pay this person" can differ, and leave and payroll key on different ids.                                                                               | high         |
| CRM->FIN-02 | An invoice's customer is four free strings. The same customer is a row in `crm_customers`. Nothing reconciles them, so revenue per customer is unanswerable and a renamed customer silently splits into two.                                                                                                           | high         |
| ERP->FIN-01 | `finance.Bill.supplier_id` is `@required @gt(0)` but the backend never joins `erp_suppliers`, and the UI fabricates the id by slugifying a typed name.                                                                                                                                                                 | high         |
| ERP->POS-01 | `erp.Product.stock_quantity` and `pos.MenuItem.stock` are two independent numbers for the same physical thing, and `pos.MenuItem` carries no cost at all, so item margin is uncomputable.                                                                                                                              | high         |
| IDENTITY-01 | 18 of 28 person-or-company entity pairs have no join. This blocks or degrades most of the above.                                                                                                                                                                                                                       | high         |

## What the matrix shows at a glance

**Producers cluster, consumers cluster.** `erp`, `crm` and `support` reference
no other domain in their contracts at all (`02-entity-graph.md`): they are pure
producers, and every proposal that makes one of them a consumer is new ground.
`finance` and `insights` are almost pure consumers. That asymmetry is why the
dependency graph comes out acyclic without much effort (`12-coupling.md`).

**Finance is the busiest consumer**, with five inbound proposals (from erp
twice more, crm, flow, hr). That concentration is itself a design signal: it is
the reason `08-roadmap.md` recommends one Finance-owned aggregate endpoint
rather than five separate cross-domain reads from the Finance app, which would
breach the two-foreign-reads-per-screen limit on the invoice and bill screens.

**The `would-invert` cells are load-bearing.** Six pairs are marked
`would-invert`, meaning this direction has no connection because the reverse
edge is the one that carries the fact: `booking -> hr`, `crm -> erp`,
`crm -> hr`, `finance -> hr`, `pos -> erp`, `pos -> hr`. In each the producing
domain holds nothing the consumer needs that the reverse edge does not already
move.

**Two pairs are proposed in both directions, and that is the audit's real
cycle risk.** `erp <-> finance` (five connections) and `crm <-> finance`
(three). Both survive, but only because the two directions move different
facts and neither reads the other's derived output: ERP tells Finance who a
supplier is and what stock is worth, while Finance tells ERP what has been
billed and paid. Under a strict node-level reading those are cycles between
two domains, so `12-coupling.md` does not wave them away: it models the graph
at connection granularity, states which fact flows which way, and shows that
no connection depends on the output of a connection pointing back at it.
`erp <-> finance` is the tighter of the two and is the reason the roadmap
sequences `ERP->FIN-01` before `FIN->ERP-02`.

**Two whole domains have contracts, backend modules and no consumer.**
`booking` (7 onion modules) and `support` (5) are built server-side and have no
app. Every proposal into them fails bar rule 3 for want of a surface, and every
proposal out of them is limited to what a tenant can actually generate, which
today is nothing for booking. This is a product-strategy finding, not a
plumbing one, and it is the largest single block of `no-consumer-surface` cells
in the matrix (18).
