# Blocked

Things this audit could not finish, the exact failure, and what would unblock
them.

## 1. Parallel research subagents unavailable

**What failed.** Four read-only research agents were dispatched at the start of
Phase 2 (identity resolution, cross-domain foreign keys, the `hr_category_id`
trace, and the existing cross-domain read audit). All four terminated early
with:

```
Agent terminated early due to an API error: You've hit your org's monthly
spend limit
```

**Impact.** None on scope, only on depth-per-hour. All four research tracks
were re-run inline in the main session using scripted extraction over
`docs/cross-app/inventory/contracts.csv` plus targeted reads, which is slower
and therefore samples fewer low-value fields by hand. Where a claim rests on
the scripted pass rather than a human read of the `.onk` file, the document
says so.

**What would unblock it.** A higher org spend limit (`/usage-credits`), then
re-dispatch the four prompts. They are reproducible: the research questions
are restated verbatim as headings in `02-entity-graph.md` and
`05-connections/README.md`.

## 2. No live verification of any endpoint

**What failed.** Nothing failed; this is a deliberate constraint from the
brief. Apps default to the production API (`https://api.voxire.dev/`) with no
local override, so no dev server was started and no request was issued. Every
statement about runtime behaviour in this audit is static analysis.

**Impact.** Claims of the form "this endpoint returns X" are contract-and-code
claims, not observed responses. The specific risk is an endpoint that exists in
Go but 404s in production because of router wiring this sweep did not execute,
or a field the handler leaves unset. Both are called out where they matter.

**What would unblock it.** A local onion instance with `API_URL` pointed at it
(see the note in the session memory about local runtime config sending apps
remote), then re-running the 67 unmatched paths from
`sweep-backend-modules.mjs` as real requests.

## 3. `<dynamic:...>` request paths not fully resolved

**What failed.** 13 of 743 request call sites build their path from a constant
or expression rather than a literal, so the sweep records
`<dynamic:API_ROUTES.GET.GET_ALL>` and similar. Six are `apps/menu` store
constants, one is HR's `AVAILABILITY_PATH`, and the rest are pass-through
wrappers in `packages/api`.

**Impact.** Marginal. All 13 are intra-domain calls in apps whose cross-app
role is already established by their literal paths. None of them changed a
verdict in `03-sharing-matrix.md`.

**What would unblock it.** Constant-folding in `sweep-endpoints.mjs`: resolve
imported identifiers to their declaration when the declaration is a string
literal in the same package.

## Round two additions

### 1. pablo (mobile app) not audited

`tryrtylr/pablo`, the mobile client, is a fourth frontend consumer of the same
API. It had no local clone; cloning and auditing an unfamiliar Rust/TypeScript
mobile codebase to this audit's citation standard did not fit the remaining
run. Recorded as uncovered in `00-repos.md`.

**What would unblock it.** Clone `tryrtylr/pablo` and run a request-sweep over
its networking layer, mapping calls to contract domains like
`sweep-endpoints.mjs` does.

### 2. onion sqlc generation required a local generate step

Generated sqlc code is gitignored, so the round-two worktree could not build
until `scripts/generate_sqlc.sh` ran locally (132 modules) after syncing the
schemas submodule pointer. This matches the brief's warning about fresh clones.
No action needed beyond documenting: any CI or reviewer building this branch
must run the generator first.

**What would unblock it.** Committing generated output is a repo policy change;
alternatively a CI job that generates before building.

### 3. A parallel working session shares two primary clones

Another session actively committed implementation branches in
`~/Development/onion` and `~/Development/rtylr` during this audit, switching
branch heads mid-run twice. All round-two work therefore lives in dedicated
worktrees (`~/Development/onion-worktrees/cross-app-round2`,
`~/Development/rtylr-worktrees/cross-app-round2`). Nothing was blocked; the
worktree isolation decision (R2-6/R2-8) kept both sessions running.

### 4. Live endpoint verification still not performed

Same constraint as round one: no dev server started, no requests issued against
`api.voxire.dev`. Every runtime claim remains static-analysis. The 67 unmatched
frontend paths from round one were reduced but not eliminated by better route
extraction; remaining unknowns are recorded in `01-inventory.md` coverage notes
and `backend_handlers.csv`.
