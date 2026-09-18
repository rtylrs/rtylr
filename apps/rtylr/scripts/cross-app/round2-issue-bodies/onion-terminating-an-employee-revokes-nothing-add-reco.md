## Benefit

The day after this ships, ending someones employment removes their app access automatically instead of leaving their login and every app entitlement active until a human remembers.

## Why

Gating and correctness (taxonomy 6) plus trigger and automation (4). Round one filed this as rtylr#834; round two re-verified the backend and it still holds: termination writes two status columns and stops (`internal/modules/hr_core/internal/routes/handlers/employee_handlers.go:425`, `internal/modules/staff/sql/queries/staff.sql:26`), while no module anywhere reads a terminated status to revoke `user_app_permissions` or edit any apps array. The asymmetry is now sharper: team invite provisioning already writes across the boundary in the opposite direction, INSERTing an hr_employees row cross-schema (`internal/modules/team/sql/queries/team.sql:192-201`). Hire couples entitlement to employment; termination does not. This pairing over hub-side manual offboarding because the entitlement layer is what every surface reads.

## ROI

| Item                   | Value                                                                                                                              |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | every tenant with staff turnover; reference tenant: ~8 departures/yr                                                               |
| Manual work today      | admin must remember to strip apps per leaver; missed cases are silent                                                              |
| Modelled monthly value | 8 departures/yr x 15 min deprovisioning = 2h/mo reclaimed; one avoided incident of ex-employee access is far larger (unquantified) |
| Build cost             | 3 developer days (job + hook + tests)                                                                                              |
| Payback                | under 1 month                                                                                                                      |
| Confidence             | Inferred from code (counts modelled from the reference tenant)                                                                     |
| Cost of not doing it   | unrevoked access persists indefinitely per missed case; security exposure with no compensating control                             |

## What exists, what is missing

Status write: employee_handlers.go:425 sets `terminated`; staff.sql:26 terminates hr_staff. No consumer of terminated status exists (`grep -rn terminated internal/modules` shows only writers). Entitlement storage: user_app_permissions (shared schema), per-user apps array served by hubprofile (`internal/platform/hubprofile/profile.go:14`). Provisioning precedent writing cross-schema: team.sql:192-201.

## How to build it

Shape B4b then B3 of docs/cross-app/17-backend-plumbing.md (rtylr round-two branch). First the reconciliation job: a jobs entry finding users whose linked employee rows are terminated while permissions remain active, reporting before revoking. Then the synchronous half: on employee status change to terminated, deactivate matching user_app_permissions rows in the same request path via an explicit provider interface owned by team/user (not a direct module import). Idempotent, scoped by business_id, multi-business users lose access only to the terminating tenant.

## Coupling cost

Dependency: team/user own entitlement writes; hr emits intent. Direction hr -> hub-side modules via interface; owner: hr_core. Tenant predicate: business_id on both employee lookup and permission rows; indexes exist (owner-leading on user_app_permissions). Degrade: if HR is disabled for a tenant, job finds nothing and changes nothing; behaviour identical to today.

Dossier: tryrtylr/rtylr branch docs/cross-app-round2, docs/cross-app/05-connections/HR-HUB-01-revoke-access-on-termination.md. Frontend issue: rtylr#834. Related: onion#762 (token revocation on auth events).

Labelled audit:cross-app-round2-2026-08-22, area:hr_core, type:gap, severity:critical.
