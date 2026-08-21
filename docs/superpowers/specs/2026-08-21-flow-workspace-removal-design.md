# Flow: remove the Workspace concept entirely

Date: 2026-08-21
Status: **approved, implementation not started**
Part of: workspace-removal initiative, sub-projects 2-4 of 4 (see
`2026-08-20-flow-business-role-bypass-design.md` for sub-project 1, already
shipped)

## Goal

Flatten Flow's data model and UI from Workspace > Project > Task to
Project > Task, with every project living directly on the business. When
this is done, `flow_workspaces`, `flow_workspace_members`, the workspace
switcher, and the `/w/:workspaceId/...` route prefix no longer exist
anywhere in the product.

## Decision already made: full v1 cutover, not a slow deprecation

Two design choices were on the table and both are settled:

1. **Contract shape**: a full `v2` cutover (new business-scoped contracts,
   no `workspace_id`) rather than an additive-only v1 extension. Chosen
   over the safer incremental option because the alternative left
   `workspace_id` sitting in the wire format indefinitely with no forcing
   function to ever finish the removal.
2. **Cutover timing**: v1 workspace-scoped endpoints get retired **in this
   initiative**, not left running alongside v2 forever - but "retired"
   means the frontend stops calling them before the backend deletes them,
   never the other way around. Nothing may break for a real user at any
   point in this sequence.

This is a bigger undertaking than "contracts + backend schema" alone - it
pulls a working slice of the old sub-project 4 (frontend surface removal)
into this same initiative, because leaving v1 endpoints callable with no
frontend caller left to call them is not a real cutover.

## Groundwork already shipped (prerequisites, both merged)

Two bugs were found and fixed while scoping this work, both now merged to
`onion` main:

- `EnsureMyTasksProject` wrote the caller's `userID` into the created
  project's `business_id` column instead of their resolved tenant id.
- `CreatePersonalWorkspace`'s INSERT never set `business_id` at all,
  leaving every personal workspace at the schema's `DEFAULT 0`.

Both are fixed going forward. `scripts/repair_flow_business_id.sql` (in
`onion`) holds the dry-run queries to find and fix rows written before the
fix - it has not been run for real against production yet. **Run and
review that script's SELECTs, then its UPDATEs, before starting Phase 1
below** - Phase 1's business-scoped list/create paths are only correct if
every row's `business_id` is trustworthy.

## What's already confirmed true about the current system

- `flow_projects.business_id` and `flow_workspaces.business_id` both
  already exist and are correctly populated on every _new_ row (as of the
  fixes above). The schema-level work for "projects live on the business"
  is already done; what's missing is everything that _reads_ by business
  instead of by workspace.
- `owner_id` on Project is the creating user, not the tenant - already a
  separate concern from `business_id`, no rename needed.
- Project-level membership (`flow_project_members`, with `admin`/`member`/
  `viewer` roles) is fully independent of workspace membership and is
  **not** going away. The per-project role gate (`RequireProjectRole`,
  shipped in onion PR #48, predates this initiative) keeps working
  unchanged throughout.
- Workspace-level membership (`flow_workspace_members`) is the only
  workspace-specific authorization path being removed. Sub-project 1's
  business-role bypass (owner/admin/operator get full access regardless of
  project or workspace role) already covers those three roles; anyone else
  who currently gets project access _only_ through workspace membership
  needs a real `flow_project_members` row before that fallback goes away
  (see Phase 1's backfill).
- "Personal workspace" (`is_personal = true`) is not a concept that needs
  replacing - once workspaces are gone, "My Tasks" is simply a personal
  project living on the user's business, exactly like any other project.
- Onion's `.onk`-based contract convention (see `voxie/CLAUDE.md`) treats
  a field rename, message rename, or removed field as a breaking change
  requiring a new `v2` directory - this initiative's `flow.project.v2` /
  `flow.task.v2` etc. follow that convention, not an ad-hoc one.

## Phase 1: contracts + backend, old endpoints re-implemented on the new model

**Deliverable: a shippable backend/contract change with zero frontend
changes required.** The current Flow frontend keeps working, unaware
anything happened underneath it.

### Contracts (`voxie`)

For each domain that currently carries `workspace_id` (project, task,
cycle, milestone, view, label, dashboard, setting - see the file list
below), add a `v2` directory:

- `Project` (v2): drops `workspace_id` entirely. Everything else
  unchanged.
- `ListProjectsRequest` (v2): no `workspace_id` field at all - the
  business is resolved from the caller's auth context, the same way every
  other business-scoped list endpoint in this codebase already works
  (see `hr`/`crm`/`finance` domains for the existing pattern - none of them
  take a business id on the wire, it's always resolved server-side).
- `CreateProjectRequest` (v2): `workspace_id` removed; nothing replaces it
  - the business is implicit from auth.
- Task/Cycle/Milestone/View/Label/Dashboard/Setting v2 messages: same
  treatment - drop `workspace_id`, keep every other field name and type
  unchanged so the Go/TS diff is minimal and mechanical.
- v1 messages and the v1 service directories stay exactly as they are.
  Nothing is deleted from `voxie` in this phase.

### Backend (`onion`)

- Add v2 API handlers (`ProjectAPIv2`, `TaskAPIv2`, etc. - or extend the
  existing `*API` structs with new methods per the codebase's own
  convention, whichever the implementer finds keeps the diff smaller) that
  read/write purely by `business_id`, no workspace concept touched at all.
- **Re-implement the existing v1 handlers to derive their workspace-shaped
  response from the same business-primary data**, instead of `workspace_id`
  being the source of truth. Concretely: `ListProjects` v1 keeps accepting
  a `workspace_id` on the wire, but internally it becomes "list this
  business's projects, then filter to the ones whose stored `workspace_id`
  matches" rather than a workspace-indexed query. This is what makes phase
  1 safe to ship with zero frontend changes - the response shape and
  values a v1 caller sees do not change.
- Workspace-membership backfill (**run once, before any code in this phase
  that starts relying on project membership being complete**):
  1. Write the migration: for every `flow_workspace_members` row, if the
     user's business role is not owner/admin/operator AND they don't
     already have a `flow_project_members` row for a given project in
     that workspace, insert one with an equivalent role.
  2. Run it in dry-run mode first (`SELECT` the same set, output the
     affected user/project count and list - no `INSERT`).
  3. A human reviews that output.
  4. Only then, run it for real.
  5. This must complete and be verified before Phase 4 retires the
     workspace-membership fallback in `RequireProjectRole` - if it hasn't
     run, retiring that fallback silently revokes access for anyone who
     was relying on it.

### Testing

- Every v2 handler gets the same unit-test coverage pattern already
  established for v1 (`RequireProjectRole`/`RequireTaskRole` with a mocked
  business-role resolver, per sub-project 1's own test pattern).
- A regression test proving v1's response is byte-for-byte unchanged
  before and after the internal re-implementation, for at least
  `ListProjects` and `GetFlowDashboard` (the two most complex read paths).
- The backfill migration's dry-run output is reviewed by a human before
  the real run - not a code review, an actual "does this list of affected
  users look right" check.

### Rollback

Phase 1 introduces new code paths without removing anything old, so
rollback is a normal revert - no data migration to undo except the
membership backfill, which is additive (new `flow_project_members` rows)
and safe to leave in place even if the rest of the phase is reverted.

## Phase 2: frontend data layer swaps to v2, UI unchanged

**Deliverable: the frontend stops calling v1 endpoints, but no screen
looks different.** Every hook/API call that currently threads
`workspaceId` through gets swapped to the v2, business-scoped call
instead. From the audit done during brainstorming, this touches:

`useProjects.ts`, `useCurrentProject.ts`, `useColumns.ts`, `useTasks.ts`,
`useLabels.ts`, `useViews.ts`, `useCycles.ts`, `useMilestones.ts`,
`useWorkDay.ts`, `useInsights.ts`, `useDashboard.ts` (workspace-dashboard's
own hook - see Phase 3 for what happens to the _feature_ it backs),
`shared/services/flowApi.ts`, `shared/lib/currentUser.ts` (the
workspace-role fallback in `useCanEditTask`/`useCanDeleteResource` becomes
dead code and gets deleted here, not carried into Phase 3).

No route changes, no component removal - `ProjectList.tsx`,
`ProjectBoard.tsx`, etc. keep rendering exactly what they render today,
just fed by a v2 call instead of a v1 one.

### Testing

Existing component/hook tests should pass unchanged if a screen's visible
behavior is genuinely unchanged - that's the acceptance bar for this phase.
Any test that had to change because it was asserting on `workspaceId`
plumbing (not visible behavior) is expected and fine; any test that had to
change because the screen now _looks_ different is a sign this phase
overstepped its own scope.

## Phase 3: frontend surface removal

**Deliverable: the workspace concept is gone from the UI.**

- Delete `features/workspace-manager/` (`WorkspaceManager.tsx`,
  `AddWorkspaceMemberDialog.tsx`, `useWorkspaceManager.ts`) and
  `features/workspace-dashboard/` (`WorkspaceDashboard.tsx`,
  `useWorkspaces.ts`) entirely - no replacement, these features cease to
  exist.
- Remove the workspace switcher from `FlowLayout.tsx`.
- Route changes in `router.tsx`/`routes/AppRoutes.tsx`: `/w/:workspaceId`
  and `/w/:workspaceId/p/:projectId` become `/p/:projectId` (project
  routes lose the workspace prefix entirely - there is no business-level
  index route needed since the project list itself is the top-level view).
  Add a redirect from the old `/w/:workspaceId/p/:projectId` shape to the
  new `/p/:projectId` shape for any bookmarked/shared links still in the
  wild, at least for one release cycle.
- `CommandBar.tsx`/`QuickCreate.tsx`: remove any workspace-picker UI:
  creating something no longer involves choosing a workspace.
- `shared/lib/personalWorkspace.ts` and `PERSONAL_WORKSPACE_ID` sentinel:
  delete - nothing needs a workspace sentinel once "My Tasks" is just a
  project.
- `shared/stores/flowStore.ts`: remove `activeWorkspaceId` and any
  workspace-scoped state.
- `shared/lib/useLayoutRealtime.ts`: remove the `workspace:<id>` realtime
  topic subscription; keep `user:<id>` and whatever project-level
  subscriptions already exist.

### Testing

Full manual click-through of every route Flow exposes (per this app's own
`CLAUDE.md`, there is no live browser verification infrastructure that
doesn't require a running auth app and a real login - budget time for
that rather than skipping it). Automated test suite should be green with
zero `workspace`-referencing test files remaining in `apps/flow`.

## Phase 4: backend cleanup

**Deliverable: workspaces no longer exist anywhere, including the
database.**

- Confirm (via logs/metrics, not assumption) that v1 workspace-scoped
  endpoints have received zero traffic for a full release cycle after
  Phase 3 ships - this is the actual gate for this phase, not a calendar
  date.
- Delete the v1 API handlers, v1 `.onk` contracts (or leave them
  `Deprecated` per `voxie`'s own field-removal process if anything
  external still plausibly references them - judgment call at
  implementation time).
- Drop `flow_workspace_members` and the workspace-membership fallback in
  `RequireProjectRole`/`RequireTaskRole`/`RequireWorkspaceRole` - safe only
  because Phase 1's backfill already gave every real workspace member an
  equivalent project-membership row.
- Drop `flow_workspaces` and `workspace_id` columns on every table that
  carries one (`flow_projects`, `flow_project_columns`, `flow_tasks`,
  `flow_cycles`, `flow_milestones`, `flow_views`, `flow_work_types`,
  `flow_labels`, `flow_notifications` - see the column audit in this doc's
  history for the full list).

### Rollback

This phase is the only genuinely hard-to-reverse one (dropped columns,
dropped tables). Do not start it without a fresh, verified database
backup, and do not run it against production without the same explicit
go-ahead every other prod-data-touching action in this codebase requires.

## Sequencing and independence

Each phase ships as its own reviewable, independently-working change -
matching this repo's own convention of one PR (or a small tightly-scoped
batch) per unit of work. Phase 2 cannot start until Phase 1's backfill is
verified. Phase 3 cannot start until Phase 2 is fully merged (a screen
still calling v1 mid-way through Phase 3's route changes would break).
Phase 4 cannot start until Phase 3 has been live for a full release cycle
with confirmed-zero v1 traffic.
