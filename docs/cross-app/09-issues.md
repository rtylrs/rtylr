# 09. Issues filed

29 issues filed against `tryrtylr/rtylr` on 2026-08-21: **21 connection issues**, one per dossier, and **8 contract issues**, one per drafted `.onk` change. All carry the batch label `audit:cross-app-2026-08-21`.

Every body carries these headings in this order, as the brief requires: Benefit, Why, ROI, "What exists, what is missing", How to build it, Coupling cost, then a link to the dossier on this branch. Risks and Effort are appended where the dossier has them. The bodies are generated from the dossiers by `scripts/cross-app/build-issue-bodies.mjs`, which **fails rather than filing** if Benefit, Why or ROI is missing, so an issue cannot exist without them.

## Deduplication

Open issues were listed before filing. The repository had **two** open issues, #795 and #784, and neither duplicates anything here.

#795 ("two dead/unused role systems") is adjacent and worth reading alongside this audit: it already establishes that `APP_CONFIG.allowedRoles` is never consulted and that the real app-entry boundary is the per-user `apps` array in `packages/auth/src/guards/AppGuard.tsx:173`. **No new issue was filed for that**, and `07-plumbing.md` was corrected after reading it, because the brief itself points at `APP_CONFIG` as the entitlement check and it is not one. Logged as row 15 of `10-doc-drift.md`.

## Connection issues

