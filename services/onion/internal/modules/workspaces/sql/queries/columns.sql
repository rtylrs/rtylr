-- name: GetColumnsByProject :many
SELECT id, project_id, workspace_id, name, color, sort_order, created_by, created_at, updated_at
FROM flow_project_columns
WHERE project_id = ?
ORDER BY sort_order ASC
LIMIT ? OFFSET ?;

-- name: CreateColumn :execresult
INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateColumn :execresult
UPDATE flow_project_columns
SET name = ?, color = ?, sort_order = ?, updated_at = NOW()
WHERE flow_project_columns.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)
  );

-- name: ClearTasksSectionID :exec
UPDATE flow_tasks
SET section_id = NULL
WHERE section_id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );

-- name: DeleteColumn :execresult
DELETE FROM flow_project_columns
WHERE flow_project_columns.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)
  );

-- name: MoveTaskToColumn :execresult
UPDATE flow_tasks
SET section_id = sqlc.narg('section_id'), updated_at = NOW()
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  )
  AND (
    sqlc.arg('column_id') = 0
    OR EXISTS (
      SELECT 1 FROM flow_project_columns c
      WHERE c.id = sqlc.arg('column_id') AND c.workspace_id = flow_tasks.workspace_id
    )
  );

-- name: SeedDefaultColumn :execresult
INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
VALUES (?, ?, ?, ?, ?, ?);
