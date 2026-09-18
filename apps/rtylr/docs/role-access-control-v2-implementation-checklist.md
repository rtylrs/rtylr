# Role and Access Control V2 implementation checklist

Status as of 2026-08-26. This checklist maps the implementation against
`docs/role-access-control-v2-plan.md`. `Partial` means the core path exists but
one or more acceptance criteria remain. `Blocked` means verification or
implementation needs an unavailable repository, service, or workflow.

## Plan section mapping

| Plan section                                           | Status                               | Evidence or remaining work                                                                                                                                                                                                                                                                                |
| ------------------------------------------------------ | ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2, 3, 4: role vocabulary and app independence          | Partial                              | Shared Dash roles, tenant app ids, app assignments, licenses, scopes, and app role seeds exist. Independent role selection is wired through Dash invite and member updates.                                                                                                                               |
| 5, 14: migration and compatibility                     | Partial                              | Backfill SQL, normalized dual writes, legacy role mapping, and source tracking exist. Ambiguity reporting, representative-tenant migration tests, rollback tests, and full legacy cleanup remain.                                                                                                         |
| 6: permission model                                    | Partial                              | Canonical permission types, aliases, effective-access evaluation, seeded catalog, and the reviewed permission inventory exist. Field-level authorization coverage remains.                                                                                                                                |
| 7, 22.3: feature and hide/unhide catalog               | Partial                              | App-level gates and targeted feature gates were migrated. The reviewed machine-readable and Markdown matrix now covers every canonical permission key; route-by-route enforcement and QA review remain.                                                                                                   |
| 8: data model                                          | Partial                              | Additive normalized authorization, provisioning, audit, group membership, scope, and role tables were added. Migration execution is blocked because the local MySQL socket is unavailable.                                                                                                                |
| 9: contracts and APIs                                  | Complete for the implemented surface | Voxie source models and generated Go and TypeScript clients include effective access, normalized roles, assignments, groups, provisioning, and OAuth profile data. Request-contract audit passes.                                                                                                         |
| 10: backend authorization                              | Partial                              | Resolver, auth hydration, fail-closed middleware, team persistence, Flow resource enforcement, POS checks, roles checks, and OAuth profile propagation exist. Several legacy app handlers still use old role checks and need systematic migration.                                                        |
| 11: shared frontend                                    | Complete for the implemented surface | Shared types, safe effective-access parsing, AppGuard, RoleGuard, permission helpers, app bridges, and Dash store hydration are implemented and typechecked.                                                                                                                                              |
| 12: Dash                                               | Partial                              | Invite, edit, normalized role display, app assignments, group ids, provisioning mode, app catalog, normalized custom-role create, edit, and archive flows, and the read-only Flow ownership boundary are implemented. Effective-access preview, full hierarchy, delegation, and complete audit UX remain. |
| 13.1, 13.3 to 13.10: tenant frontends                  | Partial                              | ERP, Finance, CRM, HR, Flow, Insights, Menu, and Recruit consume effective access for migrated gates. Full route, API, export, search, report, job, and sensitive-field coverage is not complete.                                                                                                         |
| 13.2, 22.12: POS                                       | Partial and blocked                  | POS login stores server-provided effective access and location selection uses central terminal permission. Typecheck and build are blocked because the checkout declares `@voxire/*` workspace packages but its `packages` directory is empty. Offline terminal sync and cache invalidation remain.       |
| 15: role-change behavior                               | Partial                              | License revocation, direct and group union, source tracking, normalized-first access refresh, and normalized custom-role edits are covered. Multi-role compatibility rules, immediate termination, and rehire flows remain.                                                                               |
| 16, 22.17: security and provisioning                   | Partial                              | Direct credentials are hashed, temporary password change is required, provisioning state is persisted, and protected routes are blocked while pending. Work-email verification, email OTP, authenticator or WebAuthn enrollment, recovery, rate limiting, and session invalidation workflows remain.      |
| 17: testing plan                                       | Partial                              | Resolver, Flow, POS backend, roles, OAuth, frontend permissions, and full frontend suites pass. The plan's complete POS threshold, HR conflict, migration, rollback, and cross-business scenario matrix is not complete.                                                                                  |
| 18: quality gates                                      | Partial                              | Frontend typecheck, build, tests, backend build, vet, targeted tests, contract generation, request audit, and diff checks pass. Database migration execution is blocked. The all-files OneKit boundary check reports 90 pre-existing legacy-wrapper violations.                                           |
| 20: acceptance criteria                                | Partial                              | Core hub ownership, independent app access, fail-closed access, Flow resource ownership, central POS checks, and the normalized custom-role lifecycle are implemented. Full backend coverage, provisioning workflows, migration rollback, and complete route enforcement remain.                          |
| 21, 22.16: authority boundaries                        | Complete for reviewed paths          | Dash is authoritative for cross-app roles, licenses, groups, and assignments. Flow owns project membership and HR retains its permission-role definitions.                                                                                                                                                |
| 22.13 to 22.15: lifecycle, audit, and policy additions | Partial                              | Audit schema, source-aware assignment records, normalized role lifecycle APIs, permission-impact preview, and paginated role audit APIs exist. Full retention policy, termination integration, delegation ceilings, cache invalidation, approval separation, and operational audit UX remain.             |
| Mobile client                                          | Blocked                              | No `pablo` checkout was present in the inspected development tree.                                                                                                                                                                                                                                        |

## Verification record

Passing commands:

- `pnpm test`: 651 files and 6,826 tests passed.
- `pnpm typecheck`: 20 tasks passed across 26 packages.
- `pnpm build`: passed, with all 10 application build tasks successful.
- `go build ./...`: passed in `onion`.
- `go vet ./...`: passed in `onion`.
- Targeted backend authorization tests: passed for team, authorization,
  workspaces, POS, roles, OAuth, and user modules.
- `make generate`: passed in `voxie`.
- `pnpm contracts:audit-requests`: passed with 612 calls and zero errors.
- `git diff --check`: passed in `rtylr`, `onion`, `voxie`, and `pos`.

Known blocked or failing checks:

- `pnpm typecheck` in `pos` cannot resolve its existing workspace dependencies
  because `/Users/abedamouneh/Development/pos/packages` is empty.
- The local MySQL migration test cannot start because the MySQL socket at
  `/tmp/mysql.sock` is unavailable.
- `pnpm contracts:enforce:all` reports 90 legacy OneKit boundary violations in
  existing app API wrappers and safe HTTP calls. The RBAC changes did not add a
  new legacy wrapper.
