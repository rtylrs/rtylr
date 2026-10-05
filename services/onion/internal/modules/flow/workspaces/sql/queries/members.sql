-- name: CountWorkspaceMembers :one
SELECT COUNT(*) FROM flow_workspace_members WHERE workspace_id = ?;

-- name: ListWorkspaceMembers :many
SELECT
    wm.id, wm.workspace_id, wm.user_id, wm.role, wm.joined_at,
    w.business_id
FROM flow_workspace_members wm
JOIN flow_workspaces w ON w.id = wm.workspace_id
WHERE wm.workspace_id = ?
ORDER BY wm.joined_at ASC
LIMIT ? OFFSET ?;

-- name: UpdateWorkspaceMemberRole :exec
UPDATE flow_workspace_members
SET role = ?
WHERE workspace_id = ? AND user_id = ?;

-- The tenant filter is load-bearing, not defensive. flow_workspaces is
-- business-scoped and GetWorkspacesByOwner reads it by business_id, so a lookup
-- on owner_id alone can hand back another business's personal workspace and
-- CreatePersonalWorkspace will treat it as already done. #703 fixed the INSERT
-- to write a real business_id but left this lookup unscoped, which is why
-- TestWorkspaces_CreatePersonalWorkspace_Idempotent fails intermittently.
-- name: GetPersonalWorkspaceByOwner :one
SELECT id
FROM flow_workspaces
WHERE owner_id = ? AND business_id = ? AND (is_personal = TRUE OR is_default = TRUE)
ORDER BY is_personal DESC, is_default DESC, id ASC
LIMIT 1;

-- name: CreatePersonalWorkspace :execresult
INSERT INTO flow_workspaces (owner_id, business_id, name, icon, color, is_default, is_personal, created_by)
VALUES (?, ?, 'Personal', '🏠', '#64C5B9', TRUE, TRUE, ?);

-- name: CountProjectMembers :one
SELECT COUNT(*) FROM flow_project_members WHERE project_id = ?;

-- name: ListProjectMembers :many
SELECT
    pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
    w.business_id
FROM flow_project_members pm
JOIN flow_projects w ON w.id = pm.project_id
WHERE pm.project_id = ?
ORDER BY pm.joined_at ASC
LIMIT ? OFFSET ?;

-- name: AddProjectMember :exec
INSERT INTO flow_project_members (project_id, user_id, role)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE role = VALUES(role);

-- name: RemoveProjectMember :exec
DELETE FROM flow_project_members
WHERE project_id = ? AND user_id = ?;

-- name: UpdateProjectMemberRole :exec
UPDATE flow_project_members
SET role = ?
WHERE project_id = ? AND user_id = ?;

-- name: TransferProjectOwnership :exec
UPDATE flow_projects
SET owner_id = ?
WHERE id = ? AND owner_id = ?;

-- name: MoveProject :exec
UPDATE flow_projects
SET workspace_id = ?
WHERE id = ? AND owner_id = ?;

-- name: GetUserWorkspaceRole :one
SELECT role
FROM flow_workspace_members
WHERE workspace_id = ? AND user_id = ?
LIMIT 1;

-- name: GetProjectMemberRole :one
SELECT role
FROM flow_project_members
WHERE project_id = ? AND user_id = ?
LIMIT 1;

-- name: GetProjectOwnerWorkspace :one
SELECT owner_id, workspace_id
FROM flow_projects
WHERE id = ?
LIMIT 1;

-- name: GetWorkspaceOwner :one
SELECT owner_id
FROM flow_workspaces
WHERE id = ?
LIMIT 1;
