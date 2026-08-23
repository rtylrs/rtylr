# Flow: workspace removal, replacement model and migration design

Date: 2026-08-21 (consolidated 2026-08-22 after the feature-by-feature review session)
Status: **decisions locked; implementation not started**
Companion to `2026-08-21-flow-workspace-removal-design.md` (the approved four-phase
baseline). This document supersedes it wherever the two disagree, records every ruling
from the review session, and is the single reference implementation PRs link to.
Every claim about current behavior was verified against onion (`internal/modules/workspaces`),
voxie (`api/flow/**`) and apps/flow/src, not assumed.

Team-facing mirror: `vault/rtylr/apps/flow/workspace-removal.md` carries the condensed
ledger and the live phase tracker.

## Part 0: executive summary

Flow loses the Workspace layer entirely: data model becomes Business > Project > Task,
with the business resolved server-side from auth and never present on the wire.
Projects become private-by-default with an optional public listing mode; access is
creator plus explicit members, with department managers gaining visibility through a
config-driven policy and business owner/admin/operator keeping unconditional access
(the shipped hub-role bypass). Deleting becomes a recycle bin with configurable
retention. Custom task statuses ship as an immediate fast-follow additive initiative,
not part of the breaking cutover. Nothing breaks for a real user at any phase; prod
currently has no clients, which lowers stakes but does not change sequencing discipline.

## Part 1: decision ledger

Rulings from the 2026-08-21/22 review sessions, grouped by feature. IDs are stable;
PRs cite them.

### Projects

| ID  | Decision                                                                                                                                                                                                                                      |
| --- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P1  | Sidebar project list stays flat, sorted as today. Grouping may become a per-business config later only if a proven need emerges.                                                                                                              |
| P2  | Archive becomes a recycle bin: archived projects land there, retention is a per-business config key (default 30 days), and onion's worker purges expired bins.                                                                                |
| P3  | The archive confirmation dialog must state the scheduled permanent-deletion date explicitly. Bin rows show time remaining. Purge-now requires an explicit confirmation modal naming the cascade scope.                                        |
| P4  | Restore from bin: creator and business owner/admin only. Early purge: same set.                                                                                                                                                               |
| P5  | Anyone may create projects (API already permits this; D3 made it explicit and documented in v2 handlers).                                                                                                                                     |
| P6  | Projects gain `visibility`: `private` (default) or `public`. Public means metadata-only discoverability: name, icon, color, description, progress render on the owner's profile card grid; opening tasks still requires membership or bypass. |
| P7  | Visibility flips are creator + project admin, available at creation (dialog toggle) and later in Project Settings.                                                                                                                            |
| P8  | Sidebars are curated for everyone, bypass roles included: created + member projects only. Bypass grants ACCESS everywhere, never listing ("access yes, listed no").                                                                           |
| P9  | Department managers see proxied projects in a dedicated "Your reports" sidebar group (see M3). Conscious asymmetry with P8: proxy is a targeted relationship, bypass is god-mode and stays unlisted to avoid enormous lists.                  |
| P10 | Existing projects migrate with `visibility = 'private'`, preserving their backfilled circles exactly. Only new projects exercise public.                                                                                                      |

### My Tasks

| ID  | Decision                                                                                                                                                                                                                                                                  |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| M1  | Quick-add needs only a title (plus optional fields): the silent first-label stamping is removed and labels become optional everywhere. Fixes the day-one dead end where a new business (zero labels) cannot add its first task.                                           |
| M2  | The My Tasks bucket project is always private, exempt from the manager-proxy rule, unless its owner explicitly grants access. Identified exactly by the new `is_personal` column, deleting the frontend name/icon heuristics ("My Tasks"+checkmark, "Inbox"+inbox emoji). |
| M3  | Dept-manager proxy: managers can see projects of people in departments they manage. Implemented as a config-driven resolver rule; default ON. Real person-to-person reporting line is fast-follow B.                                                                      |
| M4  | Papercuts approved: kill bucket heuristics (M2), delete dead files (`urlUtils.ts` builders matching no live route, unmounted `routes/AppRoutes.tsx`), fix cold deep links via GetProject (C9).                                                                            |

