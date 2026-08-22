## Benefit

A project manager opens a Flow project and sees what its labour has cost so far, with rates never leaving HR.

## Why

Derivation (1) and rollup (7). Every input already exists and is joined server-side by the labourcost port landed in #743: SQL inside hr_payroll joins hr_time_entries to hr_employees.hourly_rate (`internal/modules/hr_payroll/sql/queries/labour_cost.sql`), workspaces consumes it through SetLabourCostProvider with a noop fallback (`internal/modules/workspaces/internal/application/service.go:113-129`, wired `cmd/api/main.go:466`). Nothing exposes it: no RPC calls Service.GetProjectLabourCost, so the capability answers no one. This pairing (expose through workspaces) beats a new hr.cost service because the port, permission posture and noop fallback all exist; only the wire binding is missing.

## ROI

| Item                   | Value                                                                                            |
| ---------------------- | ------------------------------------------------------------------------------------------------ |
| Affected tenants       | services tenants running Flow + HR (reference tenant: 40-person, 6 projects)                     |
| Manual work today      | PMs export hours to a spreadsheet and multiply by remembered rates                               |
| Modelled monthly value | 6 projects x 2 cost checks/mo x 20 min = 4h reclaimed per PM team; overruns caught weeks earlier |
| Build cost             | 1.5 developer days backend (RPC + contract binding), 2 days frontend panel                       |
| Payback                | about 1 month                                                                                    |
| Confidence             | Modelled from assumption on check frequency; code facts measured                                 |

## What exists, what is missing

Port: internal/platform/labourcost/contract.go (ProjectCost: CostCents, CostedMinutes, UncostedMinutes, UncostedEmployeeCount). Implementation: hr_payroll repository.go:450. Consumer plumbing live but unconsumed. Contract drafted on voxie branch docs/cross-app-round2 (flow/project/v1 service + models). No .onk today: contracts.csv has no labour message.

## How to build it

Bind the drafted getProjectLabourCost RPC on flow.project.v1 to Service.GetProjectLabourCost in the workspaces API layer; resolve business id via tenant.ResolveBusinessID; keep billable-only scope documented. Permission-gate on a Flow project-level visibility rule following presence.gos reduced-projection reasoning. Frontend panel consumes with rate-coverage labelling.

## Coupling cost

Direction flow -> hr via existing port; owner hr_payroll. Tenant predicate business_id on both sides of the join (labour_cost.sql WHERE e.business_id = ? AND emp.business_id = ?); serves off idx_time_entries_owner-style keys plus project filter, verify against docs/tenant-index-matrix.md and add business-leading index if EXPLAIN disagrees. No transaction. Degrade: noop fallback returns zero-value struct; UI must render unknown, not zero.

Dossier: docs/cross-app/05-connections/HR-FLOW-01-project-labour-cost.md (round-two backend section). Issues: rtylr#837, rtylr#862 partially. Related onion PR #743.

Labelled audit:cross-app-round2-2026-08-22, area:workspaces, type:gap, severity:high.