| Issue | Id | Wave | Severity | Title |
| --- | --- | --- | --- | --- |
| [835](https://github.com/tryrtylr/rtylr/issues/835) | `ERP->FIN-01` | 1 | high | supplier identity for bills from erp to finance |
| [845](https://github.com/tryrtylr/rtylr/issues/845) | `ERP->FIN-02` | 1 | medium | inventory valuation for the balance sheet from erp to finance |
| [850](https://github.com/tryrtylr/rtylr/issues/850) | `ERP->FIN-03` | 1 | medium | received-but-unbilled purchase orders from erp to finance |
| [836](https://github.com/tryrtylr/rtylr/issues/836) | `FLOW->FIN-01` | 1 | high | unbilled billable hours from flow to finance |
| [854](https://github.com/tryrtylr/rtylr/issues/854) | `HR->FIN-01` | 1 | high | payroll total to the journal from hr to finance |
| [838](https://github.com/tryrtylr/rtylr/issues/838) | `HR->FLOW-02` | 1 | medium | approved leave gates assignment from hr to flow |
| [834](https://github.com/tryrtylr/rtylr/issues/834) | `HR->HUB-01` | 1 | critical | revoke app access on employment end from hr to hub |
| [842](https://github.com/tryrtylr/rtylr/issues/842) | `SUP->CRM-01` | 1 | medium | ticket history on the customer page from support to crm |
| [840](https://github.com/tryrtylr/rtylr/issues/840) | `CRM->FIN-02` | 2 | high | customer identity for invoices from crm to finance |
| [841](https://github.com/tryrtylr/rtylr/issues/841) | `FIN->CRM-01` | 2 | medium | outstanding balance on the customer page from finance to crm |
| [844](https://github.com/tryrtylr/rtylr/issues/844) | `FIN->ERP-01` | 2 | medium | bill history on the supplier page from finance to erp |
| [837](https://github.com/tryrtylr/rtylr/issues/837) | `HR->FLOW-01` | 2 | high | project labour cost from hr to flow |
| [849](https://github.com/tryrtylr/rtylr/issues/849) | `HR->INS-01` | 2 | medium | labour cost for margin reporting from hr to insights |
| [846](https://github.com/tryrtylr/rtylr/issues/846) | `REC->HR-01` | 2 | high | hired candidate becomes an employee from recruit to hr |
| [839](https://github.com/tryrtylr/rtylr/issues/839) | `CRM->FIN-01` | 3 | high | signed contract drafts an invoice from crm to finance |
| [853](https://github.com/tryrtylr/rtylr/issues/853) | `ERP->INS-01` | 3 | medium | product cost for margin reporting from erp to insights |
| [843](https://github.com/tryrtylr/rtylr/issues/843) | `ERP->POS-01` | 3 | high | one stock number and one cost per item from erp to pos |
| [851](https://github.com/tryrtylr/rtylr/issues/851) | `FIN->ERP-02` | 3 | medium | bill settlement closes the purchase order from finance to erp |
| [852](https://github.com/tryrtylr/rtylr/issues/852) | `POS->CRM-01` | 3 | medium | order history on the customer page from pos to crm |
| [848](https://github.com/tryrtylr/rtylr/issues/848) | `HR->HR-01` | 4 | high | one person record, one pay rate from hr to hr |
| [847](https://github.com/tryrtylr/rtylr/issues/847) | `IDENTITY-01` | 4 | high | one resolver for people and companies from hub to all |

## Contract issues

These land in **voxie**, not in this repository, and have a different owner from the consuming app work. Each cross-links the connections it serves.

| Issue | Draft | Severity | Title |
| --- | --- | --- | --- |
| [855](https://github.com/tryrtylr/rtylr/issues/855) | `06-proposed-contracts.md` item 1 | high | add range, project and billability filters to hr time entries |
| [856](https://github.com/tryrtylr/rtylr/issues/856) | `06-proposed-contracts.md` item 2 | medium | add a date range and pagination to hr leave requests |
| [857](https://github.com/tryrtylr/rtylr/issues/857) | `06-proposed-contracts.md` item 3 | high | give a finance invoice a crm customer_id |
| [858](https://github.com/tryrtylr/rtylr/issues/858) | `06-proposed-contracts.md` item 4 | high | link a pos menu item to its erp product |
| [859](https://github.com/tryrtylr/rtylr/issues/859) | `06-proposed-contracts.md` item 5 | medium | link a finance bill to the erp purchase order it settles |
| [860](https://github.com/tryrtylr/rtylr/issues/860) | `06-proposed-contracts.md` item 6 | high | make finance bill supplier_id nullable and add supplier_name |
| [861](https://github.com/tryrtylr/rtylr/issues/861) | `06-proposed-contracts.md` item 7 | medium | reference a crm customer from a support customer |
| [862](https://github.com/tryrtylr/rtylr/issues/862) | `06-proposed-contracts.md` item 8 | high | add hr labour-cost and insights product-margin aggregate endpoints |

## Labels used

| Label | Meaning | Created by this audit? |
| --- | --- | --- |
| `audit:cross-app-2026-08-21` | this batch | yes |
| `area:pos` | POS domain | yes |
| `area:support` | Support domain | yes |
| `area:hr`, `area:flow`, `area:finance`, `area:crm`, `area:erp`, `area:insights`, `area:recruit`, `area:dash`, `area:repo` | existing area labels | no |
| `type:gap` | "Missing: half-built, absent states, never finished" | no |
| `severity:critical` / `high` / `medium` | existing severity scale | no |

Both the producer and the consumer app carry an `area:` label on every connection issue, so an app owner sees every issue that touches their surface from either side. `hub` is labelled `area:dash`, because dash is hub's operator surface and there is no `area:hub` label.

Only one issue is `severity:critical`: **#834**, the employment-termination access revocation. It is the only security exposure in the audit.

## Caveats carried into the issue bodies

Three wave-1 and wave-2 issues ship **before** HR's duplicated person record is unified (`HR->HR-01`, wave 4), so each carries a partial-coverage caveat in its body rather than leaving it in the dossier:

- **#838** (`HR->FLOW-02`): no warning fires for a tenant recording leave in the older `staff_id` system.
- **#837** (`HR->FLOW-01`): reads a stale rate if a tenant maintains rates on `hr_staff`.
- **#834** (`HR->HUB-01`): must watch both `status` columns or be bypassable.

Shipping them first is still right: the alternative is blocking a critical security fix behind a multi-week migration. See `08-roadmap.md`.

## Regenerating

```bash
node scripts/cross-app/build-issue-bodies.mjs /tmp/issue-bodies
```

The script writes one body per issue plus a `manifest.json` of titles and labels. It does not file anything; filing was a separate `gh issue create` per row.
