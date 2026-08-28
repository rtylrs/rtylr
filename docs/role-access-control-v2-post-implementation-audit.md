# RTYLR Role and Access Control V2 post-implementation audit

Date: 2026-08-27

## Verdict

NOT RELEASE READY under the strict release request. The critical cross-application authority, stale POS authorization, app entry enforcement, scope validation, MFA lifecycle, OTP verification, recovery, migration safety, and Dash guided-provisioning gaps are addressed in source. The generated direct gate now proves concrete allowed boundary reach and modeled app or permission denials for all 102 applicable permissions. Production sign-off remains blocked because the generated probe does not independently close every row-specific scope, wrong-tenant, stale-auth, delegation, response, side-effect, and audit dimension required by the release request.

No commits or pushes were made. Existing dirty worktrees were preserved.

## 2026-08-27 closure verification

- The canonical matrix contains 104 records and 22 action columns. There are 102 applicable employee permissions and two explicitly not-applicable public Menu display permissions.
- Every applicable matrix action has a concrete handler or API test reference. The direct suite runs 34 Onion package targets and passed with exit code 0.
- Each applicable non-Hub action has direct assertions for granted boundary reach, missing permission, missing application license, revoked application license, permission assigned to the wrong application, and resolved no-access state. Hub actions have direct granted and missing-permission assertions.
- The direct probe intentionally stops at the real authorization boundary when a nil service would otherwise perform business work. It therefore does not claim operation success, response/body correctness, no-data-leak behavior, side-effect absence, audit emission, scope isolation, wrong-tenant behavior, stale authorization, or delegation ceilings for every row.
- Fresh MySQL migration replay applied 173 migrations with zero failures. The sequential idempotence replay skipped all 173 with zero failures.
- RTYLR forced typecheck and build passed with 20 typecheck tasks and 10 build tasks. The complete unit suite passed with 651 files and 6,826 tests. The targeted E2E rerun for changed Flow fixtures passed 37 tests.
- POS passed 177 test files and 2,144 tests, typecheck, build, native cargo check, and 125 Rust tests. Voxie `onek check api`, generation, and typecheck passed.
- Gosec, Semgrep, Trivy, and osv-scanner completed. Trivy source and image scans report zero vulnerabilities, secrets, and misconfigurations after excluding ignored worktrees, upgrading Alpine packages, and applying the documented non-reachable Onion advisory exception for `GO-2026-5932`.

## Scope and evidence

Repositories reviewed:

- Frontend: `/Users/abedamouneh/Development/rtylr`
- Backend: `/Users/abedamouneh/Development/onion`
- Canonical contracts: `/Users/abedamouneh/Development/voxie`
- POS client: `/Users/abedamouneh/Development/pos`
- Supporting documentation: `/Users/abedamouneh/Development/vault`

Primary implementation surfaces:

- POS authority: `/Users/abedamouneh/Development/onion/internal/modules/pos_users`, `/Users/abedamouneh/Development/pos/src/features/staff`, and `/Users/abedamouneh/Development/voxie/api/pos/user/v1/models.onk`
- Authorization middleware: `/Users/abedamouneh/Development/onion/internal/platform/server/http/auth_middleware.go`
- Effective access model: `/Users/abedamouneh/Development/onion/internal/platform/authorization`
- Scope validation: `/Users/abedamouneh/Development/onion/internal/modules/roles/internal/infrastructure/normalized_repository.go`
- MFA and recovery: `/Users/abedamouneh/Development/onion/internal/modules/security` and `/Users/abedamouneh/Development/rtylr/apps/dash/src/features/security/Security.tsx`
- Email OTP: `/Users/abedamouneh/Development/onion/internal/modules/user` and `/Users/abedamouneh/Development/rtylr/apps/auth/src/features/email-verification/VerifyEmail.tsx`
- Migrations: `/Users/abedamouneh/Development/onion/internal/platform/database/schemas/shared/zz_role_access_v2*.sql`
- Dash provisioning workflow: `/Users/abedamouneh/Development/rtylr/apps/dash/src/features/team/InviteUser.tsx`

## Critical findings and disposition

| Finding                                                   | Disposition                                                | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --------------------------------------------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| POS can edit Dash-owned roles and capabilities            | Addressed                                                  | POS create and update no longer accept or persist role/capability mutations. The staff UI exposes roster, PIN, and status operations only. Verification resolves role and capabilities from normalized Dash authorization and returns a managed-in-Dash indicator and authorization version.                                                                                                                                                                                                                   |
| POS persists stale authorization and replays offline work | Addressed with fail-closed behavior                        | Employee authorization version is captured at verification and queue time. The employee session is no longer persisted. Offline replay requires the same employee, an active session, and the same authorization version. Legacy queue rows without the metadata are refused instead of replayed.                                                                                                                                                                                                              |
| App-level backend enforcement is incomplete               | Concrete boundary coverage added, strict dimensions remain | HTTP middleware maps every licensed application path to an effective-access app decision. Concrete handler or API tests now cover all 102 applicable matrix permissions, with direct allowed and modeled denial cases. The generated probe does not independently prove every row-specific scope, wrong-tenant, stale-auth, delegation, response, side-effect, or audit dimension.                                                                                                                             |
| Scope ownership and schema constraints are incomplete     | Addressed with application and migration safeguards        | Assignment writes validate that POS locations and Flow projects belong to the active business. Integrity migration preflights orphaned records, invalid scope types, misowned scope IDs, duplicate active assignments, role-permission orphans, and invalid assignment targets. It adds conditional checks, unique indices, and foreign keys where the schema supports them. Polymorphic scope IDs remain application-validated because one column cannot directly reference both location and project tables. |
| MFA, OTP, and recovery are missing                        | Addressed                                                  | TOTP secrets are encrypted, recovery codes are bcrypt-hashed and single-use, verification is time-limited, and middleware blocks sensitive access until enrollment and verification. Email verification uses a six-digit, hashed, expiring, consumed-once OTP. Dash provides enrollment, QR setup, challenge verification, recovery, and recovery-code presentation. Auth redirects MFA-gated users to Dash, and Dash’s Security route is the enforced setup/challenge entry point.                            |
| Migration uses unsafe `INSERT IGNORE`                     | Addressed and executed                                     | RBAC base and backfill migrations contain no `INSERT IGNORE`. They use controlled upserts or `NOT EXISTS`, report affected-row counts, and fail closed on ambiguous or orphaned data. A fresh disposable MySQL run applied 173 migrations with zero failures, and a sequential second run skipped all 173.                                                                                                                                                                                                     |
| Dash UI does not match required mockup workflows          | Addressed for the primary provisioning flow                | Member creation now has Identity, Access setup, Review, and Activate stages, including sign-in method, per-app roles, POS locations, group selection, effective activation guard, and final lifecycle summary. Roles expose application ownership and permission editing. Member detail exposes effective access and ownership. Full history, impact replacement tooling, and advanced bulk review remain secondary follow-ups.                                                                                |

## Implementation details

### POS authority boundary

The POS API is now limited to POS-owned operational data. Role, discount, void, refund, drawer, and capability fields are no longer writable through POS create or update paths. Existing legacy columns remain readable for compatibility, but they are not used as the authorization source.

The POS client links operators to Dash for role and capability management. PIN verification returns the normalized authorization version, and the client does not persist the employee authorization snapshot across restarts.

### Revocation and offline replay

Every queued order now carries the employee identifier and authorization version when the queue entry is created. Replay checks the current authenticated employee and version before submitting. A missing or changed version produces a terminal failure for that queue item, which prevents an old operation from being silently retried after revocation.

The server remains authoritative on replay. App access and exact POS feature permissions are evaluated on the request path, so a client-side queue check cannot grant access that the backend has revoked.

### Backend app and feature enforcement

The HTTP middleware now rejects authenticated requests to `/pos`, `/erp`, `/hr`, `/crm`, `/finance`, `/flow`, `/insights`, `/menu`, and `/recruit` when the effective authorization document does not contain a licensed, accessible app. Hub, common, and health paths are not treated as application paths.

POS handlers use exact permissions such as `orders:view`, `orders:create`, `orders:void`, `orders:refund`, `orders:discount`, `reports:view`, `cash:drop`, `cash:pickup`, `cash:approve`, `terminals:manage`, `employees:manage`, and `locations:manage`. Tests cover application path mapping and denied access behavior.

The remaining assurance item is a systematic negative-test pass for every non-POS feature-matrix action. The global app boundary is in place, but a release reviewer should still prove each sensitive create, edit, delete, approve, export, publish, and configure operation at its handler or service boundary.

### Tenant, scope, and schema integrity

Normalized role assignment writes reject:

- business IDs that do not match the active tenant;
- POS locations owned by another business;
- Flow projects owned by another business;
- unknown scope types and non-positive IDs;
- role assignments with invalid user or group targets.

The integrity migration preflights authorization members, licenses, role assignments, role permissions, assignment scopes, and scope ownership before adding constraints. It stops rather than silently repairing ambiguous production data.

The schema still represents `scope_id` as a polymorphic identifier. The migration adds a type and positive-ID check, while the repository validates ownership against the correct tenant table. A future normalized scope table would allow direct foreign keys for this last polymorphic relationship.

### Authentication lifecycle

The provisioning status now includes `mfa_verified`, calculated from the required/enrolled state and the five-minute verification window. The auth redirect sends users with required but incomplete MFA to Dash. Dash blocks other routes until the user visits Security and completes enrollment or enters an authenticator or recovery code.

Email verification is no longer a no-op. Resend issues a generic-success response to avoid account enumeration, while verification consumes only a matching, unexpired, single-use hash and activates the normalized member when the remaining provisioning requirements are satisfied.

### Migration safety

The RBAC migrations are idempotent without silent ignore behavior. They use deterministic upserts for known role-permission rows, explicit duplicate prevention for backfill rows, conditional constraint creation, and `ROW_COUNT()` reporting. A production rollout must still run the migration on a MySQL snapshot, record preflight output, compare affected counts, and verify the post-migration authorization counts.

### Dash workflow coverage

The primary member flow now matches the mockup’s sequence:

1. Identity: name, work email, invitation or direct credentials.
2. Access setup: Dash role, app licenses, per-app custom roles, group, and POS locations.
3. Review: identity, sign-in method, Dash role, applications, and lifecycle guard.
4. Activate: initial state, application summary, authorization ownership, and final create/invite action.

The effective-access panel on member detail shows app, source, display role, permissions, and scope counts. Role management labels Dash-owned authority and custom role editing includes application context.

## Verification results

Passed:

- Onion: `go test -count=1 ./...`
- Onion: `go build ./...`
- Onion: `go vet ./...`
- Onion: `make generate` for all 132 SQLC modules
- Onion direct matrix gate: 34 package targets passed, with concrete boundary assertions for all 102 applicable permissions
- Fresh disposable MySQL migration: 173 applied, 0 skipped, 0 failed
- Sequential migration replay: 0 applied, 173 skipped, 0 failed
- Canonical Voxie and workspace Voxie: `onek check api`, generation, and typecheck completed
- RTYLR: `TURBO_FORCE=true pnpm build`, 20 forced typecheck tasks and 10 builds passed
- RTYLR: `pnpm test`, 651 files and 6,826 tests passed after the dependency update
- RTYLR targeted changed Flow E2E suites: 37 passed, 0 skipped, 0 failed
- POS: `pnpm test`, 177 files and 2,144 tests passed; typecheck and build passed
- POS native checks: cargo check and cargo test passed, with 125 Rust tests passed
- RTYLR: `pnpm contracts:audit-requests`, 619 OneKit calls across 1,198 files, 0 errors
- RTYLR: `pnpm contracts:enforce:all`, 0 app transport violations
- RTYLR: `pnpm format:check`, all files matched Prettier style
- Gosec 2.28.0 with generated files excluded: 0 findings
- Semgrep 1.174.0: 0 findings
- Trivy 0.74.0: source and built frontend image scans passed with 0 vulnerabilities, secrets, and misconfigurations
- osv-scanner 2.5.1: 0 vulnerable packages
- Built frontend image: non-root `nginx` user, healthcheck present, live HTTP smoke test passed
- `gofmt` reported no modified Go files requiring formatting
- `git diff --check` passed for the session-touched files in RTYLR and Onion
- No `INSERT IGNORE` remains in `zz_role_access_v2*.sql`

Classified exceptions:

- Onion `.trivyignore` contains only `GO-2026-5932`. The advisory is for the unmaintained `x/crypto/openpgp` package, onion imports only `x/crypto/bcrypt`, and no upstream fixed version exists. osv-scanner and gosec report no reachable issue.
- Trivy source scanning excludes `.claude/worktrees`, which is git-ignored local worktree data and is not part of a release artifact.
- The full Playwright suite had already passed earlier in this session with 111 tests. The post-fixture-change rerun covered all nine changed Flow fixture files and passed 37 tests.

## Release gates still open

These are the remaining strict-evidence blockers before production sign-off:

- Add row-specific direct assertions for scope isolation and wrong-tenant denials across the applicable matrix actions.
- Add direct stale-authorization and delegation-ceiling assertions where those dimensions apply.
- Add direct response/body, no-data-leak, side-effect, and audit assertions for each applicable action, or link existing handler tests that prove those exact outcomes.
- Obtain policy approval for the precise `.trivyignore` exception, or replace it with an accepted VEX declaration.
- A normalized scope table remains an architectural follow-up only if direct database foreign keys for polymorphic scopes are required. It is not the current strict-evidence blocker.
