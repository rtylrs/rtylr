# Audit prompt: `voxie` contracts

You are performing an independent code audit of the shared API contracts, and filing your findings as GitHub issues on `tryrtylr/voxie`.

## What this repo is

Voxie owns the `.onk` contract definitions and the committed generated clients (`gen/go` and `gen/ts`) that both the backend and every frontend app depend on. It is the single source of truth for the shape of every API surface in the product.

That makes this the highest leverage and lowest volume audit of the set. A wrong field type here becomes a bug in Go and TypeScript simultaneously, and consumers pin their own submodule commits, so a mistake can persist in one consumer long after it is fixed here.

|             |                                                                                    |
| ----------- | ---------------------------------------------------------------------------------- |
| Repo        | `tryrtylr/voxie`, branch `main`                                                    |
| Contracts   | `api/<namespace>/<capability>/v1/*.onk`                                            |
| Generated   | `gen/go/` and `gen/ts/`, both committed                                            |
| Consumed by | `tryrtylr/onion` (Go), and `tryrtylr/rtylr` via `packages/voxie` (TypeScript)      |
| Namespaces  | booking, common, console, crm, erp, finance, flow, hr, hub, insights, pos, support |

## Measured facts to start from

These were measured against the current `main` of both consumers, and each is a candidate finding worth confirming and filing:

- **`booking`, `console`, and `support` have no consumer at all.** No frontend app and no shared package imports them. Determine whether they are planned surfaces, dead contracts, or consumed only by a repo outside this pair (the POS terminal lives in a separate repo, so check before concluding they are dead).
- **`apps/menu` and `apps/recruit` import zero generated types** and use hand-written interfaces instead, against the project rule that contract types must be imported rather than redeclared. Every request and response shape in those two apps can drift silently. Determine whether contracts exist for those surfaces at all: if they do not, that is a contract coverage gap and belongs here.
- Namespaces consumed today: common, crm, erp, finance, flow, hr, hub, insights, pos, validation.

## Before you read any code

```bash
git clone git@github.com:tryrtylr/voxie.git
cd voxie
pnpm install
pnpm typecheck    # expect clean
```

To audit against real consumers, also clone both:

```bash
git clone https://github.com/tryrtylr/rtylr.git && cd rtylr \
  && git submodule update --init --recursive && pnpm install
git clone https://github.com/tryrtylr/onion.git && cd onion \
  && git submodule update --init --recursive && make generate
```

Consumers intentionally pin different voxie commits, including older ones. A consumer being behind is **not** by itself a defect. It only matters if the drift causes a real inconsistency you can demonstrate.

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
gh issue list --repo tryrtylr/voxie --state all --limit 300 --search "<keyword>"
```

Roughly 350 audit issues already exist across these repos from the 2026-08-12 to 2026-08-16 rounds, and most are closed. If an existing issue covers your finding, **comment on it with your new evidence** rather than opening a second one. Re-check immediately before each create, not only once at the start, because other audits are filing while you work.

## Issue format

Titles are **imperative statements of the fix**, matching the existing convention. Not a description of the problem.

Good: `Persist public menu reviews instead of showing a fake success`
Good: `Scope the team query by business id`
Bad: `Reviews are broken`
Bad: `voxie: validation issues`

Every issue carries these labels:

- Batch: `audit:voxie-2026-08-16`
- Severity: one of `severity:critical`, `severity:high`, `severity:medium`, `severity:low`
- Type: one of `type:bug`, `type:gap`, `type:debt`
- Topic, where it applies: `security`, `money`, `a11y`, `i18n`, `rtl`, `responsive`, `performance`, `correctness`, `test-coverage`
- Always `codex`

Create the batch label once before you start filing:

```bash
gh label create "audit:voxie-2026-08-16" --repo tryrtylr/voxie --color 0e8a16 --description "2026-08-16 voxie audit" --force
```

**Assign every issue to yourself** with `--assignee "@me"`, so the follow-up fix has an owner.

```bash
gh issue create --repo tryrtylr/voxie \
  --title "<imperative fix>" \
  --body-file body.md \
  --label "audit:voxie-2026-08-16,severity:high,type:bug,codex" \
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

1. **`voxie audit 2026-08-16: summary and trust assessment`** Labels `audit:voxie-2026-08-16`, `codex`. State your scope and guardrails, your baseline results, links to every issue you filed, and a plain answer to the question a reader actually has: **can this unit be trusted right now, and where can it not?**

2. **`voxie audit 2026-08-16: unverified suspicions`** Labels `audit:voxie-2026-08-16`, `type:gap`, `severity:low`, `codex`. Everything you suspected but could not confirm, and what would settle each one. This is where things that fail the evidence bar go, so that they are recorded without polluting the tracker with unproven claims.

3. **A backend handoff issue**, where your audit raised questions only the backend can answer. File it on `tryrtylr/onion` with `type:gap` and the appropriate severity, listing the endpoints or behaviours to verify. Skip this one if you have no backend questions.

## Cap

File at most **10** issues for this unit, not counting the structural issues above. If more clear the evidence bar, file the 10 highest severity and list the rest in the summary issue. **Do not pad to reach 10.** If only 4 findings are genuinely provable, file 4.

## What NOT to file

- Style, formatting, or naming preferences.
- Anything already covered by an open or closed issue, including the prior audit rounds.
- Speculation with no reproduction path. That belongs in the unverified suspicions issue.
- The known baseline noise listed above.
- Missing features the product has not built yet, as opposed to connections that are broken or half wired.

## Finally

Post one summary comment on your highest severity issue: how many you filed, how many candidates you rejected for failing the evidence bar, and specifically what you concluded about the `booking`, `console`, and `support` namespaces and about contract coverage for the menu and recruit apps.
