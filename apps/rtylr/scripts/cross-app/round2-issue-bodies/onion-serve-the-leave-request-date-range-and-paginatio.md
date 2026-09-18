## Benefit

Planning screens can ask who is away between two dates, so work stops being assigned to people on holiday.

## Why

Context in place (5) and gating (6). ListLeaveRequestsRequest filters by employee and status only; asking overlap requires fetching all leave and filtering client-side, which no consumer should do. Drafted in voxie (branch docs/cross-app-round2, hr/leave/v1/models.onk: from_date/to_date with documented overlap semantics plus limit/offset). Backend follow-up lands here. Pairs with rtylr#838 (HR->FLOW-02) whose warning feature needs exactly this query.

## ROI

| Item                   | Value                                                         |
| ---------------------- | ------------------------------------------------------------- |
| Affected tenants       | all Flow+HR tenants                                           |
| Manual work today      | PMs check the leave calendar by memory before assigning       |
| Modelled monthly value | avoids ~2 misassignments/mo x 45 min rework = 1.5h (modelled) |
| Build cost             | 0.5 day backend after contract lands                          |
| Payback                | under 1 month                                                 |
| Confidence             | Inferred from code                                            |

## What exists, what is missing

Contract: drafted on voxie branch. Storage: hr_leave_requests has start_date/end_date/status (schemas/hr/schema.sql:24 region). Query: listLeaveRequests lacks range predicates today. Tenancy note: hr_leave ownerID semantics differ from other HR modules (presence.go comment); pass caller-scoped ids accordingly.

## How to build it

Add range predicates with the documented overlap semantics (start <= to AND end >= from), limit/offset pagination, index check against hr_leave leading columns; expose via existing RPC; unit test the boundary-overlap case named in the draft comment.

## Coupling cost

Consumer flow reads hr via existing listLeaveRequests; owner hr_leave. Predicate: caller identity per module convention, not business_id blindly. Degrade: consumers hide the panel when HR off; empty set otherwise.

Draft: voxie branch docs/cross-app-round2 api/hr/leave/v1/models.onk. Issues: rtylr#856 (contract), rtylr#838 (feature).

Labelled audit:cross-app-round2-2026-08-22, area:workspaces, type:gap, severity:medium.
