# Audit prompt: `apps/insights`

You are performing an independent code audit of a single app inside the rtylr frontend monorepo, and filing your findings as GitHub issues on `tryrtylr/rtylr`.

## Product context

RTYLR is a multi-tenant commerce operating system sold to restaurants, agencies, law firms, retail groups and other businesses. Every business shares the same code, so behaviour that differs between customer types belongs in configuration rather than in the source.

## Your unit

|                |                                                                                                            |
| -------------- | ---------------------------------------------------------------------------------------------------------- |
| Path           | `apps/insights`                                                                                            |
| Purpose        | Analytics deep dive across revenue, customers, products, financial, operations, plus reports and settings. |
| Dev port       | 5187                                                                                                       |
| Route count    | 10 route definitions in `src/router.tsx`                                                                   |
| Components     | 17 `.tsx` files, 59 `.ts` files                                                                            |
| Tests          | 47 test files                                                                                              |
| Locales        | `src/locales/{en,ar,fr}/`                                                                                  |
| Contracts used | insights                                                                                                   |

Key routes: `/`, `/revenue`, `/customers`, `/products`, `/financial`, `/operations`, `/reports`, `/settings`

Audit **only** `apps/insights`, plus any shared code it depends on where the defect is genuinely caused by this app's usage. Findings that live inside `packages/` belong to the packages audit, not this one.

## Before you read any code

Clone and set up the repo. Run these in order and do not skip the install:

```bash
git clone https://github.com/tryrtylr/rtylr.git
cd rtylr
git submodule update --init --recursive
pnpm install
```

**Read this warning before you interpret any test result.** `packages/` is a git submodule whose contents are workspace dependencies. Whenever that submodule pointer moves, `node_modules` goes stale and vitest fails in bulk in ways that look exactly like real regressions. On one single commit of `main` this produced 0 failures, 3 failures, and 305 failures depending only on install state. If you see widespread failures, run `git submodule update --init --recursive && pnpm install` and re-run before you believe any of it. Never file an issue based on a red suite you have not re-run after a clean install.

Establish the baseline before auditing, so you can tell your findings apart from what was already there:

```bash
pnpm typecheck   # expect zero errors
pnpm test        # expect all passing
```

Known pre-existing baseline noise, which is **not** yours to file:

- `pnpm lint` reports 2 `no-this-alias` errors inside `packages/`, plus roughly 316 warnings. Judge a change by whether it adds errors beyond those two.
- The `tailwind.config.ts` files are documentation only. Tailwind v4 does not read them. Live tokens are the `@theme` block in `packages/tailwind-config/theme.css`.

## Discovery only, and why it matters here

You are running on a teammate's own machine, which may have uncommitted work in progress. Treat the checkout as read only.

**Allowed:** `git fetch --all --prune`, reading any file, running tests, running typecheck and build, read-only browser navigation.

**Forbidden:** `git pull`, `git checkout`, `git switch`, `git rebase`, `git merge`, `git stash`, creating branches, committing, and editing or deleting any file. Leave dirty worktrees exactly as you found them.

**Do not write real data.** Every app defaults to the production API (`https://api.voxire.com/`) when no `.env.local` is present. Do not click save, delete, send, or pay. Do not run flows that mutate.

## Claim this unit first, and confirm nobody else has it

Fifteen units are being audited by eight people at once. **Two people auditing the same unit produces duplicate issues**, which costs real triage time and has already happened on this project. The claim below is the only thing preventing that, so do not skip it.

### Step 1: check for an existing claim

Before you read a single file:

```bash
gh issue view 685 --repo tryrtylr/rtylr --comments | grep -iE "CLAIM|DONE|RELEASE" | grep -i "insights"
```

Also check whether this unit has already been filed:

```bash
gh issue list --repo tryrtylr/rtylr --label "audit:insights-2026-08-16" --state all --limit 50
```

**If either shows an active claim, or issues already filed under this batch label, stop now.** Do not audit this unit. Report back to whoever gave you this prompt and say the unit is already taken, naming who has it. Do not pick a different unit on your own.

A claim counts as released only if the same person later posted `RELEASE: insights`. A `DONE: insights` means the unit is finished, not free.

### Step 2: post your claim

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "CLAIM: insights | @<your-github-handle> | started $(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Use that exact format. The keyword, the unit name, your handle, and a UTC timestamp. Everything else in the comment is optional.

### Step 3: re-check for a race

Wait 60 seconds, then run the Step 1 check again. Two agents can claim within seconds of each other, and the timestamp is what breaks the tie.

**If someone else claimed this unit with an earlier timestamp, you yield.** Post:

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "RELEASE: insights | yielding to @<their-handle>, they claimed first"
```

Then stop and report back. If your timestamp is earlier, continue.

### Step 4: file directly, then close out

Once the claim holds, you file issues yourself as you confirm them. **No approval step is needed.** The evidence bar below is what keeps quality up, not a reviewer, so apply it strictly to your own findings.

When you have finished, post:

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "DONE: insights | <N> issues filed | <links to the summary issue>"
```

If you abandon the run partway, post `RELEASE: insights | not completed, <reason>` so somebody else can pick it up. Never leave a claim hanging.

## Deduplicate before you file

This unit **has been audited before**. Search these labels specifically, because most of their issues are closed and a finding that looks new to you may already have been fixed:

- `audit:insights-2026-08-16`

```bash
gh issue list --repo tryrtylr/rtylr --label "audit:insights-2026-08-16" --state all --limit 200
```

Search both states, by keyword, for every finding:

```bash
gh issue list --repo tryrtylr/rtylr --state all --limit 300 --search "<keyword>"
```

Roughly 350 audit issues already exist across these repos from the 2026-08-12 to 2026-08-16 rounds, and most are closed. If an existing issue covers your finding, **comment on it with your new evidence** rather than opening a second one. Re-check immediately before each create, not only once at the start, because other audits are filing while you work.

## Issue format

Titles are **imperative statements of the fix**, matching the existing convention. Not a description of the problem.

Good: `Persist public menu reviews instead of showing a fake success`
Good: `Scope the team query by business id`
Bad: `Reviews are broken`
Bad: `insights: validation issues`

Every issue carries these labels:

- Batch: `audit:insights-2026-08-16`
- Area: `area:insights`
- Severity: one of `severity:critical`, `severity:high`, `severity:medium`, `severity:low`
- Type: one of `type:bug`, `type:gap`, `type:debt`
- Topic, where it applies: `security`, `money`, `a11y`, `i18n`, `rtl`, `responsive`, `performance`, `correctness`, `test-coverage`
- Always `codex`

Create the batch label once before you start filing:

```bash
gh label create "audit:insights-2026-08-16" --repo tryrtylr/rtylr --color 0e8a16 --description "2026-08-16 insights audit" --force
```

**Assign every issue to yourself** with `--assignee "@me"`, so the follow-up fix has an owner.

```bash
gh issue create --repo tryrtylr/rtylr \
  --title "<imperative fix>" \
  --body-file body.md \
  --label "audit:insights-2026-08-16,area:insights,severity:high,type:bug,codex" \
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

1. **`insights audit 2026-08-16: summary and trust assessment`** Labels `audit:insights-2026-08-16`, `area:insights`, `codex`. State your scope and guardrails, your baseline results, links to every issue you filed, and a plain answer to the question a reader actually has: **can this unit be trusted right now, and where can it not?**

2. **`insights audit 2026-08-16: unverified suspicions`** Labels `audit:insights-2026-08-16`, `type:gap`, `severity:low`, `codex`. Everything you suspected but could not confirm, and what would settle each one. This is where things that fail the evidence bar go, so that they are recorded without polluting the tracker with unproven claims.

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

When you are done, post one short summary comment on the highest severity issue you filed, listing: how many issues you filed, how many candidate findings you rejected for failing the evidence bar, and anything you could not check and why. Being explicit about what you did not cover is more useful than implying full coverage.
