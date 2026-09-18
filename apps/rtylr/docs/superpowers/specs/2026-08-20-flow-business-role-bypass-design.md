# Flow permissions: business-wide roles bypass project/workspace membership

Date: 2026-08-20
Status: **approved, implementation in progress**
Part of: workspace-removal initiative, sub-project 1 of 4 (see "Sequencing" below)

## The bug this fixes

A business-wide "Owner" account was locked out of editing tasks in every
project, everywhere, in their own business. Root cause: Flow's authorization
(`internal/modules/workspaces/internal/application/service.go`) only ever
resolves a caller's role through three Flow-internal paths - an explicit
`flow_project_members` row, project ownership, or `flow_workspace_members`
membership (`projectRoleFor`, `workspaceRoleFor`). None of those three paths
has any notion of the business-wide role a user holds in `@voxire/auth`/hub
(owner, admin, operator, staff, viewer). A business owner who was never
explicitly added to a specific project's member list, never created that
project themselves, and never has a workspace-membership row for that
project's workspace, has _no_ path to access - by design, today, correctly
per the existing code, and incorrectly per what a "business owner" should
mean.

This was already fixed halfway tonight on the frontend only (`currentUser.ts`
gained a workspace-role fallback), which helps when the account genuinely
holds a workspace role, but does nothing when it holds none - which is
exactly what "every project, everywhere" indicates for this account.

## Sequencing

This is sub-project 1 of a 4-part initiative to remove the Workspace concept
from Flow entirely (flattening Workspace > Project > Task to Project > Task,
projects living directly on the business). The other three sub-projects -
contract migration, backend schema migration, and frontend removal - are
separately specced later. This sub-project stands on its own: it fixes the
actual authorization bug immediately, and it is a prerequisite for the later
ones, because once a business-wide role bypass exists, removing the
workspace-membership fallback in sub-project 3 cannot silently revoke access
the way it would today.

## Design

### Backend (`onion`)

`internal/platform/hubprofile` already defines a `Resolver` interface -
`GetProfile(ctx, userID, businessID) -> Profile{Role, IsOwner, ...}` - built
for exactly this cross-module need, and already implemented by the `team`
module (`hub_profile_resolver.go`) for at least one other consumer. No new
abstraction: `workspaces.Service` gets this same resolver injected.

`businessID` is already resolved at every call site in
`internal/modules/workspaces/internal/api/task_service.go` via
`resolveAuth` -> `tenant.ResolveBusinessID(ctx, authUser.BusinessID,
authUser.UserID)`, so no new plumbing is needed to get it to the
authorization layer.

The bypass is checked in the three `Require*` gate functions themselves -
`RequireProjectRole`, `RequireTaskRole`, `RequireWorkspaceRole` - not inside
`projectRoleFor`/`workspaceRoleFor`. Those two functions resolve a role
_within Flow's own project/workspace system_; a business-wide bypass is not
a Flow project role, it is a separate authorization concern layered on top.
Keeping it out of `projectRoleFor` keeps that function's contract exactly as
legible as its own doc comment already describes it.

```go
// In each Require*Role function, before falling through to the existing chain:
if profile, err := s.hubProfiles.GetProfile(ctx, userID, businessID); err == nil {
    if profile.IsOwner || profile.Role == "admin" || profile.Role == "operator" {
        return nil // domain.FlowRoleAdmin-equivalent: full access, no further checks
    }
}
// Resolver error falls through silently to the existing chain - a hub-service
// hiccup can only fail to grant this shortcut, never grant access it shouldn't.
```

Owner, admin, and operator all bypass, at full admin-level strength (edit,
delete, archive, manage members) - matching how project/workspace ownership
already maps to `domain.FlowRoleAdmin` elsewhere in this same file.

### Frontend (`apps/flow`)

Fixing only the backend leaves the bug user-visible: the frontend would still
show the read-only banner for a business owner/admin/operator, because
`useCanEditTask`/`useCanDeleteResource` (`shared/lib/currentUser.ts`) never
read business-wide role at all. Flow's own `FlowUser` type has no role field;
the role lives in `@voxire/auth`'s `useUserStore`, which Flow currently only
touches for sign-out (`FlowLayout.tsx`).

Both hooks read `@voxire/auth`'s user role first. If it's owner, admin, or
operator, return `true` immediately - before consulting project or workspace
role at all, mirroring the backend's own precedence exactly. Tonight's
workspace-role fallback stays underneath, unchanged: workspaces are not gone
yet (that is sub-projects 2-4), so a workspace member who is not a
business-wide owner/admin/operator still needs that path.

### Error handling

- Backend: resolver error -> fall through, do not grant. A user with a real
  project/workspace role is unaffected either way.
- Frontend: if `@voxire/auth`'s user/role hasn't loaded yet (undefined),
  treat as "no bypass" and fall through to the existing project/workspace
  checks - fail closed, consistent with every other check in this file.

### Testing

- Backend: unit tests on `RequireProjectRole`, `RequireTaskRole`,
  `RequireWorkspaceRole` with a mocked `hubprofile.Resolver` returning each of
  owner/admin/operator/staff/viewer, plus a resolver-error case asserting
  fallthrough rather than a granted or hard-failed request.
- Frontend: extend `currentUser.test.ts`'s `useCanEditTask` suite (added
  tonight) with an owner/admin/operator business-role case that bypasses with
  _no_ project role and _no_ workspace role at all, and a staff/viewer case
  that does not bypass and still falls through to the existing checks.

### Out of scope for this sub-project

- Removing `flow_workspace_members`, `workspace_id`, or any workspace UI -
  that is sub-projects 2-4.
- Making the bypass roles configurable per business - raised earlier tonight
  as a real want, but it is a distinct feature (a company-level setting with
  its own contract) layered on top of this fix, not part of it.
- The `business_id DEFAULT 0` tenant-isolation risk flagged during
  investigation - real, but belongs to sub-project 3 (schema migration),
  since it is about column defaults, not this authorization path.
