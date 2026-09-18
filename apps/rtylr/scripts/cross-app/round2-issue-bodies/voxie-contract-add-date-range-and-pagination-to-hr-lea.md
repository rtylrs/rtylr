## Benefit

Callers can ask who is on approved leave between two dates in one paginated request.

## Why

Serves HR->FLOW-02 (leave gates assignment). ListLeaveRequestsRequest filters by employee_id and status only; overlap questions require unbounded client-side filtering. Cross-links: rtylr#856 (misfiled round-one contract issue), rtylr#838 (consumer feature).

## ROI

| Item                   | Value                                               |
| ---------------------- | --------------------------------------------------- |
| Affected tenants       | Flow+HR tenants                                     |
| Manual work today      | PMs consult leave calendar manually                 |
| Modelled monthly value | ~1.5h avoided rework (modelled)                     |
| Build cost             | contract 0.25d; backend 0.5d; frontend part of #838 |
| Payback                | < 1 month after landing                             |
| Confidence             | Inferred from code                                  |

## What exists, what is missing

Drafted on branch docs/cross-app-round2, api/hr/leave/v1/models.onk: from_date/to_date @nullable queries with documented overlap semantics; limit/offset added matching hr house style.

## How to build it

Land contract here; onion serves (paired issue); packages pointer bump; rtylr consumes. Order: voxie, packages, onion, rtylr.

## Coupling cost

No new domain edge; refines existing hr->flow read. Filters optional: absent means include-all, preserving current callers.

Labelled audit:cross-app-round2-2026-08-22, area:hr, type:gap, severity:medium.
