# Audit prompt: shared packages (`@voxire/*`)

You are performing an independent code audit of the shared workspace packages used by every rtylr frontend app, and filing your findings as GitHub issues on **`tryrtylr/packages`**.

## Why this unit matters most

These packages are consumed by all ten apps. A defect in `@voxire/api` or `@voxire/auth` is a defect in every app at once, so the blast radius here is wider than any single app. Weight your severity accordingly: a fail-open permission check in `@voxire/permissions` is more serious than the same bug in one screen.

## Your unit

|                |                                                                                                                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Repo           | `tryrtylr/packages`, consumed by `tryrtylr/rtylr` as the `packages/` submodule                                                                                                                     |
| File issues on | `tryrtylr/packages`                                                                                                                                                                                |
| Packages       | `api`, `assets`, `auth`, `caret`, `components`, `config`, `consent`, `i18n`, `menu`, `observability`, `permissions`, `tailwind-config`, `types`, `ui`, `utils`, plus `voxie` as a nested submodule |

What each package owns:

- **`api`** Axios client, `safeFetch`, `unwrap`, `ResultAsync`, realtime, and the interceptors that convert camelCase to snake_case on the way out and back on the way in.
- **`auth`** `AppGuard`, `ProtectedRoute`, `RoleGuard`, `useUserStore`, cookie handling, `signOut`.
- **`config`** `APP_URLS`, `APP_CONFIG`, `AppFlagGuard`, runtime config, industries, themes.
- **`permissions`** RBAC helpers.
- **`ui`** Low-level primitives including `ErrorBoundary`.
- **`components`** Higher-level shared components.
- **`consent`** Cookie and consent management.
- **`observability`** Sentry and PostHog client wiring.
- **`tailwind-config`** The shared preset. The live design tokens are the `@theme` block in `theme.css`.
- **`voxie`** Nested submodule holding `.onk` contracts and committed generated `gen/ts`. **Do not audit voxie here.** It has its own prompt and its own repo.

## Before you read any code

The practical way to work on these packages is inside a consumer, so you can typecheck and test them against real usage:

```bash
git clone https://github.com/tryrtylr/rtylr.git
cd rtylr
git submodule update --init --recursive
pnpm install
```

The packages then live at `packages/`, which is a full checkout of `tryrtylr/packages`.

**Read this warning before you interpret any test result.** Because these packages are workspace dependencies, moving the submodule pointer leaves `node_modules` stale, and vitest then fails in bulk in ways that look exactly like real regressions. On one single commit this produced 0, 3, and 305 failures depending only on install state. If you see widespread failures, run `pnpm install` and re-run before believing any of it.

Baseline, so you can separate your findings from what was already there:

```bash
pnpm typecheck   # expect zero errors
pnpm test        # expect all passing
pnpm lint        # 2 pre-existing no-this-alias errors inside packages/, plus ~316 warnings
```

Those 2 `no-this-alias` errors are pre-existing and live in this repo. You may file **one** issue covering them if you can show they matter. Do not file them repeatedly.

## Discovery only, and why it matters here

You are running on a teammate's own machine, which may have uncommitted work in progress. Treat the checkout as read only.

**Allowed:** `git fetch --all --prune`, reading any file, running tests, running typecheck and build, read-only browser navigation.

**Forbidden:** `git pull`, `git checkout`, `git switch`, `git rebase`, `git merge`, `git stash`, creating branches, committing, and editing or deleting any file. Leave dirty worktrees exactly as you found them.

**Do not write real data.** Every app defaults to the production API (`https://api.voxire.com/`) when no `.env.local` is present. Do not click save, delete, send, or pay. Do not run flows that mutate.

## Two-phase protocol: report first, file second

**Do not create any issues during the audit.** Eight people are running audits in parallel, and an unreviewed batch from eight agents at once would flood the trackers.

**Phase 1, report.** When the audit is complete, post your **complete findings list as a single comment** on the tracking issue:

```
https://github.com/tryrtylr/rtylr/issues/685
```

Format that comment as a numbered list. For each candidate finding give: the proposed imperative title, the exact `file:line`, a one line failure scenario, the proposed severity, and the proposed labels. Also state your baseline results and anything you could not check. Then stop and wait.

