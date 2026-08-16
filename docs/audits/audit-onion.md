# Audit prompt: `onion` backend

You are performing an independent code audit of the rtylr backend, and filing your findings as GitHub issues on `tryrtylr/onion`.

## The repo

|            |                                                                                                           |
| ---------- | --------------------------------------------------------------------------------------------------------- |
| Repo       | `tryrtylr/onion`, branch `main`                                                                           |
| Language   | Go 1.26                                                                                                   |
| Shape      | Modular monolith. 133 modules under `internal/modules/`, shared infrastructure under `internal/platform/` |
| Size       | Roughly 2000 Go files excluding generated code, and 630 test files                                        |
| Layering   | handler to service to repository. SQL belongs in the repository layer only                                |
| Submodules | `voxie` (the `.onk` contracts and generated `gen/go`), `internal/platform/database/schemas` (SQL schema)  |
| Database   | MySQL, accessed through sqlc-generated code                                                               |

`internal/platform/` holds cross-cutting infrastructure including `auth`, `cache`, `audit`, `database`, `email`, `emailqueue`, `financejournal`, `wsbus`, and `config`.

## Before you read any code

```bash
git clone https://github.com/tryrtylr/onion.git
cd onion
git submodule update --init --recursive
make generate     # REQUIRED: sqlc code is gitignored and must be generated
go build ./...
```

**`make generate` is not optional.** The sqlc output under `internal/modules/*/internal/infrastructure/persistence/sqlc/` is gitignored by design, so a fresh clone will not compile until you run it. If the build reports "no required module provides package .../persistence/sqlc", that is a missing codegen step and not a defect worth filing.

Baseline, so you can separate your findings from what is already there:

```bash
go build ./...    # expect clean
go vet ./...      # expect clean
make test         # see the known failure below
```

**Known pre-existing failure, which is not yours to file:** `TestInvoiceChildListsHaveServerBounds` fails with `expected two bounded invoice child lists, found 3`. A third invoice child-list query lacks server-side bounds. It is already known and already tracked. Everything else should pass. If you see other failures, re-check your setup before believing them.

Integration tests skip when `DB_HOST` is unset, so a plain `make test` run exercises unit tests only. If you can stand up a local MySQL you will cover far more, but **never point this at a production database**. If you cannot run integration tests, say so explicitly in your final summary rather than implying you covered them.

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
gh issue view 685 --repo tryrtylr/rtylr --comments | grep -iE "CLAIM|DONE|RELEASE" | grep -i "onion"
```

Also check whether this unit has already been filed:

```bash
gh issue list --repo tryrtylr/onion --label "audit:onion-2026-08-16" --state all --limit 50
```

**If either shows an active claim, or issues already filed under this batch label, stop now.** Do not audit this unit. Report back to whoever gave you this prompt and say the unit is already taken, naming who has it. Do not pick a different unit on your own.

A claim counts as released only if the same person later posted `RELEASE: onion`. A `DONE: onion` means the unit is finished, not free.

### Step 2: post your claim

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "CLAIM: onion | @<your-github-handle> | started $(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Use that exact format. The keyword, the unit name, your handle, and a UTC timestamp. Everything else in the comment is optional.

### Step 3: re-check for a race

Wait 60 seconds, then run the Step 1 check again. Two agents can claim within seconds of each other, and the timestamp is what breaks the tie.

**If someone else claimed this unit with an earlier timestamp, you yield.** Post:

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "RELEASE: onion | yielding to @<their-handle>, they claimed first"
```

Then stop and report back. If your timestamp is earlier, continue.

### Step 4: file directly, then close out

Once the claim holds, you file issues yourself as you confirm them. **No approval step is needed.** The evidence bar below is what keeps quality up, not a reviewer, so apply it strictly to your own findings.

When you have finished, post:

```bash
gh issue comment 685 --repo tryrtylr/rtylr \
  --body "DONE: onion | <N> issues filed | <links to the summary issue>"
```

If you abandon the run partway, post `RELEASE: onion | not completed, <reason>` so somebody else can pick it up. Never leave a claim hanging.

## Deduplicate before you file

This unit **has been audited before**. Search these labels specifically, because most of their issues are closed and a finding that looks new to you may already have been fixed:

- `audit:dash-backend-2026-08-12`
- `audit:crm-2026-08-13`
- `audit:insights-2026-08-16`

```bash
gh issue list --repo tryrtylr/onion --label "audit:dash-backend-2026-08-12" --state all --limit 200
gh issue list --repo tryrtylr/onion --label "audit:crm-2026-08-13" --state all --limit 200
gh issue list --repo tryrtylr/onion --label "audit:insights-2026-08-16" --state all --limit 200
```

Search both states, by keyword, for every finding:

```bash
gh issue list --repo tryrtylr/onion --state all --limit 300 --search "<keyword>"
```

Roughly 350 audit issues already exist across these repos from the 2026-08-12 to 2026-08-16 rounds, and most are closed. If an existing issue covers your finding, **comment on it with your new evidence** rather than opening a second one. Re-check immediately before each create, not only once at the start, because other audits are filing while you work.

## Issue format

Titles are **imperative statements of the fix**, matching the existing convention. Not a description of the problem.

Good: `Persist public menu reviews instead of showing a fake success`
Good: `Scope the team query by business id`
Bad: `Reviews are broken`
Bad: `onion: validation issues`

Every issue carries these labels:

- Batch: `audit:onion-2026-08-16`
- Severity: one of `severity:critical`, `severity:high`, `severity:medium`, `severity:low`
- Type: one of `type:bug`, `type:gap`, `type:debt`
- Topic, where it applies: `security`, `money`, `a11y`, `i18n`, `rtl`, `responsive`, `performance`, `correctness`, `test-coverage`
- Always `codex`

Create the batch label once before you start filing:

```bash
gh label create "audit:onion-2026-08-16" --repo tryrtylr/onion --color 0e8a16 --description "2026-08-16 onion audit" --force
```

**Assign every issue to yourself** with `--assignee "@me"`, so the follow-up fix has an owner.

```bash
gh issue create --repo tryrtylr/onion \
  --title "<imperative fix>" \
  --body-file body.md \
  --label "audit:onion-2026-08-16,severity:high,type:bug,codex" \
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

1. **`onion audit 2026-08-16: summary and trust assessment`** Labels `audit:onion-2026-08-16`, `codex`. State your scope and guardrails, your baseline results, links to every issue you filed, and a plain answer to the question a reader actually has: **can this unit be trusted right now, and where can it not?**

2. **`onion audit 2026-08-16: unverified suspicions`** Labels `audit:onion-2026-08-16`, `type:gap`, `severity:low`, `codex`. Everything you suspected but could not confirm, and what would settle each one. This is where things that fail the evidence bar go, so that they are recorded without polluting the tracker with unproven claims.

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

Post one summary comment on your highest severity issue: how many issues you filed, how many candidates you rejected for failing the evidence bar, **which of the 133 modules you actually audited**, and whether you were able to run integration tests. Being explicit about coverage gaps is more useful than implying you covered everything.
