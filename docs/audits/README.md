# Audit prompts, 2026-08-16 round

Self-contained prompts for running an independent code audit with Codex. Each file is one unit of work: hand it to the agent and it claims the unit, sets up, audits it, and files GitHub issues against the correct repository.

**Tracking and coordination: [rtylr#685](https://github.com/tryrtylr/rtylr/issues/685).** Claim your unit there before you start.

## How to run one

Give Codex **one file at a time**. Each prompt is fully self-contained, repeating the setup, guardrails, evidence bar and issue format rather than referring to a shared document, so nothing else needs to be pasted alongside it.

## Priority order

Nine units were already audited between 2026-08-12 and 2026-08-16, producing roughly 350 issues of which most are now closed. Those units will have a high duplicate rate on a second pass, so the order below reflects expected yield rather than importance.

### Priority 1: never audited

| Prompt                | Unit                                   | Files issues on     |
| --------------------- | -------------------------------------- | ------------------- |
| `audit-erp.md`        | `apps/erp` (port 5179)                 | `tryrtylr/rtylr`    |
| `audit-flow.md`       | `apps/flow` (5183)                     | `tryrtylr/rtylr`    |
| `audit-packages.md`   | The 15 shared `@voxire/*` packages     | `tryrtylr/packages` |
| `audit-rtylr-repo.md` | Monorepo cross-cutting concerns        | `tryrtylr/rtylr`    |
| `audit-voxie.md`      | `.onk` contracts and generated clients | `tryrtylr/voxie`    |
| `audit-vault.md`      | Documentation vault                    | `voxire/vault`      |

### Priority 2: backend

| Prompt           | Unit                    | Files issues on  |
| ---------------- | ----------------------- | ---------------- |
| `audit-onion.md` | Go backend, 133 modules | `tryrtylr/onion` |

Only three onion areas have prior coverage, all from the frontend side. The remaining 130 modules are untouched, so the prompt directs the agent to prioritise tenancy and money modules and to state which modules it did not reach.

### Priority 3: re-audit after this week's merges

| Prompt              | Unit                   | Prior coverage                                   |
| ------------------- | ---------------------- | ------------------------------------------------ |
| `audit-dash.md`     | `apps/dash` (5175)     | `audit:dash-2026-08-12`, `audit:dash-2026-08-13` |
| `audit-hr.md`       | `apps/hr` (5180)       | `audit:hr-2026-08-12`                            |
| `audit-auth.md`     | `apps/auth` (5176)     | `audit:auth-2026-08-13`                          |
| `audit-crm.md`      | `apps/crm` (5181)      | `audit:crm-2026-08-13`                           |
| `audit-finance.md`  | `apps/finance` (5182)  | `audit:finance-2026-08-14`                       |
| `audit-menu.md`     | `apps/menu` (5174)     | `audit:menu-2026-08-15`                          |
| `audit-recruit.md`  | `apps/recruit` (5184)  | `audit:recruit-2026-08-15`                       |
| `audit-insights.md` | `apps/insights` (5187) | `audit:insights-2026-08-16`                      |

Three of these units file **outside** `tryrtylr/rtylr`, and the vault sits under the `voxire` organisation rather than `tryrtylr`. `lynnsolh` and `mohamad-kasem-commits-crimes` are not collaborators on the vault, so whoever takes that unit needs access first.

## What every prompt enforces

**Discovery only.** Fetch and read freely, run tests, but never pull, checkout, switch branches, rebase, stash, commit or edit a file. Dirty worktrees are left exactly as found, which matters because these run on teammates' own machines where a checkout could clobber work in progress.

**A claim protocol, so two people never audit the same unit.** Before reading any code the agent checks [#685](https://github.com/tryrtylr/rtylr/issues/685) for an existing claim on its unit and checks whether that unit's batch label already has issues. If either shows the unit is taken, it stops and reports back rather than picking a different unit on its own. Otherwise it posts `CLAIM: <unit> | @handle | <UTC timestamp>`, waits 60 seconds, and re-checks: whoever holds the earlier timestamp keeps the unit, and the other yields with `RELEASE:`. It closes with `DONE: <unit> | <N> issues filed`, or `RELEASE:` if it abandons the run, so a claim is never left hanging.

Filing is direct, with no approval step. That puts the whole quality burden on the evidence bar below, which the prompts tell the agent to enforce on itself precisely because nobody is reviewing the batch before it lands.

**Five audit dimensions.** Correctness and silent failures; security, tenancy and RBAC; mobile-first, accessibility and i18n; architecture conformance and dead code; and missing cross-app connections.

**An evidence bar.** Every issue needs an exact `file:line`, a concrete failure scenario with specific inputs leading to a specific wrong outcome, and the quoted rule it violates. If the agent cannot write the failure scenario, it is told not to file. Unproven hunches go in the unverified suspicions issue instead of the tracker.

**Deduplication against prior rounds.** Each prompt names the prior audit labels for its unit and requires searching closed issues, not just open ones.

**Label conventions** matching the previous rounds: `audit:<unit>-2026-08-16`, `area:<x>`, `severity:critical|high|medium|low`, `type:bug|gap|debt`, topic labels (`security`, `money`, `a11y`, `i18n`, `rtl`, `responsive`, `performance`, `correctness`, `test-coverage`), and always `codex`. Issue titles are imperative statements of the fix. Every issue is self-assigned so the follow-up has an owner.

**A cap of 12 issues per unit** (10 for voxie), with explicit instruction not to pad.

**Three structural issues** close each run, matching the previous rounds: a summary with a trust assessment, an unverified suspicions issue, and a backend handoff issue where the audit raised questions only the backend can answer.

## Setup traps the prompts warn about

These cost real debugging time on this project, and one of them blocked a previous audit run outright:

- **Stale `node_modules` after a submodule move.** `packages/` is a submodule holding workspace dependencies. When its pointer moves, vitest fails in bulk in a way indistinguishable from real regressions. One commit of `main` produced 0, 3, and 305 failures purely from install state. Always `git submodule update --init --recursive && pnpm install` and re-run before believing a red suite.
- **Onion needs `make generate` first.** The sqlc output is gitignored by design, so a fresh clone will not compile until codegen runs.
- **Known baseline noise.** rtylr lint has 2 pre-existing `no-this-alias` errors inside `packages/`. Onion has one known failing test, `TestInvoiceChildListsHaveServerBounds`. Neither should be filed.
- **Apps default to the production API** when no `.env.local` is present, so the prompts forbid any flow that mutates data.

## Cross-app wiring, measured

The missing-connections dimension is grounded in what is actually wired today, measured against `main` on 2026-08-16:

- Only **auth, dash, menu, recruit** reference `APP_URLS`. **crm, erp, finance, flow, hr, insights have no cross-app navigation at all.**
- **`menu` and `recruit` import zero generated contract types** and hand-write their interfaces instead, which is the widest drift risk in the codebase since these are the only two customer-facing apps.
- **`booking`, `console` and `support` contracts have no consumer** in either the frontend or the shared packages.
- `flow` is the only app importing another domain's contracts (`hr`).
- **Dash and HR both ship roles and permissions admin screens**, which is flagged in both prompts as a potential security finding if they write divergent state.

Each prompt requires confirming that a backend contract exists before filing a missing connection, so the output is real gaps rather than a feature wish list.
