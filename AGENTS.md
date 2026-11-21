# Working in RTYLR

These instructions apply throughout this repository. More specific `AGENTS.md` files in subdirectories take precedence. Keep this guide practical and current; do not turn it into a work log.

## Repository map

This is `rtylrs/rtylr`, a pnpm monorepo. Inspect the checkout, branch, status and package scripts before relying on older layouts or sibling repositories.

| Path                   | Responsibility                                                                                        |
| ---------------------- | ----------------------------------------------------------------------------------------------------- |
| `clients/`             | pnpm workspace root; every JavaScript/TypeScript package lives under it                               |
| `clients/app/`         | Portal: business modules and the public customer menu in `src/modules`; SSO, Recruit and Rennili (the pager) in `entrypoints/` |
| `clients/web/landing/` | Next.js marketing site, public legal pages and localized copy; production uses static export          |
| `clients/web/docs/`    | Product documentation site                                                                            |
| `clients/pos/`         | POS frontend and native/Tauri code; web and native checks are separate                                |
| `clients/mcp/`         | MCP server and tool integration                                                                       |
| `clients/packages/`    | Shared `@voxire/*` libraries and the generated contracts package; changes can affect several products |
| `voxie/`               | API contracts (TypeSpec, compiled by onekit's `onkt`) and generated Go/TypeScript/OpenAPI output      |
| `services/onion/`      | Go backend: modules, platform packages, sqlc, worker                                                  |
| `services/searxng/`    | SearXNG image and settings                                                                            |
| `.github/`             | Release and deployment automation                                                                     |

Deployment repositories (`rtylrs/infra`, `rtylrs/pipelines`) are separate. A frontend change does not prove the backend supports a field, enforces a permission or implements a retention policy.

## Scope and implementation

- Preserve unrelated work. Use an isolated branch or worktree when the checkout holds someone else's changes; never reset or clean their files.
- Make the smallest coherent change. Avoid unrelated refactors, dependency upgrades, formatting sweeps and speculative abstractions.
- Follow nearby structure and reuse existing components, hooks, API helpers and styling. Put shared behavior in the right package instead of copying it between apps.
- Keep TypeScript types meaningful. Do not use `any`, unsafe assertions or suppressed errors in place of understanding a contract.
- Use the generated client, transport helpers and existing result/error conventions. Do not invent request fields or hand-edit generated code.
- Contract changes belong in `voxie/api/`. Run `make -C voxie generate` so the contract, its generated consumers and the Go handlers land in one change. `voxie/gen/` is generated and gitignored: never commit or edit it (`pnpm -C clients --filter @rtylr/portal typecheck` also generates it when missing).
- Preserve tenant boundaries, authorization, offline behavior and session handling. Client-side visibility is not server-side authorization. Surface failures honestly instead of showing them as empty or successful results.
- Keep user-facing text in the locale system. Update English, French and Arabic together, preserve interpolation and check RTL where relevant. Do not enable an unshipped locale because its strings exist.
- Preserve accessible names, keyboard behavior, focus, semantic markup and loading/error states when changing UI.

## Comments and documentation

- **No huge comments. Add only comments that are actually needed.** Prefer clear names and small functions.
- Explain a non-obvious reason, invariant, external constraint or dangerous edge case in a few lines. Do not narrate what the code already says.
- No essay-length headers, change histories, task transcripts, decorative banners or rejected approaches in source files. Put substantial rationale in a document and link it where useful.
- Update or remove stale comments when behavior changes. Use a TODO only for a concrete item with enough context to act on.
- Do not claim a check passed, a feature exists or a service is compliant without verification.

## Setup and commands

Use the pnpm version declared in `clients/package.json` and the lockfile in `clients/`. From the repository root:

```sh
pnpm -C clients install --frozen-lockfile
make -C voxie generate   # contracts are generated, not committed
git config core.hooksPath .githooks
```

Run package scripts with `pnpm -C clients --filter <package-name> <script>`. Script names differ between products, so read the package's `package.json` first.

```sh
# Portal and public entrypoints
pnpm -C clients --filter @rtylr/portal typecheck
pnpm -C clients --filter @rtylr/portal build

# Focused Portal/shared-package tests (pass real test-file paths)
clients/app/scripts/vitest.sh run clients/app/entrypoints/auth/src/features/register/Register.test.tsx

# Full Portal/shared-package suite, including its collected-file check
pnpm -C clients/app test

# Broader checks when the change warrants them
pnpm -C clients -r --filter @rtylr/portal typecheck:all
pnpm -C clients -r --filter './packages/*' typecheck

# Landing checks read exported HTML, so build first
pnpm -C clients --filter @rtylr/landing build
pnpm -C clients --filter @rtylr/landing test:unit

# Changed app transport boundaries
pnpm -C clients/app contracts:enforce
```

pnpm filters resolve relative to `clients/`, and some aggregate scripts under `clients/app` depend on that: verify what actually ran. **"No projects matched" or zero collected tests is not a passing check.**

## Testing that earns its cost

- Test behavior and realistic risks, not implementation trivia. Prefer a regression test that fails for the original bug and passes with the fix.
- Start with affected tests and relevant typechecks/builds. Widen to shared consumers or the full suite when a change crosses app boundaries, touches contracts/auth/payments/consent, or leaves real uncertainty. Complete required hooks and CI checks.
- Do not write tests that assert constants, file contents, class names, trivial getters or a copy of the implementation. Avoid broad snapshots and heavy mocking that let broken behavior pass.
- Documentation-only changes need a content, link and formatting review, not an app build. Copy or styling changes need the relevant locale, rendering or accessibility checks, not unrelated unit tests.
- Use real translations when testing rendered copy. Mock external boundaries, not the behavior under test.
- Use browser tests for browser-dependent behavior such as navigation, consent, persistence or focus. Use existing fixtures and block external/production traffic. Tests must never send real campaigns, payments, telemetry or customer mutations.
- POS native changes need the Rust/Tauri checks; a web build does not validate native behavior. Landing checks must exercise the static export where it matters.
- Do not rerun passing checks without a change or concern that justifies it. Report what ran, what failed and what was not checked, and separate baseline failures from regressions with evidence.

## Privacy, security and legal claims

- Keep secrets and customer data out of code, logs, fixtures and commits. Use synthetic test data and existing configuration mechanisms.
- Optional analytics must respect both acceptance and withdrawal. Do not enable session recording through a generic analytics choice or bypass shared consent handling.
- Consult `rtylrs/vault` (`rtylr/legal/README.md`) before changing public policies, data collection, vendors, retention, billing promises or security claims. Keep localized policies consistent with actual behavior.
- Do not invent an operating entity, jurisdiction, certification, audit result, signed agreement, deletion guarantee or transfer safeguard. Distinguish drafts and proposed controls from completed agreements and deployed controls.
- Preserve existing customer commitments and required notice/versioning when changing terms. A page edit does not mean customers received or accepted an amendment.
- Respect repository and dependency licenses. Do not relicense code or promise commercial/self-host rights without verified authorization.

## Delivery

- Review the final diff, format changed files and run `git diff --check`. Keep generated output, reports and local artifacts out of the commit unless deliberately required.
- Keep PRs focused. Describe the problem, the resulting behavior, meaningful validation and remaining limitations; omit task chronology and unsupported claims.
- **Screenshots are required in the PR for any change with a visible effect** (UI, copy, styling, layout, emails, docs pages). Show the affected states (before/after where useful, plus RTL or mobile when relevant) and embed them in the PR description. A PR with no visible change should say so.
- When a PR is requested, commit, push and open or update it, then verify its state and report its link. Do not stop at a local patch.
- Opening a PR does not authorize merging, deploying, publishing packages or notifying customers. Stay in the requested scope and read the relevant release workflow before taking those actions.