### Tasks

| ID  | Decision                                                                                                                                                                                     |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T1  | Assigning a user to a task auto-grants them project membership (member role) if absent. Closes the invisible-assignment trap created by private-by-default.                                  |
| T2  | Task dependencies stay cross-project within the business.                                                                                                                                    |
| T3  | Full custom statuses approved as a dedicated fast-follow initiative (F-A below): per-project status sets, additive contracts, semantic buckets for rollups. NOT bundled into the v2 cutover. |
| T4  | Board columns remain independent of statuses (today's decoupling preserved); F-A revisits only if wanted.                                                                                    |

### Board and interaction

| ID  | Decision                                                                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| B1  | Drag-and-drop moves from HTML5 events to Pointer Events during Phase 3 (touch support).                                                                             |
| B2  | Per WCAG 2.2 AA, dragging gains a single-pointer alternative: a Move-to-column action per task card plus keyboard paths. Drag is never the only way to move a task. |

### Content scoping

| ID  | Decision                                                                                                                                                                                               |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| S1  | Labels re-scope to one business-wide pool; duplicate names tolerated indefinitely, dedupe deferred until someone complains (D2).                                                                       |
| S2  | Saved views stay shared business-wide at migration; a per-view visibility toggle is reserved for later.                                                                                                |
| S3  | Cycles become project-scoped going forward. Existing workspace-level cycles migrate onto the project they filtered, or archive if none.                                                                |
| S4  | Milestones drop the redundant nullable workspace column; board columns keep theirs until Phase 4 drops it like every table.                                                                            |
| S5  | Search (and CommandBar) goes business-wide and MUST filter results by caller visibility; private projects and tasks are invisible to non-members. Forced consequence of P6, recorded as a requirement. |

### Analytics

| ID  | Decision                                                                                                                                                                                                                    |
| --- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A1  | Numeric aggregates include all projects, private included, for everyone: numbers stay honest. Roadmap names and drill-downs show everything to hub owner/admin/operator; regular members see only projects they can access. |

### Work-time

| ID  | Decision                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| W1  | Team view roster becomes the business team (via hr_employees, already business-keyed). Who-sees-whose-days is a per-business config policy resolved server-side per (viewer, target) pair. Default policy: hub owner/admin/operator see all; flow project admins see targets sharing a project with them ("responsible for"); others denied. Editable in Flow Settings (U6). "Team leader" maps to project admin today; a first-class role may join later without schema change. |

### Platform surfaces

| ID  | Decision                                                                                                                                                                                                                                                                                                                                 |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| X1  | Realtime: `workspace:<id>` topic dies; project lifecycle events publish on a new `business:<id>` topic authorized as "caller belongs to business". Layout subscribes business + user topics. While the authorizer is open, hub-profile bypass is added to topic auth, closing the gap where owners had REST access but no socket access. |
| X2  | Caret AI context swaps `workspace_id` for `business_id`; onion's `tool_executor.go` (workspace-scoped AI queries, personal sentinel default) re-points to business/project scoping in Phase 1.                                                                                                                                           |
| X3  | Notifications unchanged; nullable `workspace_id` metadata column dropped in Phase 4.                                                                                                                                                                                                                                                     |
| X4  | Preview stub routes (project overview/timeline/dashboard, timeline, roadmap, insights, team behind APP_FLOW_PREVIEW_ROUTES) survive the migration re-pathed under `/p/:projectId/...`, still flag-OFF.                                                                                                                                   |
| X5  | Old `/w/:workspaceId/p/:projectId` URLs redirect to `/p/:projectId` for one release cycle; `/w/:workspaceId` lands redirect to `/my-issues`; catch-all follows `/my-issues`. `/task/:id` links never change; verified no cross-app deep links into `/w/...`.                                                                             |
| X6  | Keyboard chord `g d` (jump to dashboard) retires with its destination.                                                                                                                                                                                                                                                                   |

## Part 2: permission and visibility model

| Who                           | Can open                                                 | Sidebar                                                     |
| ----------------------------- | -------------------------------------------------------- | ----------------------------------------------------------- |
| Business owner/admin/operator | Everything, always (shipped bypass)                      | Created + member only (P8)                                  |
| Dept manager (proxy, M3)      | Projects of dept members per config policy               | Those projects additionally under "Your reports" group (P9) |
| Project creator               | Their project, implicit admin                            | Listed                                                      |
| Project member                | Per role: viewer reads, member edits, admin manages      | Listed                                                      |
| New assignee                  | Auto-member on assignment (T1)                           | Listed                                                      |
| Everyone else                 | Private: nothing. Public: metadata on profiles only (P6) | Not listed                                                  |
| My Tasks bucket               | Owner only, proxy-exempt (M2)                            | Hidden entirely                                             |

Cross-cutting: search filters by this entire model (S5); realtime topic auth = membership OR bypass (X1); work-time visibility runs its own config policy (W1). Assignment, comments, mentions inside inaccessible projects cannot reach non-members because every write path already gates RequireProjectRole.

## Part 3: UI inventory (new and changed surfaces)

1. Navigation model: landing stays /my-issues; sidebar carries My Tasks link, Projects group (created+member), Your reports group (proxy, when non-empty), All projects entry, existing feature links, Recycle bin section, Settings.
2. All projects page (new): searchable, filterable (status, visibility, mine/all-visible) grid or list of every project the caller can reach, with create CTA and empty/no-results states per U-standards.
3. Profile page (new): route `/u/:userId`; avatar, name, role/title, bio where available, and a card grid of that user's PUBLIC projects (metadata only). Linked from task assignees, comment authors, member pickers, workload rows.
4. Recycle bin (new): collapsed section in the sidebar project area for browse + restore; Purge-now and Empty-bin controls live in Flow Settings, admin-gated, double-confirmed (P3/P4).
5. Visibility controls (new): Private/Public toggle in the create dialog (default Private) and in Project Settings (P7).
6. Flow Settings panels (new): Recycle bin retention days and Work-time visibility policy rendered as labeled read/write controls for business admins, persisting through business config (U6).
7. Removed UI: workspace switcher, WorkspaceManager members screen, WorkspaceDashboard, move-project dialog, workspace pickers in CommandBar/QuickCreate, `g d` chord.
8. Empty states: brand-new business sees actionable empties (create-first-project CTA, quick-add focus) rather than blank panes, per U-standards.

## Part 4: UI standards addendum (ui-ux-pro-max sourced)

Severity ratings come from the skill's guideline dataset; these bind all new Flow UI.

1. Confirm before irreversible (High): purge-now and empty-bin get explicit modals naming cascade scope; archive states the deletion date (P3).
2. Single-pointer alternative to drag (High, WCAG 2.2 AA): B2's move-menu and keyboard paths are mandatory, not nice-to-have.
3. No dead-end results (Medium): search/filter surfaces suggest alternatives instead of bare zero-states.
4. Empty states guide (Medium): message plus next action everywhere content can be empty.
5. Labeled forms (High): settings panels and dialogs use real label associations; submit shows loading then success/error (High).
6. Control semantics (Critical): chips and icon-only buttons carry native roles, accessible names, exposed state (aria-pressed on toggles/filters, aria-label on restore/purge icons); actions hidden solely behind hover are prohibited.
7. Repo standing rules continue to apply: min-width mobile-first, 44px targets, 16px inputs, --flow-\* tokens only, t() for every string in en/ar/fr.

## Part 5: contracts (voxie v2)

Shape decisions:

- Domains touched: project, task, dashboard, insights, label, cycle, milestone, view, time. Notification and setting carry no workspace fields and stay v1. There is deliberately no v2 workspace domain.
- All ids normalize to int64 (onekit emits JSON strings on the wire), ending the int64/string drift; requiredness made explicit; the lost exactly-one-of constraint on ListTasksRequest dissolves into an optional project_id filter.
- `CreateTaskRequest.project_id` stays OPTIONAL (R13): omitted means file into the caller's My Tasks bucket via re-keyed EnsureMyTasksProject(user_id, business_id). Invariant stands: every task belongs to exactly one project; "outside a project" is a UI concept. Response continues echoing placement.
- `Project` gains `visibility` and `is_personal`; stale owner_id comment fixed (it is the creating user, not the tenant).
- Time domain: `ListWorkDaysForWorkspace` becomes `ListTeamWorkDays` over the business roster, gated by the W1 policy resolver.

New/enumerated RPCs (names indicative, settled at implementation):

- `GetProject(id)` (C9; enables cold deep links)
- `SetProjectVisibility(id, visibility)`
- `ListProjects` (business-scoped; filters: status, visibility, query; paginated) doubles as the All projects page feed
- Bin: `ListArchivedProjects`, `RestoreProject`, `PurgeProject` (admin-gated immediate purge); retention expiry handled by worker, not RPC
- `GetProfile(userId)` returning profile basics + public projects metadata
- Dashboard/insights/roadmap/team-workload lose their workspace_id parameters entirely
- `MoveProject` has no v2 successor (R5)

v1 endpoints remain served, re-implemented internally over business-primary data, until Phase 4 retirement; byte-for-byte regression tests required on ListProjects and GetFlowDashboard.

## Part 6: data model changes

Schema additions (Phase 1):

- `flow_projects.visibility ENUM('private','public') DEFAULT 'private'`
- `flow_projects.is_personal BOOLEAN DEFAULT FALSE`
- `flow_projects.archived_at TIMESTAMP NULL` and `purge_at TIMESTAMP NULL`
- `purge_at` stamps at archive time from the then-current retention config; later config changes affect future archives only. Live recalculation explicitly rejected for now.

Backfills and repairs:

1. Repair script run for real (personal-workspace business_id bugs), SELECTs reviewed first.
2. Script extended with dry-run counts: non-personal projects with business_id=0; tasks with business_id=0; tasks with project_id=0 (legacy orphans needing re-homing before Phase 4).
3. Membership backfill: workspace members without project rows gain equivalent roles (admin/member/viewer mapped straight); workspace OWNERS also become project admins on their workspace's projects (they hold no member row today but have access through GetWorkspaceOwner); users covered by the hub bypass are skipped. Idempotent INSERT IGNORE against (project_id, user_id).
4. Existing projects stamped visibility='private' (P10).
5. Backfill re-runs immediately before Phase 4 removes the workspace fallback, closing the drift window for projects created mid-migration.
6. Dead `flow_businesses` table dropped in a standalone cleanup (unrelated finding, C6).

## Part 7: phased execution (supersedes baseline phasing)

Phase 0 (data truth, standalone small PRs):
C2 cache-eviction fix (EnsureMyTasksProject passes userID to invalidateWorkspacesScope; must be businessID); repair script executed with human review; extended dry-run counts.

Phase 1 (contracts + backend):
voxie v2 directories per Part 5; onion v2 handlers; v1 handlers re-implemented over business-primary data; schema columns and backfills 3-5; assignment auto-membership (T1); visibility resolver with curated ListProjects semantics (P6/P8); W1 policy resolver reading business config; bin backend (archive stamps purge_at; worker job purges expired bins with full cascade including R2 attachment files, P4 scope); business:<id> realtime topic plus authorizer bypass (X1); tool_executor.go re-pointed (X2). Regression tests prove v1 responses byte-identical. Rollback: revert; backfill rows are additive and harmless.

Phase 2 (frontend data layer):
All hooks swap to v2; useCurrentProject switches to GetProject (kills sentinel defaults and cold-link fragility); quick-add simplifies to title-only per M1; TeamView/TimeReports switch to business roster hooks with conditional Dash-team CTA (shown only to users who manage the team); useCurrentWorkspaceRole fallback deleted; Caret context swap client side. No visible screen changes; tests asserting plumbing may change, visible-behavior tests must not.

Phase 3 (frontend surface removal + new UI):
Delete switcher, workspace-manager, workspace-dashboard, move dialog, urlUtils, AppRoutes legacy file, workspace locale keys; routes re-path to /p/:projectId with redirects (X5); retire g d; ship All projects page, profile route, recycle bin UIs, visibility toggles, Flow Settings panels; drag rework to Pointer Events with mandatory non-drag alternative (B2); e2e specs rewritten. Split into reviewable sub-PRs: routes+deletions, new surfaces, drag rework. Visual evidence mandatory in every sub-PR body; ui-ux-pro-max checklist part of definition of done.

**Amended 2026-08-23.** "Screenshots/GIFs mandatory" was unsatisfiable for an agent-authored PR: the environment can render a page but cannot write the browser framebuffer to a file, so there is nothing to attach. Every UI sub-PR in this initiative therefore violated the definition of done on a technicality while being better verified than the clause required.

The requirement is now **visual evidence OR a written verification log**, where a log means: the specific states exercised, the assertions made, and a mutation probe showing the test fails when the behaviour is removed. A screenshot proves a screen rendered once; a mutation probe proves the behaviour is actually pinned. Prefer images where they can be produced (a human reviewer, or a harness with file output), and prefer a log over an empty section.

This is a relaxation of the artefact, not of the standard. A UI sub-PR with neither is still incomplete.

Phase 4 (backend cleanup):
Gate: full release cycle of confirmed-zero v1 traffic (compressible given empty prod, owner's call). Same maintenance window: re-run idempotent backfill, remove workspace fallback from RequireProjectRole/RequireTaskRole/projectRoleFor/IsProjectMember/IsTaskMember, delete RequireWorkspaceRole/workspaceRoleFor, re-home or repair project_id=0 tasks, drop flow_workspaces, flow_workspace_members, and every workspace_id column (projects, tasks, labels, cycles, views, columns, milestones, notifications), update aiquery allowed-column spec (C7). Fresh verified database backup before any irreversible step.

Fast-follows (sequenced after Phase 2, independent branches):
F-A Custom statuses: flow_statuses table (per-project sets with semantic buckets backlog/active/done/cancelled for cross-pipeline rollups), additive status_id alongside the synced string status, config CRUD + board integration; no v3 ever needed.
F-B HR reporting line: manager field on staff records feeding M3's resolver as an additional rule source.

## Part 8: known traps carried from verification

- C1: business-scoped reads partially exist (workspaces list keyed by business; writes guarded AND business_id; flow_insights module fully business-keyed; tenant indexes shipped). Remaining flatten work is organizational partitioning, not tenancy groundwork.
- C2: third bug in the family (cache eviction passing userID) fixed in Phase 0.
- C3/D3: CreateProject has no backend gate today; stays permissive, documented in v2 handler comments.
- C4/X1: realtime authorizer lacks the hub bypass; closes in Phase 1.
- C7: aiquery exposes flow_tasks.workspace_id; breaks at column drop unless updated in release.
- C9: no get-project-by-id exists; frontend derives from list; v2 adds it.
- MoveTaskToColumn SQL embeds workspace-membership EXISTS clauses; rewritten in Phase 1 alongside service re-pointing, gone by Phase 4.
- No Flow emails exist today (weekly_digest stored but never dispatched); no email surface to migrate.
- Project member management UI already exists (ProjectSettings + AddProjectMemberDialog); private sharing has a home from day one.