**Phase 2, file.** Only after @AbedAmouneh replies approving that comment, create the issues. If the reply asks for changes, revise and post again. Never file without an explicit approval.

## Deduplicate before you file

This unit has **no prior audit coverage**, so you are the first pass over it. Still search open and closed issues by keyword before filing, since findings can arrive through normal development.

Search both states, by keyword, for every finding:

```bash
gh issue list --repo tryrtylr/packages --state all --limit 300 --search "<keyword>"
```

Roughly 350 audit issues already exist across these repos from the 2026-08-12 to 2026-08-16 rounds, and most are closed. If an existing issue covers your finding, **comment on it with your new evidence** rather than opening a second one. Re-check immediately before each create, not only once at the start, because other audits are filing while you work.

## Issue format

Titles are **imperative statements of the fix**, matching the existing convention. Not a description of the problem.

Good: `Persist public menu reviews instead of showing a fake success`
Good: `Scope the team query by business id`
Bad: `Reviews are broken`
Bad: `packages: validation issues`

Every issue carries these labels:

- Batch: `audit:packages-2026-08-16`
- Area: `area:packages`
- Severity: one of `severity:critical`, `severity:high`, `severity:medium`, `severity:low`
- Type: one of `type:bug`, `type:gap`, `type:debt`
- Topic, where it applies: `security`, `money`, `a11y`, `i18n`, `rtl`, `responsive`, `performance`, `correctness`, `test-coverage`
- Always `codex`

Create the batch label once before you start filing:

```bash
gh label create "audit:packages-2026-08-16" --repo tryrtylr/packages --color 0e8a16 --description "2026-08-16 packages audit" --force
```

**Assign every issue to yourself** with `--assignee "@me"`, so the follow-up fix has an owner.

```bash
gh issue create --repo tryrtylr/packages \
  --title "<imperative fix>" \
  --body-file body.md \
  --label "audit:packages-2026-08-16,area:packages,severity:high,type:bug,codex" \
  --assignee "@me"
```

Body, and fill in every section:

```markdown
## What is wrong

One or two sentences, plain language.

## Where

`path/to/file:123`

## How it fails

Given <specific input or state>, <specific wrong outcome>.

## Why it is wrong

The rule, contract, or standard violated. Quote it.

## Severity

Why this severity and not the one above or below it.

## Suggested direction

Optional, one sentence. Do not write the patch.
```

Write the body so a developer who has never seen your audit can act on it without asking you anything.

## Close the run with the structural issues

The previous rounds ended with these, and yours should too.

1. **`packages audit 2026-08-16: summary and trust assessment`** Labels `audit:packages-2026-08-16`, `area:packages`, `codex`. State your scope and guardrails, your baseline results, links to every issue you filed, and a plain answer to the question a reader actually has: **can this unit be trusted right now, and where can it not?**

2. **`packages audit 2026-08-16: unverified suspicions`** Labels `audit:packages-2026-08-16`, `type:gap`, `severity:low`, `codex`. Everything you suspected but could not confirm, and what would settle each one. This is where things that fail the evidence bar go, so that they are recorded without polluting the tracker with unproven claims.

3. **A backend handoff issue**, where your audit raised questions only the backend can answer. File it on `tryrtylr/onion` with `type:gap` and the appropriate severity, listing the endpoints or behaviours to verify. Skip this one if you have no backend questions.

## Cap

File at most **12** issues for this unit, not counting the structural issues above. If more clear the evidence bar, file the 12 highest severity and list the rest in the summary issue. **Do not pad to reach 12.** If only 4 findings are genuinely provable, file 4.

## What NOT to file

- Style, formatting, or naming preferences.
- Anything already covered by an open or closed issue, including the prior audit rounds.
- Speculation with no reproduction path. That belongs in the unverified suspicions issue.
- The known baseline noise listed above.
- Missing features the product has not built yet, as opposed to connections that are broken or half wired.

## Finally

Post one summary comment on your highest severity issue: how many you filed, how many candidates you rejected for failing the evidence bar, and what you could not check and why.
