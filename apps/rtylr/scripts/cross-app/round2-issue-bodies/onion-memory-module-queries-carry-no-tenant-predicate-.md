## Benefit

Operators get proof that one businesss AI memory can never surface in another businesss assistant session.

## Why

Gating and correctness (6); tenancy is part of every backend edge. Of memorys sqlc queries, none references business_id or owner_id (`backend_queries.csv`: 18 of 18 queries without either; handler tokens likewise none-found-in-module). Spot-check of the API layer found tenancy tokens nowhere in `internal/modules/memory/internal/api/memory_service.go`. This may be safe if memory is keyed purely by user-scoped session ids, but nothing states that invariant, no test asserts it, and echo feeds it into AI contexts where a leak would be invisible. Filed for owners to confirm-or-fix rather than asserted as a leak: static analysis cannot see key derivation.

## ROI

| Item                   | Value                                                                                                     |
| ---------------------- | --------------------------------------------------------------------------------------------------------- |
| Affected tenants       | all tenants using echo/Caret surfaces                                                                     |
| Manual work today      | none possible: nobody can audit what cannot be expressed                                                  |
| Modelled monthly value | risk removal; one cross-tenant memory leak would be a reportable incident                                 |
| Build cost             | 0.5 days to prove isolation and pin it with a test; 2 days if a scope column must be added and backfilled |
| Payback                | immediate (risk)                                                                                          |
| Confidence             | Inferred from code; severity contingent on what the sweep could not see                                   |

## What exists, what is missing

Queries: memory sqlc set, zero tenant tokens (backend_queries.csv). Handler: none-found-in-module. Consumers: echo wired with memoryModule twice in main.go:547.

## How to build it

Either document the isolation invariant (session/user-keyed only) with a citation and add a test asserting cross-user unreadability, or add business_id scoping mirroring the v8 tables convention (NOT NULL DEFAULT-free) plus backfill. Either way record the disposition here and in docs/sqlc-query-ownership.md style notes.

## Coupling cost

Intra-module hardening; no cross-domain edge. Owner: memory module maintainers. Index impact: adding business_id would want a leading composite; defer decision to fix path.

Coverage note: docs/cross-app/13-backend-inventory.md tenancy section.

Labelled audit:cross-app-round2-2026-08-22, area:memory, type:gap, severity:high.
