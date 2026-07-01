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
UPDATE flow_tasks
SET labels = ?
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );
