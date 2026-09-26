# Working in RTYLR

These instructions apply throughout this repository. Follow more specific `AGENTS.md` files in subdirectories when present. Keep this guide practical and update it when the repository changes; do not turn it into a work log.

## Repository map

This is `tryrtylr/rtylr`, a pnpm monorepo. Inspect the current checkout, branch, status and package scripts before making assumptions based on older layouts or sibling repositories.

| Path               | Responsibility                                                                               |
| ------------------ | -------------------------------------------------------------------------------------------- |
| `clients/`         | pnpm workspace root; every JavaScript/TypeScript package lives under it                      |
| `clients/web/rtylr/apps/` | Auth, Dash, CRM, ERP, HR, Finance, Flow, Insights, Recruit and Menu frontends         |
| `clients/web/landing/` | Next.js marketing site, public legal pages and localized copy; production uses static export |
| `clients/pos/`     | POS frontend and native/Tauri code; web and native checks are separate                       |
| `clients/web/docs/` | Product documentation site                                                                  |
| `clients/web/mcp/` | MCP server and tool integration                                                             |
| `clients/packages/` | Shared `@voxire/*` libraries and the generated contracts package; changes can affect multiple products |
| `voxie/`           | API contracts (`.onk`) and the generated Go/TypeScript/OpenAPI output consumed on both sides  |
| `services/onion/`  | Go backend: modules, platform packages, sqlc, worker                                         |
| `.github/`         | Release and deployment automation                                                            |

Deployment repositories (`tryrtylr/deploy`, `tryrtylr/infra`, `tryrtylr/pipelines`) are separate. A frontend change does not establish that the backend supports a field, enforces a permission or implements a retention policy.

## Scope and implementation

- Preserve unrelated work. Use an isolated branch/worktree when the checkout contains someone else's changes. Never reset or clean their files to make your task easier.
- Solve the requested problem with the smallest coherent change. Avoid unrelated refactors, dependency upgrades, formatting sweeps and new abstractions without a concrete need.
- Follow nearby feature structure and existing shared components, hooks, API helpers and styling conventions. Keep shared behavior in the appropriate package instead of copying it between apps.
- Keep TypeScript types meaningful. Avoid `any`, unsafe assertions and suppressed errors as substitutes for understanding a contract.
- Use the established generated-client/API boundary and result/error conventions. Do not invent request fields, bypass transport helpers or hand-edit generated code. Contract changes belong in `voxie/api/`; run `make -C voxie generate` so the contract, its generated consumers and the Go handlers land in one change.
- `voxie/gen/` is generated and gitignored. Never commit it; regenerate with `make -C voxie generate` (or `pnpm --filter rtylr typecheck`, which generates it when missing) instead of editing it.
- Preserve tenant boundaries, authorization, offline behavior and account/session handling. Client-side visibility is not server-side authorization. Surface failures honestly instead of disguising them as empty or successful results.
- Keep user-facing text in the existing locale system. Update affected English, French and Arabic resources together, preserve interpolation and check RTL where relevant. Do not enable an unshipped locale just because its strings exist.
- Preserve accessible names, keyboard behavior, focus, semantic markup and loading/error states when changing UI.

## Comments and documentation

- **No huge comments. Add only comments that are actually needed.** Prefer clear names and small functions.
- Explain a non-obvious reason, invariant, external constraint or dangerous edge case in a few lines. Do not narrate what the code already says.
- No essay-length component headers, change histories, task transcripts, decorative section banners or explanations of rejected approaches in source files. Put substantial architectural rationale in an appropriate document and link it only where useful.
- Update or remove stale comments when behavior changes. Use a TODO only for a concrete unresolved item with enough context to act on it.
- Document behavior, limitations and evidence accurately. Do not claim a check passed, a feature exists or a service is compliant without verification.

## Setup and commands

Use the pnpm version declared in `package.json` and the root lockfile. From the repository root:

```sh
pnpm -C clients install --frozen-lockfile
make -C voxie generate   # contracts are generated, not committed
git config core.hooksPath .githooks
```

Run package-specific scripts with `pnpm -C clients --filter <package-name> <script>`. Read that package's `package.json`: names and available scripts differ between products.

