-- name: UpdateLabel :execresult
UPDATE flow_labels
SET name = ?, color = ?
WHERE flow_labels.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_labels.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_labels.workspace_id AND wm.user_id = ?)
  );

-- name: DeleteLabel :execresult
DELETE FROM flow_labels
WHERE flow_labels.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_labels.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_labels.workspace_id AND wm.user_id = ?)
  );

-- name: SetTaskLabels :execresult
-- Authorization happens in label_service.go, which calls RequireTaskRole with
-- FlowRoleMember before this runs. RequireTaskRole resolves a v2 task through
-- its project, so it is the check that actually works.
--
-- This statement used to re-check workspace membership on top of that. A v2 task
-- has workspace_id = 0, so neither EXISTS could ever be true, and requireAffected
-- turned the zero-row UPDATE into ErrUnauthorized: the caller passed
-- authorization and was then denied by the write. Setting labels on a task in
-- any project created since project creation moved to v2 failed outright.
--
-- The project branches below are additive so a legacy v1 task, which still has a
-- real workspace_id until Phase 4, keeps resolving exactly as before.
UPDATE flow_tasks
SET labels = ?
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_tasks.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );
