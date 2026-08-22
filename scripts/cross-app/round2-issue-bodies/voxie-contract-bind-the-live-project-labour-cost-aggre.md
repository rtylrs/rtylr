## Benefit

Flow projects can display what their labour has cost, computed server-side without exposing any pay rate.

## Why

Derivation (1), rollup (7). The aggregate exists end-to-end behind the labourcost port (#743) but no RPC binds it, so no consumer can reach it. Contract follows the live implementation field-for-field (billable-only scope, uncosted coverage counts) rather than inventing a parallel shape. Cross-links: rtylr#862 (partially superseded), rtylr#837, onion paired exposure issue.

## ROI

| Item                   | Value                                       |
| ---------------------- | ------------------------------------------- |
| Affected tenants       | Flow+HR services tenants                    |
| Manual work today      | spreadsheet cost tracking per project       |
| Modelled monthly value | 4h/mo PM time; earlier overrun detection    |
| Build cost             | contract 0.25d; onion RPC 1d; flow panel 2d |
| Payback                | ~1 month                                    |
| Confidence             | Inferred from code                          |

## What exists, what is missing

Drafted on branch: GetProjectLabourCost RPC on ProjectService (@get /projects/{id}/labour-cost); ProjectLabourCost message mirrors port fields with partial-labelling doc comments. HR->INS-01 may consume the same RPC.

## How to build it

Voxie; onion binds RPC in workspaces API over the port (permission-gated, tenant-resolved); rtylr renders with partial labels when uncosted_minutes > 0.

## Coupling cost

Edge flow -> hr already exists via port; contract formalises it. Aggregate-only output preserves salary boundary (staff contract precedent).

Labelled audit:cross-app-round2-2026-08-22, area:flow, type:gap, severity:high.