Examples, all invoked from the repository root:

```sh
# One app; replace @rtylr/auth with the affected package.
pnpm -C clients --filter @rtylr/auth typecheck
pnpm -C clients --filter @rtylr/auth build

# Focused main-frontend/shared-package tests; supply actual test-file paths.
clients/web/rtylr/scripts/vitest.sh run clients/web/rtylr/apps/auth/src/features/register/Register.test.tsx

# Full main-frontend/shared-package suite, including its collected-file check.
pnpm -C clients/web/rtylr test

# Broader main-frontend checks when the change warrants them.
pnpm -C clients -r --filter './web/rtylr/apps/*' typecheck
pnpm -C clients -r --filter './packages/*' typecheck
pnpm -C clients -r --filter './web/rtylr/apps/*' build

# Landing checks read exported HTML, so build first.
pnpm -C clients --filter @rtylr/landing build
pnpm -C clients --filter @rtylr/landing test:unit

# Check changed app transport boundaries.
pnpm -C clients/web/rtylr contracts:enforce
```

pnpm filters are relative to the workspace root at `clients/`. Some aggregate scripts under `clients/web/rtylr` use workspace-root-relative filters: verify what actually ran. **“No projects matched” or zero collected tests is not a successful check.** Use explicit `pnpm -C clients` filters when needed. Inspect scripts again before relying on this workaround.

## Testing that earns its cost

- Test behavior and realistic risks, not implementation trivia. Prefer a regression test that would fail for the original bug and exercise the corrected outcome.
- Start with affected tests and relevant typechecks/builds. Expand to shared consumers or the full suite when changes cross app boundaries, affect contracts/auth/payments/consent, or leave material uncertainty. Complete required hooks and CI checks.
- Do not add tests merely to assert a constant, file contents, class names, an obvious getter or the same logic copied from the implementation. Avoid broad snapshots and extensive mocks that make broken behavior pass.
- Documentation-only changes usually need content, link/path and formatting review, not a full application build. Copy or styling changes need the applicable locale, rendering or accessibility checks; do not invent unrelated unit tests.
- Use real translations when testing rendered copy. Mock external boundaries rather than the behavior under test. Check failures and relevant edge cases, not every theoretical permutation.
- Use browser tests for behavior that depends on a browser, such as navigation, consent, persistence or focus. Use the existing fixtures and block external/production traffic. Tests must not send real campaigns, payments, telemetry or customer mutations.
- POS native changes need the relevant Rust/Tauri checks; a web build does not validate native behavior. Landing checks must exercise the static export where that affects the result.
- Once relevant checks pass, do not rerun them without a change or unresolved concern that justifies it. Report what ran, what failed and what was not checked; distinguish baseline failures from regressions with evidence.

## Privacy, security and legal claims

- Keep secrets and customer data out of code, logs, fixtures and commits. Use synthetic test data and existing configuration mechanisms.
- Optional analytics must respect both acceptance and withdrawal. Do not enable session recording through a generic analytics choice or add tracking that bypasses shared consent handling.
- Consult `tryrtylr/vault` (`rtylr/legal/README.md`) when changing public policies, data collection, vendors, retention, billing promises or security claims. Keep localized policies consistent with actual behavior.
- Do not invent an operating entity, jurisdiction, certification, audit result, signed agreement, deletion guarantee or transfer safeguard. Clearly distinguish drafts and proposed controls from completed agreements and deployed controls.
- Preserve existing customer commitments and required notice/versioning when changing terms. A page edit is not proof that customers received or accepted an amendment.
- Respect repository and dependency licenses; do not relicense code or promise commercial/self-host rights without verifying authorization.

## Delivery

- Review the final diff, format changed files, run `git diff --check` and keep generated output, reports and local artifacts out of the commit unless deliberately required.
- Keep PRs focused. Explain the problem, resulting behavior, meaningful validation and remaining limitations; omit task chronology and unsupported claims.
- When a PR is requested, commit, push and open or update it. Verify its state and report its link. Do not silently stop at a local patch.
- Opening a PR does not authorize merging, deploying, publishing packages or notifying customers. Follow the requested scope and inspect the relevant release workflow before taking those actions.
